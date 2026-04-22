module github.com/rydex/payment-service

go 1.22

require (
	github.com/joho/godotenv v1.5.1
	github.com/rydex/shared v0.0.0
	gorm.io/gorm v1.25.9
)

replace github.com/rydex/shared => ../../shared
