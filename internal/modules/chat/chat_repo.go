package chat

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct {
	conversations *mongo.Collection
	messages      *mongo.Collection
}

func NewRepository(client *mongo.Client) *Repository {
	db := client.Database("grab_chat")
	return &Repository{
		conversations: db.Collection("conversations"),
		messages:      db.Collection("messages"),
	}
}

// EnsureIndexes tạo index cho performance.
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := r.messages.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "conversation_id", Value: 1},
			{Key: "created_at", Value: -1},
		},
	})
	if err != nil {
		return fmt.Errorf("chat.EnsureIndexes messages: %w", err)
	}

	_, err = r.conversations.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "participants", Value: 1}},
	})
	if err != nil {
		return fmt.Errorf("chat.EnsureIndexes conversations: %w", err)
	}
	return nil
}

// CreateConversation tạo conversation mới.
func (r *Repository) CreateConversation(ctx context.Context, conv *Conversation) error {
	conv.CreatedAt = time.Now()
	conv.UpdatedAt = conv.CreatedAt
	result, err := r.conversations.InsertOne(ctx, conv)
	if err != nil {
		return fmt.Errorf("repo.CreateConversation: %w", err)
	}
	if oid, ok := result.InsertedID.(bson.ObjectID); ok {
		conv.ID = oid
	}
	return nil
}

// FindConversationByID tìm conversation theo ID.
func (r *Repository) FindConversationByID(ctx context.Context, id bson.ObjectID) (*Conversation, error) {
	var conv Conversation
	err := r.conversations.FindOne(ctx, bson.M{"_id": id}).Decode(&conv)
	if err != nil {
		return nil, fmt.Errorf("repo.FindConversationByID: %w", err)
	}
	return &conv, nil
}

// FindDirectConversation tìm conversation direct giữa 2 user (nếu đã tồn tại).
func (r *Repository) FindDirectConversation(ctx context.Context, userA, userB uint) (*Conversation, error) {
	var conv Conversation
	err := r.conversations.FindOne(ctx, bson.M{
		"type":         ConvDirect,
		"participants": bson.M{"$all": bson.A{userA, userB}},
	}).Decode(&conv)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// ListConversations lấy danh sách conversation mà user tham gia,
// sắp xếp mới nhất trước.
func (r *Repository) ListConversations(ctx context.Context, userID uint, limit, offset int) ([]Conversation, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "updated_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.conversations.Find(ctx, bson.M{
		"participants": userID,
	}, opts)
	if err != nil {
		return nil, fmt.Errorf("repo.ListConversations: %w", err)
	}
	defer cursor.Close(ctx)

	var convs []Conversation
	if err := cursor.All(ctx, &convs); err != nil {
		return nil, fmt.Errorf("repo.ListConversations decode: %w", err)
	}
	return convs, nil
}

// InsertMessage lưu tin nhắn mới và cập nhật updated_at của conversation.
func (r *Repository) InsertMessage(ctx context.Context, msg *Message) error {
	msg.CreatedAt = time.Now()
	result, err := r.messages.InsertOne(ctx, msg)
	if err != nil {
		return fmt.Errorf("repo.InsertMessage: %w", err)
	}
	if oid, ok := result.InsertedID.(bson.ObjectID); ok {
		msg.ID = oid
	}

	// Cập nhật updated_at để sort conversation list
	_, _ = r.conversations.UpdateOne(ctx,
		bson.M{"_id": msg.ConversationID},
		bson.M{"$set": bson.M{"updated_at": msg.CreatedAt}},
	)
	return nil
}

// ListMessages lấy tin nhắn trong conversation, mới nhất trước.
func (r *Repository) ListMessages(ctx context.Context, convID bson.ObjectID, limit, offset int) ([]Message, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.messages.Find(ctx, bson.M{
		"conversation_id": convID,
	}, opts)
	if err != nil {
		return nil, fmt.Errorf("repo.ListMessages: %w", err)
	}
	defer cursor.Close(ctx)

	var msgs []Message
	if err := cursor.All(ctx, &msgs); err != nil {
		return nil, fmt.Errorf("repo.ListMessages decode: %w", err)
	}
	return msgs, nil
}
