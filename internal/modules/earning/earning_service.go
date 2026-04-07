package earning

import (
	"context"
	"errors"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// AddTrip ghi nhận một cuốc xe. Tự tính net_amount.
func (s *Service) AddTrip(ctx context.Context, driverID uint, req AddTripRequest) (*Trip, error) {
	if req.GrossAmount <= 0 {
		return nil, errors.New("gross_amount must be positive")
	}

	tripDate, err := time.Parse("2006-01-02", req.TripDate)
	if err != nil {
		return nil, errors.New("trip_date must be YYYY-MM-DD format")
	}

	// Tính lãi ròng
	net := req.GrossAmount - req.PlatformFee - req.FuelCost - req.OtherCost

	trip := &Trip{
		DriverID:    driverID,
		GrossAmount: req.GrossAmount,
		PlatformFee: req.PlatformFee,
		FuelCost:    req.FuelCost,
		OtherCost:   req.OtherCost,
		NetAmount:   net,
		Note:        req.Note,
		TripDate:    tripDate,
	}
	if err := s.repo.CreateTrip(ctx, trip); err != nil {
		return nil, err
	}
	return trip, nil
}

// ListTrips lấy danh sách cuốc xe.
func (s *Service) ListTrips(ctx context.Context, driverID uint, dateStr string, limit, offset int) ([]Trip, error) {
	return s.repo.ListTrips(ctx, driverID, dateStr, limit, offset)
}

// DailySummary tổng kết ngày.
func (s *Service) DailySummary(ctx context.Context, driverID uint, dateStr string) (*DailySummary, error) {
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}
	return s.repo.DailySummary(ctx, driverID, dateStr)
}

// MonthlySummary tổng kết tháng.
func (s *Service) MonthlySummary(ctx context.Context, driverID uint, monthStr string) (*MonthlySummary, error) {
	if monthStr == "" {
		monthStr = time.Now().Format("2006-01")
	}
	return s.repo.MonthlySummary(ctx, driverID, monthStr)
}
