package mood

import (
	"net/http"

	"grab/internal/modules/user"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"gorm.io/gorm"
)

// RegisterRoutes khởi tạo module Mood và đăng ký route.
//
//	POST /api/v1/mood/posts                     → đăng bài xả mood
//	GET  /api/v1/mood/posts                     → feed bài viết
//	POST /api/v1/mood/posts/{postId}/like       → like bài viết
//	POST /api/v1/mood/posts/{postId}/comments   → bình luận
//	GET  /api/v1/mood/posts/{postId}/comments   → danh sách comment
func RegisterRoutes(mux *http.ServeMux, db *gorm.DB, mongoClient *mongo.Client, userHandler *user.Handler) {
	repo := NewRepository(mongoClient)
	svc := NewService(repo, db)
	h := NewHandler(svc)

	auth := userHandler.AuthMiddleware

	mux.Handle("POST /api/v1/mood/posts", auth(http.HandlerFunc(h.CreatePost)))
	mux.Handle("GET /api/v1/mood/posts", auth(http.HandlerFunc(h.ListPosts)))
	mux.Handle("POST /api/v1/mood/posts/{postId}/like", auth(http.HandlerFunc(h.LikePost)))
	mux.Handle("POST /api/v1/mood/posts/{postId}/comments", auth(http.HandlerFunc(h.CreateComment)))
	mux.Handle("GET /api/v1/mood/posts/{postId}/comments", auth(http.HandlerFunc(h.ListComments)))
}
