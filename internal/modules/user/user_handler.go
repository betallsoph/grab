package user

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"grab/internal/core/httputil"
)

// Handler nhận HTTP request, parse dữ liệu rồi ủy quyền cho Service xử lý.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// --- Handlers ---

// Register godoc
// @Summary      Đăng ký tài khoản tài xế
// @Description  Tạo tài khoản mới bằng số điện thoại và mật khẩu
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      RegisterRequest  true  "Thông tin đăng ký"
// @Success      201   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Router       /auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	httputil.LimitBody(r)

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.Register(r.Context(), req); err != nil {
		msg := err.Error()
		if strings.HasPrefix(msg, "phone and password are required") {
			httputil.WriteError(w, http.StatusBadRequest, msg)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, map[string]string{"message": "registered successfully"})
}

// Login godoc
// @Summary      Đăng nhập
// @Description  Xác thực bằng SĐT + mật khẩu, trả về JWT token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      LoginRequest   true  "Thông tin đăng nhập"
// @Success      200   {object}  LoginResponse
// @Failure      401   {object}  map[string]string
// @Router       /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	httputil.LimitBody(r)

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		if err.Error() == "phone or password is incorrect" {
			httputil.WriteError(w, http.StatusUnauthorized, err.Error())
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}

// UpdateLocation godoc
// @Summary      Cập nhật vị trí GPS
// @Description  Mobile bắn tọa độ GPS lên mỗi vài giây, lưu vào Redis Geo
// @Tags         Driver
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      UpdateLocationRequest  true  "Tọa độ GPS"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Router       /drivers/location [post]
func (h *Handler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	httputil.LimitBody(r)

	userID, ok := r.Context().Value(ContextKeyUserID).(uint)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req UpdateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if !httputil.ValidCoords(req.Latitude, req.Longitude) {
		httputil.WriteError(w, http.StatusBadRequest, "invalid coordinates")
		return
	}

	if err := h.service.UpdateLocation(r.Context(), userID, req); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update location")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "location updated"})
}

// --- Middleware ---

// AuthMiddleware validate JWT trong header Authorization: Bearer <token>.
// Nếu hợp lệ, inject userID vào context rồi gọi handler tiếp theo.
// Các module khác (SOS, Chat...) dùng middleware này bằng cách import package user.
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			httputil.WriteError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		// Format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			httputil.WriteError(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		claims, err := h.service.ParseJWT(parts[1])
		if err != nil {
			httputil.WriteError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		// Nhúng userID vào context để các handler downstream đọc ra
		ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
