package earning

import "time"

// Trip ghi nhận một cuốc xe trong ngày.
type Trip struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	DriverID    uint      `gorm:"not null;index" json:"driver_id"`
	GrossAmount float64   `gorm:"not null" json:"gross_amount"`    // tổng thu
	PlatformFee float64   `gorm:"not null" json:"platform_fee"`    // phí nền tảng (%)
	FuelCost    float64   `gorm:"default:0" json:"fuel_cost"`      // xăng
	OtherCost   float64   `gorm:"default:0" json:"other_cost"`     // phí khác (gửi xe, phạt...)
	NetAmount   float64   `gorm:"not null" json:"net_amount"`      // lãi ròng = gross - fees - costs
	Note        string    `gorm:"size:500" json:"note,omitempty"`
	TripDate    time.Time `gorm:"not null;index" json:"trip_date"`
	CreatedAt   time.Time `json:"created_at"`
}

// DailySummary tổng kết một ngày.
type DailySummary struct {
	Date         string  `json:"date"` // YYYY-MM-DD
	TripCount    int     `json:"trip_count"`
	TotalGross   float64 `json:"total_gross"`
	TotalFees    float64 `json:"total_fees"`
	TotalFuel    float64 `json:"total_fuel"`
	TotalOther   float64 `json:"total_other"`
	TotalNet     float64 `json:"total_net"`
}

// MonthlySummary tổng kết một tháng.
type MonthlySummary struct {
	Month        string  `json:"month"` // YYYY-MM
	TripCount    int     `json:"trip_count"`
	TotalGross   float64 `json:"total_gross"`
	TotalFees    float64 `json:"total_fees"`
	TotalFuel    float64 `json:"total_fuel"`
	TotalOther   float64 `json:"total_other"`
	TotalNet     float64 `json:"total_net"`
	DailyAvgNet  float64 `json:"daily_avg_net"` // trung bình lãi ròng/ngày
}

// --- Request DTOs ---

type AddTripRequest struct {
	GrossAmount float64 `json:"gross_amount"`
	PlatformFee float64 `json:"platform_fee"`
	FuelCost    float64 `json:"fuel_cost"`
	OtherCost   float64 `json:"other_cost"`
	Note        string  `json:"note"`
	TripDate    string  `json:"trip_date"` // YYYY-MM-DD
}
