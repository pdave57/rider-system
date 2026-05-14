package service

import (
	"errors"

	"github.com/runns/auth-service/dto"
	"github.com/runns/shared/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

//initiate dependency injection
type AuthService struct{ db *gorm.DB }

//dependency injection
func NewAuthService(db *gorm.DB) *AuthService { return &AuthService{db: db} }

func (s *AuthService) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	if req.FullName == "" || req.Email == "" || req.Phone == "" || req.Password == "" {
		return nil, errors.New("full_name, email, phone and password are required")
	}
	if req.Role != models.RoleClient && req.Role != models.RoleRider && req.Role != models.RoleAdmin {
		return nil, errors.New("role must be client, rider or admin")
	}
	var existing models.User
	if err := s.db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return nil, errors.New("email already registered")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}
	user := &models.User{
		FullName: req.FullName, 
		Email: req.Email,
		Phone: req.Phone, 
		Password: string(hashed), 
		Role: req.Role, 
		IsActive: true,
	}
	switch req.Role {
	case models.RoleRider:
		user.RiderProfile = &models.RiderProfile{VehicleType: req.VehicleType, VehiclePlate: req.VehiclePlate}
	case models.RoleClient:
		user.ClientProfile = &models.ClientProfile{}
	}
	if err := s.db.Create(user).Error; err != nil {
		return nil, errors.New("failed to create user: " + err.Error())
	}
	token, err := generateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}
	return &dto.AuthResponse{Token: token, User: dto.ToUserResponse(user)}, nil
}

func (s *AuthService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password are required")
	}
	var user models.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, errors.New("invalid credentials")
	}
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	token, err := generateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}
	return &dto.AuthResponse{Token: token, User: dto.ToUserResponse(&user)}, nil
}

func (s *AuthService) ChangePassword(userID uint, req dto.ChangePasswordRequest) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	return s.db.Model(&user).Update("password", string(hashed)).Error
}
