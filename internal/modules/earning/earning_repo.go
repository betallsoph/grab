package earning

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

func (r *Repository) Migrate() error {
	return r.db.AutoMigrate(&Trip{})
}

func (r *Repository) CreateTrip(ctx context.Context, t *Trip) error {
	if err := r.db.WithContext(ctx).Create(t).Error; err != nil {
		return fmt.Errorf("repo.CreateTrip: %w", err)
	}
	return nil
}

func (r *Repository) ListTrips(ctx context.Context, driverID uint, dateStr string, limit, offset int) ([]Trip, error) {
	q := r.db.WithContext(ctx).
		Where("driver_id = ?", driverID).
		Order("trip_date DESC, created_at DESC")

	if dateStr != "" {
		q = q.Where("DATE(trip_date) = ?", dateStr)
	}

	if limit > 0 {
		q = q.Limit(limit)
	} else {
		q = q.Limit(50)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}

	var trips []Trip
	if err := q.Find(&trips).Error; err != nil {
		return nil, fmt.Errorf("repo.ListTrips: %w", err)
	}
	return trips, nil
}

// DailySummary tính tổng kết ngày bằng SQL aggregate.
func (r *Repository) DailySummary(ctx context.Context, driverID uint, dateStr string) (*DailySummary, error) {
	var s DailySummary
	err := r.db.WithContext(ctx).
		Table("trips").
		Select(`
			TO_CHAR(trip_date, 'YYYY-MM-DD') as date,
			COUNT(*) as trip_count,
			COALESCE(SUM(gross_amount), 0) as total_gross,
			COALESCE(SUM(platform_fee), 0) as total_fees,
			COALESCE(SUM(fuel_cost), 0) as total_fuel,
			COALESCE(SUM(other_cost), 0) as total_other,
			COALESCE(SUM(net_amount), 0) as total_net
		`).
		Where("driver_id = ? AND DATE(trip_date) = ?", driverID, dateStr).
		Scan(&s).Error
	if err != nil {
		return nil, fmt.Errorf("repo.DailySummary: %w", err)
	}
	if s.Date == "" {
		s.Date = dateStr
	}
	return &s, nil
}

// MonthlySummary tính tổng kết tháng bằng SQL aggregate.
func (r *Repository) MonthlySummary(ctx context.Context, driverID uint, monthStr string) (*MonthlySummary, error) {
	var s MonthlySummary
	err := r.db.WithContext(ctx).
		Table("trips").
		Select(`
			TO_CHAR(trip_date, 'YYYY-MM') as month,
			COUNT(*) as trip_count,
			COALESCE(SUM(gross_amount), 0) as total_gross,
			COALESCE(SUM(platform_fee), 0) as total_fees,
			COALESCE(SUM(fuel_cost), 0) as total_fuel,
			COALESCE(SUM(other_cost), 0) as total_other,
			COALESCE(SUM(net_amount), 0) as total_net
		`).
		Where("driver_id = ? AND TO_CHAR(trip_date, 'YYYY-MM') = ?", driverID, monthStr).
		Scan(&s).Error
	if err != nil {
		return nil, fmt.Errorf("repo.MonthlySummary: %w", err)
	}
	if s.Month == "" {
		s.Month = monthStr
	}
	// Tính trung bình lãi ròng mỗi ngày đã chạy trong tháng
	if s.TripCount > 0 {
		var activeDays int64
		r.db.WithContext(ctx).
			Table("trips").
			Where("driver_id = ? AND TO_CHAR(trip_date, 'YYYY-MM') = ?", driverID, monthStr).
			Select("COUNT(DISTINCT DATE(trip_date))").
			Scan(&activeDays)
		if activeDays > 0 {
			s.DailyAvgNet = s.TotalNet / float64(activeDays)
		}
	}
	return &s, nil
}
