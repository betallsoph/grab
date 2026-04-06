package user

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// Repository xử lý toàn bộ truy vấn database cho User.
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create thêm một user mới vào PostgreSQL.
func (r *Repository) Create(ctx context.Context, u *User) error {
	if err := r.db.WithContext(ctx).Create(u).Error; err != nil {
		return fmt.Errorf("repo.Create: %w", err)
	}
	return nil
}

// FindByPhone tìm user theo số điện thoại.
// Trả về gorm.ErrRecordNotFound nếu không tìm thấy.
func (r *Repository) FindByPhone(ctx context.Context, phone string) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).
		Where("phone = ?", phone).
		First(&u).Error
	if err != nil {
		return nil, fmt.Errorf("repo.FindByPhone: %w", err)
	}
	return &u, nil
}

// FindByID tìm user theo ID.
func (r *Repository) FindByID(ctx context.Context, id uint) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).First(&u, id).Error
	if err != nil {
		return nil, fmt.Errorf("repo.FindByID: %w", err)
	}
	return &u, nil
}
