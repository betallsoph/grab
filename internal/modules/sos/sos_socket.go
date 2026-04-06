package sos

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// --- Client ---

// Client đại diện cho một kết nối WebSocket của tài xế tới Hub SOS.
type Client struct {
	hub    *Hub
	userID uint
	conn   *websocket.Conn
	send   chan []byte // buffer gửi xuống client
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingInterval   = (pongWait * 9) / 10 // 54s — gửi ping trước khi pong timeout
	maxMessageSize = 512
)

// readPump đọc message từ WebSocket, dispatch theo type.
// Mỗi client chạy đúng 1 goroutine readPump.
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
				log.Printf("[SOS] readPump driver=%d err=%v", c.userID, err)
			}
			return
		}

		var msg IncomingMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Printf("[SOS] readPump driver=%d invalid JSON: %v", c.userID, err)
			continue
		}

		switch msg.Type {
		case MsgTypeSOS:
			c.hub.triggerSOS(c)
		case MsgTypeSOSCancel:
			c.hub.cancelSOS(c)
		default:
			log.Printf("[SOS] readPump driver=%d unknown type=%s", c.userID, msg.Type)
		}
	}
}

// writePump bơm message từ channel send xuống WebSocket.
// Mỗi client chạy đúng 1 goroutine writePump.
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
				// Hub đóng channel → gửi close frame
				c.conn.WriteMessage(websocket.CloseMessage, nil) //nolint:errcheck
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("[SOS] writePump driver=%d err=%v", c.userID, err)
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

// Hub quản lý toàn bộ kết nối WebSocket SOS.
// Kiến trúc: register/unregister qua channel, clients map được bảo vệ bởi mutex.
type Hub struct {
	clients    map[uint]*Client // userID -> *Client
	mu         sync.RWMutex
	register   chan *Client
	unregister chan *Client
	service    *Service
}

func NewHub(service *Service) *Hub {
	return &Hub{
		clients:    make(map[uint]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		service:    service,
	}
}

// Run là event loop chính của Hub, chạy trong goroutine riêng.
// Xử lý đăng ký / hủy đăng ký client.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// Nếu driver đã có kết nối cũ → đóng kết nối cũ (latest wins)
			if old, exists := h.clients[client.userID]; exists {
				close(old.send)
				delete(h.clients, old.userID)
			}
			h.clients[client.userID] = client
			h.mu.Unlock()
			log.Printf("[SOS] driver=%d connected (total=%d)", client.userID, h.clientCount())

		case client := <-h.unregister:
			h.mu.Lock()
			if current, exists := h.clients[client.userID]; exists && current == client {
				close(client.send)
				delete(h.clients, client.userID)
			}
			h.mu.Unlock()
			log.Printf("[SOS] driver=%d disconnected (total=%d)", client.userID, h.clientCount())
		}
	}
}

func (h *Hub) clientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// AddClient tạo Client mới, đăng ký vào Hub và bắt đầu read/write pump.
func (h *Hub) AddClient(userID uint, conn *websocket.Conn) {
	client := &Client{
		hub:    h,
		userID: userID,
		conn:   conn,
		send:   make(chan []byte, 64),
	}
	h.register <- client
	go client.writePump()
	go client.readPump()
}

// --- SOS logic ---

// triggerSOS xử lý khi một tài xế bấm nút báo biến.
//
// Flow:
//  1. Lấy tọa độ hiện tại của tài xế từ Redis
//  2. GEOSEARCH tìm tất cả tài xế trong bán kính 5km
//  3. Lấy thông tin cá nhân (tên, SĐT) để gửi kèm
//  4. Broadcast SOSAlert xuống từng client đang online trong danh sách
func (h *Hub) triggerSOS(sender *Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	lat, lng, err := h.service.GetDriverLocation(ctx, sender.userID)
	if err != nil {
		log.Printf("[SOS] triggerSOS driver=%d no location: %v", sender.userID, err)
		return
	}

	nearby, err := h.service.FindNearbyDrivers(ctx, sender.userID)
	if err != nil {
		log.Printf("[SOS] triggerSOS driver=%d GEOSEARCH failed: %v", sender.userID, err)
		return
	}

	info, err := h.service.GetDriverInfo(ctx, sender.userID)
	if err != nil {
		log.Printf("[SOS] triggerSOS driver=%d get info failed: %v", sender.userID, err)
		info = &DriverInfo{FullName: "Tài xế", Phone: "N/A"}
	}

	log.Printf("[SOS] 🚨 driver=%d triggered SOS at (%.5f, %.5f), found %d nearby",
		sender.userID, lat, lng, len(nearby))

	now := time.Now().Unix()

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, nd := range nearby {
		recipient, online := h.clients[nd.UserID]
		if !online {
			continue
		}

		alert := SOSAlert{
			Type:       MsgTypeSOSAlert,
			FromDriver: sender.userID,
			FullName:   info.FullName,
			Phone:      info.Phone,
			Latitude:   lat,
			Longitude:  lng,
			DistanceKm: nd.DistanceKm,
			Timestamp:  now,
		}

		data, err := json.Marshal(alert)
		if err != nil {
			continue
		}

		// Non-blocking send: nếu buffer đầy thì skip client này
		// để tránh 1 client chậm block toàn bộ broadcast
		select {
		case recipient.send <- data:
		default:
			log.Printf("[SOS] broadcast skipped driver=%d (send buffer full)", nd.UserID)
		}
	}
}

// cancelSOS broadcast thông báo hủy SOS cho các tài xế gần đó.
func (h *Hub) cancelSOS(sender *Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	nearby, err := h.service.FindNearbyDrivers(ctx, sender.userID)
	if err != nil {
		log.Printf("[SOS] cancelSOS driver=%d GEOSEARCH failed: %v", sender.userID, err)
		return
	}

	log.Printf("[SOS] driver=%d cancelled SOS", sender.userID)

	cancelMsg, _ := json.Marshal(map[string]any{
		"type":        MsgTypeSOSCancel,
		"from_driver": sender.userID,
		"timestamp":   time.Now().Unix(),
	})

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, nd := range nearby {
		if recipient, online := h.clients[nd.UserID]; online {
			select {
			case recipient.send <- cancelMsg:
			default:
			}
		}
	}
}
