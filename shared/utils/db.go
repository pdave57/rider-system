package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgres(dbName string) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		GetEnv("DB_HOST", "localhost"),
		GetEnv("DB_PORT", "5432"),
		GetEnv("DB_USER", "postgres"),
		GetEnv("DB_PASSWORD", ""),
		dbName,
		GetEnv("DB_SSLMODE", "disable"),
	)
	lvl := logger.Silent
	if GetEnv("APP_ENV", "production") == "development" {
		lvl = logger.Info
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(lvl)})
	if err != nil {
		log.Fatalf("[DB] connect %s: %v", dbName, err)
	}
	log.Printf("[DB] connected to %s", dbName)
	return db
}

func NewRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", GetEnv("REDIS_HOST", "localhost"), GetEnv("REDIS_PORT", "6379")),
		Password: GetEnv("REDIS_PASSWORD", ""),
	})
}

func GetEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
