package survivalmap

import (
	"net/http"

	"grab/internal/modules/user"
	"gorm.io/gorm"
)

// RegisterRoutes khởi tạo module Map Sinh Tồn và đăng ký route.
//
// Tất cả route đều cần JWT (AuthMiddleware).
//
//	POST /api/v1/map/pois              → tạo POI mới
//	GET  /api/v1/map/pois/nearby       → tìm POI xung quanh
//	GET  /api/v1/map/pois/{id}         → xem chi tiết POI
//	POST /api/v1/map/pois/{id}/upvote  → upvote POI
func RegisterRoutes(mux *http.ServeMux, db *gorm.DB, userHandler *user.Handler) {
	repo := NewRepository(db)
	svc := NewService(repo)
	h := NewHandler(svc)

	auth := userHandler.AuthMiddleware

	mux.Handle("POST /api/v1/map/pois", auth(http.HandlerFunc(h.CreatePOI)))
	mux.Handle("GET /api/v1/map/pois/nearby", auth(http.HandlerFunc(h.SearchNearby)))
	mux.Handle("GET /api/v1/map/pois/{id}", auth(http.HandlerFunc(h.GetPOI)))
	mux.Handle("POST /api/v1/map/pois/{id}/upvote", auth(http.HandlerFunc(h.UpvotePOI)))
}
