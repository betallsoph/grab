package mood

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct {
	posts    *mongo.Collection
	comments *mongo.Collection
}

func NewRepository(client *mongo.Client) *Repository {
	db := client.Database("grab_mood")
	return &Repository{
		posts:    db.Collection("mood_posts"),
		comments: db.Collection("mood_comments"),
	}
}

func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := r.posts.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "created_at", Value: -1}},
	})
	if err != nil {
		return fmt.Errorf("mood.EnsureIndexes posts: %w", err)
	}
	_, err = r.comments.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "post_id", Value: 1},
			{Key: "created_at", Value: 1},
		},
	})
	if err != nil {
		return fmt.Errorf("mood.EnsureIndexes comments: %w", err)
	}
	return nil
}

// --- Posts ---

func (r *Repository) CreatePost(ctx context.Context, p *Post) error {
	p.CreatedAt = time.Now()
	result, err := r.posts.InsertOne(ctx, p)
	if err != nil {
		return fmt.Errorf("repo.CreatePost: %w", err)
	}
	p.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

func (r *Repository) ListPosts(ctx context.Context, limit, offset int) ([]Post, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.posts.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("repo.ListPosts: %w", err)
	}
	defer cursor.Close(ctx)

	var posts []Post
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, fmt.Errorf("repo.ListPosts decode: %w", err)
	}
	return posts, nil
}

func (r *Repository) FindPostByID(ctx context.Context, id bson.ObjectID) (*Post, error) {
	var p Post
	if err := r.posts.FindOne(ctx, bson.M{"_id": id}).Decode(&p); err != nil {
		return nil, fmt.Errorf("repo.FindPostByID: %w", err)
	}
	return &p, nil
}

func (r *Repository) LikePost(ctx context.Context, id bson.ObjectID) error {
	_, err := r.posts.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$inc": bson.M{"like_count": 1}},
	)
	if err != nil {
		return fmt.Errorf("repo.LikePost: %w", err)
	}
	return nil
}

// --- Comments ---

func (r *Repository) CreateComment(ctx context.Context, c *Comment) error {
	c.CreatedAt = time.Now()
	result, err := r.comments.InsertOne(ctx, c)
	if err != nil {
		return fmt.Errorf("repo.CreateComment: %w", err)
	}
	c.ID = result.InsertedID.(bson.ObjectID)

	// Tăng comment_count trên post
	_, _ = r.posts.UpdateOne(ctx,
		bson.M{"_id": c.PostID},
		bson.M{"$inc": bson.M{"comment_count": 1}},
	)
	return nil
}

func (r *Repository) ListComments(ctx context.Context, postID bson.ObjectID, limit, offset int) ([]Comment, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: 1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.comments.Find(ctx, bson.M{"post_id": postID}, opts)
	if err != nil {
		return nil, fmt.Errorf("repo.ListComments: %w", err)
	}
	defer cursor.Close(ctx)

	var comments []Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, fmt.Errorf("repo.ListComments decode: %w", err)
	}
	return comments, nil
}
