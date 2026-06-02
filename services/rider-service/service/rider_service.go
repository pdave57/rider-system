package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/runns/shared/models"
	"gorm.io/gorm"
)

type RiderService struct {
	db     *gorm.DB       // runns_riders — RiderProfile records
	authDB *gorm.DB       // runns_auth   — User records
	rdb    *redis.Client  // Redis — geo index for dispatch
}

func NewRiderService(db, authDB *gorm.DB, rdb *redis.Client) *RiderService {
	return &RiderService{db: db, authDB: authDB, rdb: rdb}
}

func (s *RiderService) ensureProfileExists(riderID uint) error {
	var profile models.RiderProfile
	err := s.db.Where("user_id = ?", riderID).First(&profile).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Fetch registered vehicle details from authDB's rider_profiles
	var authProfile models.RiderProfile
	var vehicleType models.VehicleType
	var vehiclePlate string
	var licenseNumber string
	if s.authDB.Where("user_id = ?", riderID).First(&authProfile).Error == nil {
		vehicleType = authProfile.VehicleType
		vehiclePlate = authProfile.VehiclePlate
		licenseNumber = authProfile.LicenseNumber
	}

	// Create new profile in riders DB (runns_riders)
	profile = models.RiderProfile{
		UserID:        riderID,
		VehicleType:   vehicleType,
		VehiclePlate:  vehiclePlate,
		LicenseNumber: licenseNumber,
		IsAvailable:   false,
	}
	return s.db.Create(&profile).Error
}

func (s *RiderService) GetProfile(riderID uint) (*models.User, error) {
	// Load user from auth DB
	var user models.User
	if err := s.authDB.First(&user, riderID).Error; err != nil {
		return nil, errors.New("rider not found")
	}
	
	// Lazily ensure profile exists in riders DB
	_ = s.ensureProfileExists(riderID)

	// Load rider profile from riders DB
	var profile models.RiderProfile
	if err := s.db.Where("user_id = ?", riderID).First(&profile).Error; err == nil {
		user.RiderProfile = &profile
	}
	return &user, nil
}

func (s *RiderService) UpdateAvailability(riderID uint, available bool) error {
	if err := s.ensureProfileExists(riderID); err != nil {
		return err
	}
	return s.db.Model(&models.RiderProfile{}).Where("user_id = ?", riderID).
		Update("is_available", available).Error
}

func (s *RiderService) UpdateLocation(riderID uint, lat, lon float64) error {
	if err := s.ensureProfileExists(riderID); err != nil {
		return err
	}
	// Update DB
	if err := s.db.Model(&models.RiderProfile{}).Where("user_id = ?", riderID).
		Updates(map[string]interface{}{"current_latitude": lat, "current_longitude": lon}).Error; err != nil {
		return err
	}
	// Update Redis geo index so dispatch service can find this rider
	ctx := context.Background()
	return s.rdb.GeoAdd(ctx, "rider:locations", &redis.GeoLocation{
		Name:      fmt.Sprintf("%d", riderID),
		Latitude:  lat,
		Longitude: lon,
	}).Err()
}

func (s *RiderService) ListAvailable() ([]models.User, error) {
	// Get available rider IDs from riders DB
	var profiles []models.RiderProfile
	s.db.Where("is_available = true").Find(&profiles)

	if len(profiles) == 0 {
		return []models.User{}, nil
	}

	// Build ID list
	ids := make([]uint, len(profiles))
	profileMap := make(map[uint]*models.RiderProfile)
	for i, p := range profiles {
		ids[i] = p.UserID
		cp := profiles[i]
		profileMap[p.UserID] = &cp
	}

	// Fetch user records from auth DB
	var users []models.User
	s.authDB.Where("id IN ?", ids).Find(&users)

	// Attach profiles
	for i := range users {
		if p, ok := profileMap[users[i].ID]; ok {
			users[i].RiderProfile = p
		}
	}
	return users, nil
}
