package reminder

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Service struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewService(db *gorm.DB, rdb *redis.Client) *Service {
	return &Service{db: db, rdb: rdb}
}

// GetOrCreateConfig lấy config nhắc nhở, tạo mặc định nếu chưa có.
func (s *Service) GetOrCreateConfig(ctx context.Context, driverID uint) (*ReminderConfig, error) {
	var cfg ReminderConfig
	err := s.db.WithContext(ctx).
		Where("driver_id = ?", driverID).
		First(&cfg).Error

	if err == gorm.ErrRecordNotFound {
		cfg = ReminderConfig{DriverID: driverID}
		if err := s.db.WithContext(ctx).Create(&cfg).Error; err != nil {
			return nil, fmt.Errorf("reminder.CreateConfig: %w", err)
		}
		return &cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reminder.GetConfig: %w", err)
	}
	return &cfg, nil
}

// UpdateConfig cập nhật cấu hình nhắc nhở.
func (s *Service) UpdateConfig(ctx context.Context, driverID uint, req UpdateConfigRequest) (*ReminderConfig, error) {
	cfg, err := s.GetOrCreateConfig(ctx, driverID)
	if err != nil {
		return nil, err
	}

	if req.WaterEnabled != nil {
		cfg.WaterEnabled = *req.WaterEnabled
	}
	if req.WaterInterval != nil && *req.WaterInterval >= 15 {
		cfg.WaterInterval = *req.WaterInterval
	}
	if req.StretchEnabled != nil {
		cfg.StretchEnabled = *req.StretchEnabled
	}
	if req.StretchInterval != nil && *req.StretchInterval >= 15 {
		cfg.StretchInterval = *req.StretchInterval
	}
	if req.QuietStart != nil {
		cfg.QuietStart = *req.QuietStart
	}
	if req.QuietEnd != nil {
		cfg.QuietEnd = *req.QuietEnd
	}

	if err := s.db.WithContext(ctx).Save(cfg).Error; err != nil {
		return nil, fmt.Errorf("reminder.UpdateConfig: %w", err)
	}
	return cfg, nil
}

// CheckAndGetStatus kiểm tra có cần nhắc không dựa trên Redis TTL.
// Mobile gọi API này theo interval; nếu key hết hạn → đến lúc nhắc.
func (s *Service) CheckAndGetStatus(ctx context.Context, driverID uint) (*ReminderStatus, error) {
	cfg, err := s.GetOrCreateConfig(ctx, driverID)
	if err != nil {
		return nil, err
	}

	status := &ReminderStatus{DriverID: driverID}

	waterKey := fmt.Sprintf("reminder:water:%d", driverID)
	stretchKey := fmt.Sprintf("reminder:stretch:%d", driverID)

	// Nếu key không tồn tại → đã hết hạn → cần nhắc → tạo lại key với TTL
	if cfg.WaterEnabled {
		ttl, _ := s.rdb.TTL(ctx, waterKey).Result()
		if ttl <= 0 {
			interval := time.Duration(cfg.WaterInterval) * time.Minute
			s.rdb.Set(ctx, waterKey, "1", interval)
			status.NextWater = time.Now().Add(interval).Unix()
		} else {
			status.NextWater = time.Now().Add(ttl).Unix()
		}
	}

	if cfg.StretchEnabled {
		ttl, _ := s.rdb.TTL(ctx, stretchKey).Result()
		if ttl <= 0 {
			interval := time.Duration(cfg.StretchInterval) * time.Minute
			s.rdb.Set(ctx, stretchKey, "1", interval)
			status.NextStretch = time.Now().Add(interval).Unix()
		} else {
			status.NextStretch = time.Now().Add(ttl).Unix()
		}
	}

	return status, nil
}
