package chat

import (
	"net/http"

	"grab/internal/modules/user"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// RegisterRoutes khởi tạo module Chat và đăng ký route.
//
// REST APIs (tất cả cần JWT):
//
//	POST /api/v1/chat/conversations                       → tạo conversation
//	GET  /api/v1/chat/conversations                       → danh sách conversations
//	POST /api/v1/chat/conversations/{convId}/messages      → gửi tin nhắn (REST fallback)
//	GET  /api/v1/chat/conversations/{convId}/messages      → lịch sử tin nhắn
//
// WebSocket (cần JWT):
//
//	GET  /api/v1/chat/ws                                   → real-time chat
func RegisterRoutes(mux *http.ServeMux, mongoClient *mongo.Client, userHandler *user.Handler) {
	repo := NewRepository(mongoClient)
	svc := NewService(repo)
	hub := NewHub(svc)
	h := NewHandler(svc, hub)

	go hub.Run()

	auth := userHandler.AuthMiddleware

	mux.Handle("POST /api/v1/chat/conversations", auth(http.HandlerFunc(h.CreateConversation)))
	mux.Handle("GET /api/v1/chat/conversations", auth(http.HandlerFunc(h.ListConversations)))
	mux.Handle("POST /api/v1/chat/conversations/{convId}/messages", auth(http.HandlerFunc(h.SendMessage)))
	mux.Handle("GET /api/v1/chat/conversations/{convId}/messages", auth(http.HandlerFunc(h.GetMessages)))
	mux.Handle("GET /api/v1/chat/ws", auth(http.HandlerFunc(h.ServeWS)))
}
