package chat

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"grab/internal/core/httputil"
	"grab/internal/modules/user"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Handler struct {
	service *Service
	hub     *Hub
}

func NewHandler(service *Service, hub *Hub) *Handler {
	return &Handler{service: service, hub: hub}
}

func getUserID(r *http.Request) (uint, bool) {
	uid, ok := r.Context().Value(user.ContextKeyUserID).(uint)
	return uid, ok
}

// CreateConversation godoc
// @Summary      Tạo conversation
// @Description  Tạo chat 1-1 (direct) hoặc nhóm (group). Direct giữa 2 người nếu đã tồn tại sẽ trả lại cái cũ.
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      CreateConversationRequest  true  "Thông tin conversation"
// @Success      201   {object}  Conversation
// @Failure      400   {object}  map[string]string
// @Router       /chat/conversations [post]
func (h *Handler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	uid, ok := getUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	httputil.LimitBody(r)
	var req CreateConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	conv, err := h.service.CreateConversation(r.Context(), uid, req)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, conv)
}

// ListConversations godoc
// @Summary      Danh sách conversation
// @Description  Lấy danh sách conversation mà tài xế tham gia, mới nhất trước
// @Tags         Chat
// @Produce      json
// @Security     BearerAuth
// @Param        limit   query  int  false  "Số kết quả (mặc định 30)"
// @Param        offset  query  int  false  "Offset phân trang"
// @Success      200     {array}  Conversation
// @Router       /chat/conversations [get]
func (h *Handler) ListConversations(w http.ResponseWriter, r *http.Request) {
	uid, ok := getUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	limit, offset = httputil.ClampPagination(limit, offset, 30, 100)

	convs, err := h.service.ListConversations(r.Context(), uid, limit, offset)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if convs == nil {
		convs = []Conversation{}
	}

	httputil.WriteJSON(w, http.StatusOK, convs)
}

// SendMessage godoc
// @Summary      Gửi tin nhắn (REST)
// @Description  Gửi tin nhắn vào conversation qua REST API (backup cho WebSocket)
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        convId  path      string              true  "Conversation ID (ObjectID hex)"
// @Param        body    body      SendMessageRequest  true  "Nội dung tin nhắn"
// @Success      201     {object}  Message
// @Failure      400     {object}  map[string]string
// @Router       /chat/conversations/{convId}/messages [post]
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	uid, ok := getUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	convID := r.PathValue("convId")

	httputil.LimitBody(r)
	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	msg, err := h.service.SendMessage(r.Context(), uid, convID, req)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, msg)
}

// GetMessages godoc
// @Summary      Lịch sử tin nhắn
// @Description  Lấy tin nhắn trong conversation, mới nhất trước
// @Tags         Chat
// @Produce      json
// @Security     BearerAuth
// @Param        convId  path   string  true   "Conversation ID"
// @Param        limit   query  int     false  "Số tin nhắn (mặc định 50)"
// @Param        offset  query  int     false  "Offset phân trang"
// @Success      200     {array}  Message
// @Router       /chat/conversations/{convId}/messages [get]
func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	uid, ok := getUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	convID := r.PathValue("convId")
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	limit, offset = httputil.ClampPagination(limit, offset, 50, 200)

	msgs, err := h.service.GetMessages(r.Context(), uid, convID, limit, offset)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if msgs == nil {
		msgs = []Message{}
	}

	httputil.WriteJSON(w, http.StatusOK, msgs)
}

// ServeWS godoc
// @Summary      WebSocket Chat
// @Description  Upgrade lên WebSocket cho real-time chat. Gửi {"type":"send","conversation_id":"...","content":"hello"} để chat, {"type":"typing","conversation_id":"..."} để báo đang gõ.
// @Tags         Chat
// @Security     BearerAuth
// @Router       /chat/ws [get]
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	uid, ok := getUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[Chat] upgrade failed user=%d err=%v", uid, err)
		return
	}

	h.hub.AddClient(uid, conn)
}
