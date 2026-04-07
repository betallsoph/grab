package earning

import (
	"net/http"

	"grab/internal/modules/user"
	"gorm.io/gorm"
)

// RegisterRoutes khởi tạo module Earning và đăng ký route.
//
//	POST /api/v1/earning/trips             → ghi nhận cuốc xe
//	GET  /api/v1/earning/trips             → danh sách cuốc xe
//	GET  /api/v1/earning/summary/daily     → tổng kết ngày
//	GET  /api/v1/earning/summary/monthly   → tổng kết tháng
func RegisterRoutes(mux *http.ServeMux, db *gorm.DB, userHandler *user.Handler) {
	repo := NewRepository(db)
	svc := NewService(repo)
	h := NewHandler(svc)

	auth := userHandler.AuthMiddleware

	mux.Handle("POST /api/v1/earning/trips", auth(http.HandlerFunc(h.AddTrip)))
	mux.Handle("GET /api/v1/earning/trips", auth(http.HandlerFunc(h.ListTrips)))
	mux.Handle("GET /api/v1/earning/summary/daily", auth(http.HandlerFunc(h.DailySummary)))
	mux.Handle("GET /api/v1/earning/summary/monthly", auth(http.HandlerFunc(h.MonthlySummary)))
}
