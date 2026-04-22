module github.com/rydex/realtime-service

go 1.22

require (
	github.com/joho/godotenv v1.5.1
	github.com/rydex/shared v0.0.0
	github.com/redis/go-redis/v9 v9.5.1
)

replace github.com/rydex/shared => ../../shared
