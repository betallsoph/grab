package mood

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Post lưu trong MongoDB collection "mood_posts".
type Post struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	AuthorID     uint          `bson:"author_id" json:"author_id"`
	AuthorName   string        `bson:"author_name" json:"author_name"` // hiển thị "Ẩn danh" nếu anonymous
	IsAnonymous  bool          `bson:"is_anonymous" json:"is_anonymous"`
	Content      string        `bson:"content" json:"content"`
	Mood         string        `bson:"mood" json:"mood"` // happy, sad, angry, tired, funny
	LikeCount    int           `bson:"like_count" json:"like_count"`
	CommentCount int           `bson:"comment_count" json:"comment_count"`
	CreatedAt    time.Time     `bson:"created_at" json:"created_at"`
}

// Comment lưu trong MongoDB collection "mood_comments".
type Comment struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	PostID      bson.ObjectID `bson:"post_id" json:"post_id"`
	AuthorID    uint          `bson:"author_id" json:"author_id"`
	AuthorName  string        `bson:"author_name" json:"author_name"`
	IsAnonymous bool          `bson:"is_anonymous" json:"is_anonymous"`
	Content     string        `bson:"content" json:"content"`
	CreatedAt   time.Time     `bson:"created_at" json:"created_at"`
}

// --- Request DTOs ---

type CreatePostRequest struct {
	Content     string `json:"content"`
	Mood        string `json:"mood"`
	IsAnonymous bool   `json:"is_anonymous"`
}

type CreateCommentRequest struct {
	Content     string `json:"content"`
	IsAnonymous bool   `json:"is_anonymous"`
}

var validMoods = map[string]bool{
	"happy": true, "sad": true, "angry": true,
	"tired": true, "funny": true, "grateful": true,
}

func IsValidMood(m string) bool {
	return validMoods[m]
}
