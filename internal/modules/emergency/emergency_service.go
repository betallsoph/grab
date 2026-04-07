package emergency

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

const maxContacts = 10

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// AddContact thêm anh em thân thiết.
func (s *Service) AddContact(ctx context.Context, driverID uint, req AddContactRequest) (*EmergencyContact, error) {
	if req.ContactID == 0 {
		return nil, errors.New("contact_id is required")
	}
	if req.ContactID == driverID {
		return nil, errors.New("cannot add yourself")
	}

	// Kiểm tra giới hạn
	var count int64
	s.db.WithContext(ctx).Model(&EmergencyContact{}).
		Where("driver_id = ?", driverID).
		Count(&count)
	if count >= maxContacts {
		return nil, fmt.Errorf("maximum %d emergency contacts reached", maxContacts)
	}

	ec := &EmergencyContact{
		DriverID:  driverID,
		ContactID: req.ContactID,
		Alias:     req.Alias,
	}
	if err := s.db.WithContext(ctx).Create(ec).Error; err != nil {
		return nil, fmt.Errorf("emergency.AddContact: %w", err)
	}
	return ec, nil
}

// ListContacts lấy danh sách anh em thân thiết.
func (s *Service) ListContacts(ctx context.Context, driverID uint) ([]EmergencyContact, error) {
	var contacts []EmergencyContact
	err := s.db.WithContext(ctx).
		Where("driver_id = ?", driverID).
		Order("created_at ASC").
		Find(&contacts).Error
	if err != nil {
		return nil, fmt.Errorf("emergency.ListContacts: %w", err)
	}
	return contacts, nil
}

// RemoveContact xóa một liên hệ khẩn cấp.
func (s *Service) RemoveContact(ctx context.Context, driverID, contactID uint) error {
	result := s.db.WithContext(ctx).
		Where("driver_id = ? AND contact_id = ?", driverID, contactID).
		Delete(&EmergencyContact{})
	if result.Error != nil {
		return fmt.Errorf("emergency.RemoveContact: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("contact not found")
	}
	return nil
}

// GetContactIDs trả về danh sách userID của anh em thân thiết.
// Dùng bởi module SOS khi cần gửi thông báo riêng.
func (s *Service) GetContactIDs(ctx context.Context, driverID uint) ([]uint, error) {
	var ids []uint
	err := s.db.WithContext(ctx).
		Model(&EmergencyContact{}).
		Where("driver_id = ?", driverID).
		Pluck("contact_id", &ids).Error
	if err != nil {
		return nil, fmt.Errorf("emergency.GetContactIDs: %w", err)
	}
	return ids, nil
}
