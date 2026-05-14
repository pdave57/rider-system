package service

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/runns/shared/middleware"
	"github.com/runns/shared/models"
)

func generateToken(userID uint, email string, role models.Role) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET not set")
	}
	claims := middleware.Claims{
		UserID: userID, Email: email, Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(48 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   email,
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}
