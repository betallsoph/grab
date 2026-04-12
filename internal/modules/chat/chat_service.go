package chat

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateConversation tạo conversation direct hoặc nhóm.
// Nếu direct giữa 2 user đã tồn tại → trả lại cái cũ.
func (s *Service) CreateConversation(ctx context.Context, userID uint, req CreateConversationRequest) (*Conversation, error) {
	if len(req.Participants) == 0 {
		return nil, errors.New("participants is required")
	}

	// Đảm bảo người tạo nằm trong participants
	found := false
	for _, p := range req.Participants {
		if p == userID {
			found = true
			break
		}
	}
	if !found {
		req.Participants = append(req.Participants, userID)
	}

	switch req.Type {
	case ConvDirect:
		if len(req.Participants) != 2 {
			return nil, errors.New("direct conversation requires exactly 2 participants")
		}
		if req.Participants[0] == req.Participants[1] {
			return nil, errors.New("direct conversation requires two different users")
		}
		// Kiểm tra đã tồn tại chưa
		existing, err := s.repo.FindDirectConversation(ctx, req.Participants[0], req.Participants[1])
		if err == nil {
			return existing, nil
		}
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("service.CreateConversation find direct: %w", err)
		}

	case ConvGroup:
		if req.Name == "" {
			return nil, errors.New("group name is required")
		}
		if len(req.Participants) < 3 {
			return nil, errors.New("group requires at least 3 participants")
		}

	default:
		return nil, fmt.Errorf("invalid conversation type: %s", req.Type)
	}

	conv := &Conversation{
		Type:         req.Type,
		Name:         req.Name,
		Participants: req.Participants,
		CreatedBy:    userID,
	}
	if err := s.repo.CreateConversation(ctx, conv); err != nil {
		return nil, err
	}
	return conv, nil
}

// ListConversations trả về danh sách conversation của user.
func (s *Service) ListConversations(ctx context.Context, userID uint, limit, offset int) ([]Conversation, error) {
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.ListConversations(ctx, userID, limit, offset)
}

// SendMessage lưu tin nhắn vào MongoDB.
func (s *Service) SendMessage(ctx context.Context, userID uint, convIDHex string, req SendMessageRequest) (*Message, error) {
	convID, err := bson.ObjectIDFromHex(convIDHex)
	if err != nil {
		return nil, errors.New("invalid conversation_id")
	}

	if req.Content == "" {
		return nil, errors.New("content is required")
	}
	if req.ContentType == "" {
		req.ContentType = "text"
	}

	// Xác minh user thuộc conversation
	conv, err := s.repo.FindConversationByID(ctx, convID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}
	if !containsUser(conv.Participants, userID) {
		return nil, errors.New("you are not a participant of this conversation")
	}

	msg := &Message{
		ConversationID: convID,
		SenderID:       userID,
		Content:        req.Content,
		ContentType:    req.ContentType,
	}
	if err := s.repo.InsertMessage(ctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// GetMessages lấy lịch sử tin nhắn trong conversation.
func (s *Service) GetMessages(ctx context.Context, userID uint, convIDHex string, limit, offset int) ([]Message, error) {
	convID, err := bson.ObjectIDFromHex(convIDHex)
	if err != nil {
		return nil, errors.New("invalid conversation_id")
	}

	conv, err := s.repo.FindConversationByID(ctx, convID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}
	if !containsUser(conv.Participants, userID) {
		return nil, errors.New("you are not a participant of this conversation")
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return s.repo.ListMessages(ctx, convID, limit, offset)
}

// GetConversation trả về thông tin conversation (kiểm tra quyền).
func (s *Service) GetConversation(ctx context.Context, userID uint, convIDHex string) (*Conversation, error) {
	convID, err := bson.ObjectIDFromHex(convIDHex)
	if err != nil {
		return nil, errors.New("invalid conversation_id")
	}
	conv, err := s.repo.FindConversationByID(ctx, convID)
	if err != nil {
		return nil, err
	}
	if !containsUser(conv.Participants, userID) {
		return nil, errors.New("you are not a participant of this conversation")
	}
	return conv, nil
}

func containsUser(participants []uint, userID uint) bool {
	for _, p := range participants {
		if p == userID {
			return true
		}
	}
	return false
}
