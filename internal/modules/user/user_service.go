package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Service chứa toàn bộ logic nghiệp vụ của module User.
type Service struct {
	repo      *Repository
	redis     *redis.Client
	jwtSecret string
}

func NewService(repo *Repository, rdb *redis.Client, jwtSecret string) *Service {
	return &Service{
		repo:      repo,
		redis:     rdb,
		jwtSecret: jwtSecret,
	}
}

// Register tạo tài khoản mới. Password được hash bằng bcrypt trước khi lưu.
func (s *Service) Register(ctx context.Context, req RegisterRequest) error {
	if req.Phone == "" || req.Password == "" {
		return errors.New("phone and password are required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("service.Register hash: %w", err)
	}

	u := &User{
		Phone:        req.Phone,
		PasswordHash: string(hash),
		FullName:     req.FullName,
	}

	return fmt.Errorf("service.Register: %w", s.repo.Create(ctx, u))
}

// Login xác thực tài xế và trả về JWT nếu hợp lệ.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	if req.Phone == "" || req.Password == "" {
		return nil, errors.New("phone and password are required")
	}

	u, err := s.repo.FindByPhone(ctx, req.Phone)
	if err != nil {
		// Che giấu lý do cụ thể để tránh user enumeration attack
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("phone or password is incorrect")
		}
		return nil, fmt.Errorf("service.Login find: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("phone or password is incorrect")
	}

	token, err := s.generateJWT(u)
	if err != nil {
		return nil, fmt.Errorf("service.Login generate token: %w", err)
	}

	return &LoginResponse{
		Token: token,
		User: UserPublic{
			ID:       u.ID,
			Phone:    u.Phone,
			FullName: u.FullName,
		},
	}, nil
}

// UpdateLocation lưu vị trí GPS của tài xế vào Redis.
// Dùng hai cấu trúc song song:
//  1. GeoAdd vào key "drivers:geo"  → để module SOS truy vấn bán kính.
//  2. SET riêng mỗi tài xế với TTL 30s → mất tín hiệu 30s coi là offline.
func (s *Service) UpdateLocation(ctx context.Context, userID uint, req UpdateLocationRequest) error {
	driverKey := fmt.Sprintf("driver:loc:%d", userID)

	pipe := s.redis.Pipeline()

	pipe.GeoAdd(ctx, "drivers:geo", &redis.GeoLocation{
		Name:      fmt.Sprintf("%d", userID),
		Longitude: req.Longitude,
		Latitude:  req.Latitude,
	})

	// Lưu thêm dạng plain string để đọc nhanh khi không cần geo-query
	pipe.Set(ctx, driverKey,
		fmt.Sprintf("%.7f,%.7f", req.Latitude, req.Longitude),
		30*time.Second,
	)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("service.UpdateLocation redis: %w", err)
	}
	return nil
}

// --- JWT ---

// Claims là payload được nhúng vào JWT.
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func (s *Service) generateJWT(u *User) (string, error) {
	claims := Claims{
		UserID: u.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", u.ID),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// ParseJWT validate token string và trả về Claims.
// Được gọi bởi AuthMiddleware — export để các module khác dùng chung middleware.
func (s *Service) ParseJWT(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse jwt: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
