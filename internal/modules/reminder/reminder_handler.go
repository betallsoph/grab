package reminder

import (
	"encoding/json"
	"net/http"

	"grab/internal/core/httputil"
	"grab/internal/modules/user"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetConfig godoc
// @Summary      Lấy cấu hình nhắc nhở
// @Description  Lấy cấu hình nhắc uống nước, vận động. Tạo mặc định nếu chưa có.
// @Tags         Reminder
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  ReminderConfig
// @Router       /reminder/config [get]
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value(user.ContextKeyUserID).(uint)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	cfg, err := h.service.GetOrCreateConfig(r.Context(), uid)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, cfg)
}

// UpdateConfig godoc
// @Summary      Cập nhật cấu hình nhắc nhở
// @Description  Bật/tắt nhắc nước, vận động, thay đổi interval, giờ yên tĩnh
// @Tags         Reminder
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      UpdateConfigRequest  true  "Cấu hình mới"
// @Success      200   {object}  ReminderConfig
// @Failure      400   {object}  map[string]string
// @Router       /reminder/config [put]
func (h *Handler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value(user.ContextKeyUserID).(uint)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	httputil.LimitBody(r)
	var req UpdateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cfg, err := h.service.UpdateConfig(r.Context(), uid, req)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, cfg)
}

// CheckStatus godoc
// @Summary      Kiểm tra trạng thái nhắc nhở
// @Description  Mobile gọi để kiểm tra đã đến lúc nhắc nước/vận động chưa. Nếu key Redis hết TTL → nhắc + reset timer.
// @Tags         Reminder
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  ReminderStatus
// @Router       /reminder/check [get]
func (h *Handler) CheckStatus(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value(user.ContextKeyUserID).(uint)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	status, err := h.service.CheckAndGetStatus(r.Context(), uid)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, status)
}
