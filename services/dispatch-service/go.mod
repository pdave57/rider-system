module github.com/rydex/dispatch-service

go 1.22

require (
	github.com/joho/godotenv v1.5.1
	github.com/rydex/shared v0.0.0
	github.com/redis/go-redis/v9 v9.5.1
	gorm.io/gorm v1.25.9
)

replace github.com/rydex/shared => ../../shared
