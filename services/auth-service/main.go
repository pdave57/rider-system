package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/runns/auth-service/handler"
	"github.com/runns/auth-service/service"
	"github.com/runns/shared/models"
	"github.com/runns/shared/rabbitmq"
	"github.com/runns/shared/utils"
)

func main() {
	_ = godotenv.Load("../../deploy/env/.env")
	db := utils.NewPostgres(utils.GetEnv("DB_NAME", "runns_auth"))
	db.AutoMigrate(&models.User{}, &models.RiderProfile{}, &models.ClientProfile{})	
	
	// Initialize RabbitMQ client
    rabbitConfig := rabbitmq.Config{
        URL:      os.Getenv("RABBITMQ_URL"),
        Exchange: "rydex.events",
    }
    
    rabbitClient, err := rabbitmq.NewClient(rabbitConfig)
    if err != nil {
        log.Fatal("Failed to connect to RabbitMQ:", err)
    }
    defer rabbitClient.Close()

    // Initialize services
    eventPublisher := service.NewEventPublisher(rabbitClient)

	svc := service.NewAuthService(db, eventPublisher)
	h := handler.NewAuthHandler(svc)
	port := utils.GetEnv("AUTH_SERVICE_PORT", "8081")
	log.Printf("[auth-service] :%s", port)
	
	// Start server in a goroutine so it doesn't block
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler.NewRouter(h)); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down...")
}
