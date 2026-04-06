package survivalmap

import "time"

// --- POI Categories ---

type Category string

const (
	CatFood   Category = "food"   // Quán ăn, cơm bụi
	CatRepair Category = "repair" // Sửa xe, vá vỏ
	CatCafe   Category = "cafe"   // Cà phê võng, nước mía
	CatWC     Category = "wc"     // Nhà vệ sinh công cộng
	CatGas    Category = "gas"    // Đổ xăng
	CatRest   Category = "rest"   // Chỗ nghỉ, ghế đá, bóng mát
	CatOther  Category = "other"
)

var validCategories = map[Category]bool{
	CatFood: true, CatRepair: true, CatCafe: true,
	CatWC: true, CatGas: true, CatRest: true, CatOther: true,
}

func (c Category) IsValid() bool {
	return validCategories[c]
}

// --- Database model ---

// POI (Point of Interest) lưu trong PostgreSQL.
// Tọa độ được index bằng PostGIS để truy vấn khoảng cách.
type POI struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"not null;size:200" json:"name"`
	Category    Category  `gorm:"not null;size:20;index" json:"category"`
	Latitude    float64   `gorm:"not null" json:"latitude"`
	Longitude   float64   `gorm:"not null" json:"longitude"`
	Address     string    `gorm:"size:500" json:"address"`
	Description string    `gorm:"size:1000" json:"description"`
	CreatedBy   uint      `gorm:"not null;index" json:"created_by"` // userID của tài xế đóng góp
	UpvoteCount int       `gorm:"default:0" json:"upvote_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// --- Request / Response DTOs ---

type CreatePOIRequest struct {
	Name        string   `json:"name"`
	Category    Category `json:"category"`
	Latitude    float64  `json:"latitude"`
	Longitude   float64  `json:"longitude"`
	Address     string   `json:"address"`
	Description string   `json:"description"`
}

// NearbyQuery là tham số truy vấn POI xung quanh tài xế.
type NearbyQuery struct {
	Latitude  float64  // vị trí tài xế
	Longitude float64
	RadiusKm  float64  // bán kính tìm kiếm (mặc định 3km)
	Category  Category // lọc theo loại, rỗng = tất cả
	Limit     int
	Offset    int
}

// POIWithDistance là POI kèm khoảng cách đến tài xế (tính bằng mét).
type POIWithDistance struct {
	POI
	DistanceM float64 `json:"distance_m"`
}

type UpvoteRequest struct {
	POIID uint `json:"poi_id"`
}
