module github.com/rydex/auth-service

go 1.22

require (
	github.com/joho/godotenv v1.5.1
	github.com/rydex/shared v0.0.0
	gorm.io/gorm v1.25.9
	github.com/golang-jwt/jwt/v5 v5.2.1
	golang.org/x/crypto v0.22.0
)

replace github.com/rydex/shared => ../../shared
