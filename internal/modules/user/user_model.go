package user

import "time"

// User là bảng lưu thông tin tài xế trong PostgreSQL.
type User struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"`
	Phone        string    `gorm:"uniqueIndex;not null;size:20"`
	PasswordHash string    `gorm:"not null"`
	FullName     string    `gorm:"size:100"`
	AvatarURL    string    `gorm:"size:500"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// --- Request / Response DTOs ---

type RegisterRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type LoginRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string     `json:"token"`
	User  UserPublic `json:"user"`
}

// UserPublic là thông tin an toàn để trả về client (không có PasswordHash).
type UserPublic struct {
	ID       uint   `json:"id"`
	Phone    string `json:"phone"`
	FullName string `json:"full_name"`
}

// UpdateLocationRequest là payload từ mobile gửi lên mỗi vài giây.
type UpdateLocationRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// contextKey là kiểu riêng để tránh xung đột key trong context.
type contextKey string

// ContextKeyUserID là key để lưu userID vào request context sau khi xác thực JWT.
// Được export để các module khác (SOS, Chat...) dùng chung.
const ContextKeyUserID contextKey = "userID"
