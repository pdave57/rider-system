package dto

import "github.com/runns/shared/models"

type RegisterRequest struct {
	FullName       string             `json:"full_name"`
	Email          string             `json:"email"`
	Phone          string             `json:"phone"`
	Password       string             `json:"password"`
	Role           models.Role        `json:"role"`
	VehicleType    models.VehicleType `json:"vehicle_type,omitempty"`
	VehiclePlate   string             `json:"vehicle_plate,omitempty"`
	LicenseNumber  string             `json:"license_number,omitempty"`
	DefaultAddress string             `json:"default_address,omitempty"`
	Latitude       float64            `json:"latitude,omitempty"`
	Longitude      float64            `json:"longitude,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID       uint        `json:"id"`
	FullName string      `json:"full_name"`
	Email    string      `json:"email"`
	Phone    string      `json:"phone"`
	Role     models.Role `json:"role"`
	IsActive bool        `json:"is_active"`
}

func ToUserResponse(u *models.User) UserResponse {
	return UserResponse{ID: u.ID, FullName: u.FullName, Email: u.Email,
		Phone: u.Phone, Role: u.Role, IsActive: u.IsActive}
}
