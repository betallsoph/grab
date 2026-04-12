package mood

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"gorm.io/gorm"
)

type Service struct {
	repo *Repository
	db   *gorm.DB // đọc tên tài xế từ PostgreSQL
}

func NewService(repo *Repository, db *gorm.DB) *Service {
	return &Service{repo: repo, db: db}
}

// CreatePost tạo bài viết mới trên góc xả mood.
func (s *Service) CreatePost(ctx context.Context, userID uint, req CreatePostRequest) (*Post, error) {
	if req.Content == "" {
		return nil, errors.New("content is required")
	}
	if req.Mood != "" && !IsValidMood(req.Mood) {
		return nil, fmt.Errorf("invalid mood: %s (valid: happy, sad, angry, tired, funny, grateful)", req.Mood)
	}

	authorName := s.getDriverName(ctx, userID)
	if req.IsAnonymous {
		authorName = "Tài xế ẩn danh"
	}

	post := &Post{
		AuthorID:    userID,
		AuthorName:  authorName,
		IsAnonymous: req.IsAnonymous,
		Content:     req.Content,
		Mood:        req.Mood,
	}
	if err := s.repo.CreatePost(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

// ListPosts lấy feed bài viết, mới nhất trước.
func (s *Service) ListPosts(ctx context.Context, limit, offset int) ([]Post, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.ListPosts(ctx, limit, offset)
}

// LikePost tăng like cho bài viết.
func (s *Service) LikePost(ctx context.Context, postIDHex string) error {
	id, err := bson.ObjectIDFromHex(postIDHex)
	if err != nil {
		return errors.New("invalid post id")
	}
	return s.repo.LikePost(ctx, id)
}

// CreateComment thêm comment vào bài viết.
func (s *Service) CreateComment(ctx context.Context, userID uint, postIDHex string, req CreateCommentRequest) (*Comment, error) {
	if req.Content == "" {
		return nil, errors.New("content is required")
	}

	postID, err := bson.ObjectIDFromHex(postIDHex)
	if err != nil {
		return nil, errors.New("invalid post id")
	}

	// Kiểm tra post tồn tại
	if _, err := s.repo.FindPostByID(ctx, postID); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("post not found")
		}
		return nil, fmt.Errorf("find post: %w", err)
	}

	authorName := s.getDriverName(ctx, userID)
	if req.IsAnonymous {
		authorName = "Tài xế ẩn danh"
	}

	comment := &Comment{
		PostID:      postID,
		AuthorID:    userID,
		AuthorName:  authorName,
		IsAnonymous: req.IsAnonymous,
		Content:     req.Content,
	}
	if err := s.repo.CreateComment(ctx, comment); err != nil {
		return nil, err
	}
	return comment, nil
}

// ListComments lấy danh sách comment của bài viết.
func (s *Service) ListComments(ctx context.Context, postIDHex string, limit, offset int) ([]Comment, error) {
	postID, err := bson.ObjectIDFromHex(postIDHex)
	if err != nil {
		return nil, errors.New("invalid post id")
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return s.repo.ListComments(ctx, postID, limit, offset)
}

func (s *Service) getDriverName(ctx context.Context, userID uint) string {
	var name string
	s.db.WithContext(ctx).Table("users").
		Select("full_name").
		Where("id = ?", userID).
		Scan(&name)
	if name == "" {
		name = "Tài xế"
	}
	return name
}
