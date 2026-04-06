package sos

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	geoKey        = "drivers:geo" // key Redis Geo dùng chung với module User
	radiusKm      = 5.0          // bán kính quét SOS
	maxNearbyHits = 200          // giới hạn kết quả trả về
)

// Service chứa logic nghiệp vụ của module SOS.
type Service struct {
	rdb *redis.Client
	db  *gorm.DB
}

func NewService(rdb *redis.Client, db *gorm.DB) *Service {
	return &Service{rdb: rdb, db: db}
}

// FindNearbyDrivers dùng Redis GEOSEARCH tìm tất cả tài xế
// trong bán kính 5km xung quanh driverID.
//
// Dựa trên dữ liệu GeoAdd mà module User đã lưu ở key "drivers:geo".
func (s *Service) FindNearbyDrivers(ctx context.Context, driverID uint) ([]NearbyDriver, error) {
	memberName := fmt.Sprintf("%d", driverID)

	results, err := s.rdb.GeoSearchLocation(ctx, geoKey, &redis.GeoSearchLocationQuery{
		GeoSearchQuery: redis.GeoSearchQuery{
			Member:     memberName,
			Radius:     radiusKm,
			RadiusUnit: "km",
			Sort:       "ASC",
			Count:      maxNearbyHits,
		},
		WithCoord: true,
		WithDist:  true,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("sos.FindNearbyDrivers GEOSEARCH: %w", err)
	}

	nearby := make([]NearbyDriver, 0, len(results))
	for _, loc := range results {
		uid64, err := strconv.ParseUint(loc.Name, 10, 64)
		if err != nil {
			continue
		}
		uid := uint(uid64)

		// Loại chính mình ra khỏi kết quả
		if uid == driverID {
			continue
		}

		nearby = append(nearby, NearbyDriver{
			UserID:     uid,
			Latitude:   loc.Latitude,
			Longitude:  loc.Longitude,
			DistanceKm: loc.Dist,
		})
	}
	return nearby, nil
}

// GetDriverLocation đọc tọa độ hiện tại của một tài xế từ Redis Geo.
func (s *Service) GetDriverLocation(ctx context.Context, driverID uint) (lat, lng float64, err error) {
	memberName := fmt.Sprintf("%d", driverID)

	positions, err := s.rdb.GeoPos(ctx, geoKey, memberName).Result()
	if err != nil {
		return 0, 0, fmt.Errorf("sos.GetDriverLocation GeoPos: %w", err)
	}
	if len(positions) == 0 || positions[0] == nil {
		return 0, 0, fmt.Errorf("sos.GetDriverLocation: driver %d has no location", driverID)
	}

	return positions[0].Latitude, positions[0].Longitude, nil
}

// DriverInfo trả về thông tin cơ bản của tài xế để gửi kèm alert.
type DriverInfo struct {
	FullName string
	Phone    string
}

// GetDriverInfo đọc thông tin tài xế từ PostgreSQL.
// Chỉ lấy tên + SĐT để hiển thị trên màn hình người nhận SOS.
func (s *Service) GetDriverInfo(ctx context.Context, driverID uint) (*DriverInfo, error) {
	var info DriverInfo
	err := s.db.WithContext(ctx).
		Table("users").
		Select("full_name, phone").
		Where("id = ?", driverID).
		Scan(&info).Error
	if err != nil {
		return nil, fmt.Errorf("sos.GetDriverInfo: %w", err)
	}
	return &info, nil
}
