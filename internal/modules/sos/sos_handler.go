package sos

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"grab/internal/modules/user"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Mobile app nên CheckOrigin luôn cho phép; production nên kiểm soát chặt hơn.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Handler xử lý HTTP cho module SOS.
type Handler struct {
	hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

// ServeWS godoc
// @Summary      WebSocket SOS (Báo biến)
// @Description  Upgrade lên WebSocket. Sau khi kết nối, gửi {"type":"sos"} để broadcast báo biến cho tài xế trong 5km, {"type":"sos_cancel"} để hủy.
// @Tags         SOS
// @Security     BearerAuth
// @Router       /sos/ws [get]
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	// userID được inject bởi AuthMiddleware từ module user
	userID, ok := r.Context().Value(user.ContextKeyUserID).(uint)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[SOS] upgrade failed driver=%d err=%v", userID, err)
		return
	}

	h.hub.AddClient(userID, conn)
}
