package user

import (
	"net/http"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// RegisterRoutes khởi tạo toàn bộ dependencies của module User
// và đăng ký các route lên mux.
//
// Các route public (không cần JWT):
//
//	POST /api/v1/auth/register
//	POST /api/v1/auth/login
//
// Các route protected (cần JWT):
//
//	POST /api/v1/drivers/location
func RegisterRoutes(mux *http.ServeMux, db *gorm.DB, rdb *redis.Client, jwtSecret string) *Handler {
	repo := NewRepository(db)
	svc := NewService(repo, rdb, jwtSecret)
	h := NewHandler(svc)

	// Public routes
	mux.HandleFunc("POST /api/v1/auth/register", h.Register)
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)

	// Protected route — bọc bằng AuthMiddleware
	mux.Handle("POST /api/v1/drivers/location", h.AuthMiddleware(http.HandlerFunc(h.UpdateLocation)))

	// Trả về Handler để các module khác (SOS, Chat...) dùng AuthMiddleware
	return h
}
