package survivalmap

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Migrate tạo bảng POI và spatial index.
// Gọi một lần khi khởi động server.
func (r *Repository) Migrate() error {
	if err := r.db.AutoMigrate(&POI{}); err != nil {
		return fmt.Errorf("survivalmap migrate: %w", err)
	}

	// Tạo spatial index trên expression ST_MakePoint.
	// PostGIS dùng geography cast để tính khoảng cách chính xác trên mặt cầu.
	// "IF NOT EXISTS" để tránh lỗi khi chạy lại.
	sql := `CREATE INDEX IF NOT EXISTS idx_pois_location
		ON pois USING GIST (
			ST_MakePoint(longitude, latitude)::geography
		)`
	return r.db.Exec(sql).Error
}

// Create thêm một POI mới.
func (r *Repository) Create(ctx context.Context, poi *POI) error {
	if err := r.db.WithContext(ctx).Create(poi).Error; err != nil {
		return fmt.Errorf("repo.CreatePOI: %w", err)
	}
	return nil
}

// FindNearby tìm các POI trong bán kính radiusKm xung quanh (lat, lng).
//
// Sử dụng PostGIS:
//   - ST_MakePoint(lng, lat)::geography  → tạo geography point
//   - ST_DWithin(a, b, meters)           → lọc trong bán kính (tính bằng mét)
//   - ST_Distance(a, b)                  → tính khoảng cách chính xác (mét)
//
// Kết quả sắp xếp theo khoảng cách tăng dần.
func (r *Repository) FindNearby(ctx context.Context, q NearbyQuery) ([]POIWithDistance, error) {
	radiusM := q.RadiusKm * 1000

	query := r.db.WithContext(ctx).
		Table("pois").
		Select(`pois.*,
			ST_Distance(
				ST_MakePoint(longitude, latitude)::geography,
				ST_MakePoint(?, ?)::geography
			) AS distance_m`, q.Longitude, q.Latitude).
		Where(`ST_DWithin(
			ST_MakePoint(longitude, latitude)::geography,
			ST_MakePoint(?, ?)::geography,
			?
		)`, q.Longitude, q.Latitude, radiusM).
		Order("distance_m ASC")

	if q.Category != "" {
		query = query.Where("category = ?", q.Category)
	}

	if q.Limit > 0 {
		query = query.Limit(q.Limit)
	} else {
		query = query.Limit(50)
	}
	if q.Offset > 0 {
		query = query.Offset(q.Offset)
	}

	var results []POIWithDistance
	if err := query.Find(&results).Error; err != nil {
		return nil, fmt.Errorf("repo.FindNearby: %w", err)
	}
	return results, nil
}

// FindByID tìm POI theo ID.
func (r *Repository) FindByID(ctx context.Context, id uint) (*POI, error) {
	var poi POI
	if err := r.db.WithContext(ctx).First(&poi, id).Error; err != nil {
		return nil, fmt.Errorf("repo.FindPOIByID: %w", err)
	}
	return &poi, nil
}

// Upvote tăng upvote_count thêm 1.
func (r *Repository) Upvote(ctx context.Context, poiID uint) error {
	result := r.db.WithContext(ctx).
		Model(&POI{}).
		Where("id = ?", poiID).
		UpdateColumn("upvote_count", gorm.Expr("upvote_count + 1"))
	if result.Error != nil {
		return fmt.Errorf("repo.Upvote: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("repo.Upvote: poi %d not found", poiID)
	}
	return nil
}
