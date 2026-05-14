package service

import (
	"errors"

	"github.com/runns/shared/models"
	"gorm.io/gorm"
)

type RiderService struct{ db *gorm.DB }

func NewRiderService(db *gorm.DB) *RiderService { return &RiderService{db: db} }

func (s *RiderService) GetProfile(riderID uint) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("RiderProfile").First(&user, riderID).Error; err != nil {
		return nil, errors.New("rider not found")
	}
	return &user, nil
}

func (s *RiderService) UpdateAvailability(riderID uint, available bool) error {
	return s.db.Model(&models.RiderProfile{}).Where("user_id = ?", riderID).
		Update("is_available", available).Error
}

func (s *RiderService) UpdateLocation(riderID uint, lat, lon float64) error {
	return s.db.Model(&models.RiderProfile{}).Where("user_id = ?", riderID).
		Updates(map[string]interface{}{"current_latitude": lat, "current_longitude": lon}).Error
}

func (s *RiderService) ListAvailable() ([]models.User, error) {
	var riders []models.User
	s.db.Joins("JOIN rider_profiles ON rider_profiles.user_id = users.id").
		Where("rider_profiles.is_available = true").Preload("RiderProfile").Find(&riders)
	return riders, nil
}
