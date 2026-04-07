package reminder

import (
	"net/http"

	"github.com/redis/go-redis/v9"
	"grab/internal/modules/user"
	"gorm.io/gorm"
)

// RegisterRoutes khởi tạo module Reminder và đăng ký route.
//
//	GET /api/v1/reminder/config   → lấy cấu hình
//	PUT /api/v1/reminder/config   → cập nhật cấu hình
//	GET /api/v1/reminder/check    → kiểm tra trạng thái (mobile polling)
func RegisterRoutes(mux *http.ServeMux, db *gorm.DB, rdb *redis.Client, userHandler *user.Handler) {
	svc := NewService(db, rdb)
	h := NewHandler(svc)

	auth := userHandler.AuthMiddleware

	mux.Handle("GET /api/v1/reminder/config", auth(http.HandlerFunc(h.GetConfig)))
	mux.Handle("PUT /api/v1/reminder/config", auth(http.HandlerFunc(h.UpdateConfig)))
	mux.Handle("GET /api/v1/reminder/check", auth(http.HandlerFunc(h.CheckStatus)))
}
