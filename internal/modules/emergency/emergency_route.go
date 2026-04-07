package emergency

import (
	"net/http"

	"grab/internal/modules/user"
	"gorm.io/gorm"
)

// RegisterRoutes khởi tạo module Emergency Contact và đăng ký route.
//
//	POST   /api/v1/emergency/contacts              → thêm liên hệ
//	GET    /api/v1/emergency/contacts              → danh sách liên hệ
//	DELETE /api/v1/emergency/contacts/{contactId}  → xóa liên hệ
func RegisterRoutes(mux *http.ServeMux, db *gorm.DB, userHandler *user.Handler) {
	svc := NewService(db)
	h := NewHandler(svc)

	auth := userHandler.AuthMiddleware

	mux.Handle("POST /api/v1/emergency/contacts", auth(http.HandlerFunc(h.AddContact)))
	mux.Handle("GET /api/v1/emergency/contacts", auth(http.HandlerFunc(h.ListContacts)))
	mux.Handle("DELETE /api/v1/emergency/contacts/{contactId}", auth(http.HandlerFunc(h.RemoveContact)))
}
