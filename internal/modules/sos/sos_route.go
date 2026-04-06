package sos

import (
	"net/http"

	"github.com/redis/go-redis/v9"
	"grab/internal/modules/user"
	"gorm.io/gorm"
)

// RegisterRoutes khởi tạo module SOS và đăng ký route.
//
// Route (protected — cần JWT):
//
//	GET /api/v1/sos/ws  →  WebSocket upgrade
//
// userHandler được truyền từ bên ngoài để dùng chung AuthMiddleware.
func RegisterRoutes(mux *http.ServeMux, db *gorm.DB, rdb *redis.Client, userHandler *user.Handler) {
	svc := NewService(rdb, db)
	hub := NewHub(svc)
	h := NewHandler(hub)

	// Chạy Hub event loop trong goroutine riêng
	go hub.Run()

	mux.Handle("GET /api/v1/sos/ws",
		userHandler.AuthMiddleware(http.HandlerFunc(h.ServeWS)),
	)
}
