package chat

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// --- Client ---

type Client struct {
	hub    *Hub
	userID uint
	conn   *websocket.Conn
	send   chan []byte
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingInterval   = (pongWait * 9) / 10
	maxMessageSize = 4096 // chat message lớn hơn SOS
)

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait)) //nolint:errcheck
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait)) //nolint:errcheck
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("[Chat] readPump user=%d err=%v", c.userID, err)
			}
			return
		}

		var incoming WSIncoming
		if err := json.Unmarshal(raw, &incoming); err != nil {
			log.Printf("[Chat] readPump user=%d invalid JSON: %v", c.userID, err)
			continue
		}

		switch incoming.Type {
		case WSTypeSend:
			c.hub.handleSend(c, incoming)
		case WSTypeTyping:
			c.hub.handleTyping(c, incoming)
		default:
			log.Printf("[Chat] readPump user=%d unknown type=%s", c.userID, incoming.Type)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case data, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait)) //nolint:errcheck
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, nil) //nolint:errcheck
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait)) //nolint:errcheck
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// --- Hub ---

// Hub quản lý tất cả kết nối WebSocket chat.
// Một user có thể có nhiều kết nối (nhiều thiết bị).
type Hub struct {
	clients    map[uint]map[*Client]bool // userID -> set of clients
	mu         sync.RWMutex
	register   chan *Client
	unregister chan *Client
	service    *Service
}

func NewHub(service *Service) *Hub {
	return &Hub{
		clients:    make(map[uint]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		service:    service,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.userID] == nil {
				h.clients[client.userID] = make(map[*Client]bool)
			}
			h.clients[client.userID][client] = true
			h.mu.Unlock()
			log.Printf("[Chat] user=%d connected (devices=%d)", client.userID, len(h.clients[client.userID]))

		case client := <-h.unregister:
			h.mu.Lock()
			if conns, ok := h.clients[client.userID]; ok {
				if _, exists := conns[client]; exists {
					close(client.send)
					delete(conns, client)
					if len(conns) == 0 {
						delete(h.clients, client.userID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("[Chat] user=%d disconnected", client.userID)
		}
	}
}

// AddClient đăng ký client mới vào Hub.
func (h *Hub) AddClient(userID uint, conn *websocket.Conn) {
	client := &Client{
		hub:    h,
		userID: userID,
		conn:   conn,
		send:   make(chan []byte, 128),
	}
	h.register <- client
	go client.writePump()
	go client.readPump()
}

// handleSend xử lý khi client gửi tin nhắn qua WebSocket.
// Lưu vào MongoDB rồi push real-time cho tất cả participant online.
func (h *Hub) handleSend(sender *Client, incoming WSIncoming) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msg, err := h.service.SendMessage(ctx, sender.userID, incoming.ConversationID, SendMessageRequest{
		Content:     incoming.Content,
		ContentType: incoming.ContentType,
	})
	if err != nil {
		log.Printf("[Chat] handleSend user=%d err=%v", sender.userID, err)
		return
	}

	// Lấy danh sách participant để broadcast
	conv, err := h.service.GetConversation(ctx, sender.userID, incoming.ConversationID)
	if err != nil {
		return
	}

	outgoing := WSOutgoing{
		Type:           WSTypeNew,
		ConversationID: incoming.ConversationID,
		Message:        msg,
		SenderID:       sender.userID,
		Timestamp:      time.Now().Unix(),
	}
	data, err := json.Marshal(outgoing)
	if err != nil {
		return
	}

	h.broadcastToParticipants(conv.Participants, data)
}

// handleTyping forward sự kiện "đang gõ" cho các participant khác.
func (h *Hub) handleTyping(sender *Client, incoming WSIncoming) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conv, err := h.service.GetConversation(ctx, sender.userID, incoming.ConversationID)
	if err != nil {
		return
	}

	outgoing := WSOutgoing{
		Type:           WSTypeTyping,
		ConversationID: incoming.ConversationID,
		SenderID:       sender.userID,
		Timestamp:      time.Now().Unix(),
	}
	data, err := json.Marshal(outgoing)
	if err != nil {
		log.Printf("[Chat] handleTyping user=%d marshal err=%v", sender.userID, err)
		return
	}

	// Gửi cho tất cả participant trừ sender
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, uid := range conv.Participants {
		if uid == sender.userID {
			continue
		}
		if conns, ok := h.clients[uid]; ok {
			for c := range conns {
				select {
				case c.send <- data:
				default:
				}
			}
		}
	}
}

// broadcastToParticipants gửi data cho tất cả participant online (bao gồm sender, để sync nhiều thiết bị).
func (h *Hub) broadcastToParticipants(participants []uint, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, uid := range participants {
		if conns, ok := h.clients[uid]; ok {
			for c := range conns {
				select {
				case c.send <- data:
				default:
					log.Printf("[Chat] broadcast skip user=%d (buffer full)", uid)
				}
			}
		}
	}
}
