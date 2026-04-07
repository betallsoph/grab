package emergency

import (
	"encoding/json"
	"net/http"
	"strconv"

	"grab/internal/modules/user"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data) //nolint:errcheck
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// AddContact godoc
// @Summary      Thêm liên hệ khẩn cấp
// @Description  Thêm anh em thân thiết vào danh sách liên hệ khẩn cấp (tối đa 10)
// @Tags         Emergency
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      AddContactRequest  true  "Thông tin liên hệ"
// @Success      201   {object}  EmergencyContact
// @Failure      400   {object}  map[string]string
// @Router       /emergency/contacts [post]
func (h *Handler) AddContact(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value(user.ContextKeyUserID).(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req AddContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ec, err := h.service.AddContact(r.Context(), uid, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, ec)
}

// ListContacts godoc
// @Summary      Danh sách liên hệ khẩn cấp
// @Description  Lấy danh sách anh em thân thiết
// @Tags         Emergency
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}  EmergencyContact
// @Router       /emergency/contacts [get]
func (h *Handler) ListContacts(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value(user.ContextKeyUserID).(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	contacts, err := h.service.ListContacts(r.Context(), uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if contacts == nil {
		contacts = []EmergencyContact{}
	}
	writeJSON(w, http.StatusOK, contacts)
}

// RemoveContact godoc
// @Summary      Xóa liên hệ khẩn cấp
// @Description  Xóa anh em khỏi danh sách liên hệ khẩn cấp
// @Tags         Emergency
// @Produce      json
// @Security     BearerAuth
// @Param        contactId  path      int  true  "User ID của liên hệ cần xóa"
// @Success      200        {object}  map[string]string
// @Failure      404        {object}  map[string]string
// @Router       /emergency/contacts/{contactId} [delete]
func (h *Handler) RemoveContact(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value(user.ContextKeyUserID).(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cid, err := strconv.ParseUint(r.PathValue("contactId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid contact id")
		return
	}

	if err := h.service.RemoveContact(r.Context(), uid, uint(cid)); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "removed"})
}
