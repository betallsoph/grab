package emergency

import "time"

// EmergencyContact lưu danh sách anh em thân thiết trong PostgreSQL.
// Khi SOS, hệ thống gửi thông báo riêng cho nhóm này ngoài broadcast 5km.
type EmergencyContact struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	DriverID    uint      `gorm:"not null;index:idx_driver_contact,unique" json:"driver_id"`
	ContactID   uint      `gorm:"not null;index:idx_driver_contact,unique" json:"contact_id"` // userID của anh em
	Alias       string    `gorm:"size:100" json:"alias"`                                       // biệt danh
	CreatedAt   time.Time `json:"created_at"`
}

type AddContactRequest struct {
	ContactID uint   `json:"contact_id"`
	Alias     string `json:"alias"`
}
