package reminder

import "time"

// ReminderConfig lưu cấu hình nhắc nhở của tài xế trong PostgreSQL.
type ReminderConfig struct {
	ID             uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	DriverID       uint   `gorm:"uniqueIndex;not null" json:"driver_id"`
	WaterEnabled   bool   `gorm:"default:true" json:"water_enabled"`
	WaterInterval  int    `gorm:"default:60" json:"water_interval_min"` // phút
	StretchEnabled bool   `gorm:"default:true" json:"stretch_enabled"`
	StretchInterval int   `gorm:"default:90" json:"stretch_interval_min"`
	QuietStart     string `gorm:"size:5;default:'22:00'" json:"quiet_start"` // HH:MM — không nhắc ban đêm
	QuietEnd       string `gorm:"size:5;default:'06:00'" json:"quiet_end"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ReminderStatus là trạng thái nhắc nhở hiện tại (lưu Redis, TTL = interval).
type ReminderStatus struct {
	DriverID     uint   `json:"driver_id"`
	NextWater    int64  `json:"next_water_at,omitempty"`    // unix timestamp
	NextStretch  int64  `json:"next_stretch_at,omitempty"`
}

type UpdateConfigRequest struct {
	WaterEnabled    *bool   `json:"water_enabled,omitempty"`
	WaterInterval   *int    `json:"water_interval_min,omitempty"`
	StretchEnabled  *bool   `json:"stretch_enabled,omitempty"`
	StretchInterval *int    `json:"stretch_interval_min,omitempty"`
	QuietStart      *string `json:"quiet_start,omitempty"`
	QuietEnd        *string `json:"quiet_end,omitempty"`
}
