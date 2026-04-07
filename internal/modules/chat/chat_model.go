package chat

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// --- MongoDB collections: "conversations" + "messages" ---

type ConversationType string

const (
	ConvDirect ConversationType = "direct" // Chat 1-1
	ConvGroup  ConversationType = "group"  // Chat nhóm
)

// Conversation lưu trong MongoDB collection "conversations".
type Conversation struct {
	ID           bson.ObjectID    `bson:"_id,omitempty" json:"id"`
	Type         ConversationType `bson:"type" json:"type"`
	Name         string           `bson:"name,omitempty" json:"name,omitempty"` // tên nhóm, rỗng nếu direct
	Participants []uint           `bson:"participants" json:"participants"`
	CreatedBy    uint             `bson:"created_by" json:"created_by"`
	CreatedAt    time.Time        `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time        `bson:"updated_at" json:"updated_at"`
}

// Message lưu trong MongoDB collection "messages".
type Message struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	ConversationID bson.ObjectID `bson:"conversation_id" json:"conversation_id"`
	SenderID       uint          `bson:"sender_id" json:"sender_id"`
	Content        string        `bson:"content" json:"content"`
	ContentType    string        `bson:"content_type" json:"content_type"` // "text", "image", "voice"
	CreatedAt      time.Time     `bson:"created_at" json:"created_at"`
}

// --- Request DTOs ---

type CreateConversationRequest struct {
	Type         ConversationType `json:"type"`
	Name         string           `json:"name,omitempty"`
	Participants []uint           `json:"participants"`
}

type SendMessageRequest struct {
	Content     string `json:"content"`
	ContentType string `json:"content_type"` // "text" nếu rỗng
}

// --- WebSocket message types ---

const (
	WSTypeSend    = "send"    // client → server: gửi tin nhắn
	WSTypeNew     = "new"     // server → client: tin nhắn mới
	WSTypeTyping  = "typing"  // client ↔ server: đang gõ
	WSTypeOnline  = "online"  // server → client: user online/offline
)

// WSIncoming là message client gửi lên qua WebSocket.
type WSIncoming struct {
	Type           string `json:"type"`
	ConversationID string `json:"conversation_id"`
	Content        string `json:"content,omitempty"`
	ContentType    string `json:"content_type,omitempty"`
}

// WSOutgoing là message server đẩy xuống client qua WebSocket.
type WSOutgoing struct {
	Type           string  `json:"type"`
	ConversationID string  `json:"conversation_id"`
	Message        *Message `json:"message,omitempty"`
	SenderID       uint    `json:"sender_id,omitempty"`
	Timestamp      int64   `json:"timestamp"`
}
