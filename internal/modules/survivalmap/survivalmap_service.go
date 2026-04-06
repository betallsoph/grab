package survivalmap

import (
	"context"
	"errors"
	"fmt"
)

const (
	defaultRadiusKm = 3.0
	maxRadiusKm     = 20.0
	defaultLimit    = 50
	maxLimit        = 100
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreatePOI validate đầu vào rồi lưu POI mới vào PostgreSQL.
func (s *Service) CreatePOI(ctx context.Context, userID uint, req CreatePOIRequest) (*POI, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if !req.Category.IsValid() {
		return nil, fmt.Errorf("invalid category: %s", req.Category)
	}
	if req.Latitude < -90 || req.Latitude > 90 || req.Longitude < -180 || req.Longitude > 180 {
		return nil, errors.New("invalid coordinates")
	}

	poi := &POI{
		Name:        req.Name,
		Category:    req.Category,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Address:     req.Address,
		Description: req.Description,
		CreatedBy:   userID,
	}

	if err := s.repo.Create(ctx, poi); err != nil {
		return nil, err
	}
	return poi, nil
}

// SearchNearby tìm POI xung quanh vị trí tài xế.
// Normalize các tham số query trước khi gọi repo.
func (s *Service) SearchNearby(ctx context.Context, q NearbyQuery) ([]POIWithDistance, error) {
	if q.Latitude < -90 || q.Latitude > 90 || q.Longitude < -180 || q.Longitude > 180 {
		return nil, errors.New("invalid coordinates")
	}

	if q.RadiusKm <= 0 {
		q.RadiusKm = defaultRadiusKm
	} else if q.RadiusKm > maxRadiusKm {
		q.RadiusKm = maxRadiusKm
	}

	if q.Category != "" && !q.Category.IsValid() {
		return nil, fmt.Errorf("invalid category: %s", q.Category)
	}

	if q.Limit <= 0 {
		q.Limit = defaultLimit
	} else if q.Limit > maxLimit {
		q.Limit = maxLimit
	}

	return s.repo.FindNearby(ctx, q)
}

// GetPOI trả về chi tiết một POI.
func (s *Service) GetPOI(ctx context.Context, id uint) (*POI, error) {
	return s.repo.FindByID(ctx, id)
}

// UpvotePOI cho phép tài xế upvote một POI hữu ích.
func (s *Service) UpvotePOI(ctx context.Context, poiID uint) error {
	return s.repo.Upvote(ctx, poiID)
}
