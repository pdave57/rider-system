package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/runns/order-service/consumer"
	"github.com/runns/order-service/handler"
	"github.com/runns/order-service/service"
	"github.com/runns/shared/models"
	"github.com/runns/shared/rabbitmq"
	"github.com/runns/shared/utils"
)

func main() {
	utils.LoadEnvFile(".env", "../../deploy/env/.env")
	db := utils.NewPostgres(utils.GetEnv("DB_NAME", "runns_orders"))
	db.AutoMigrate(&models.User{}, &models.Order{})

	// Initialize RabbitMQ
	rmqConfig := rabbitmq.Config{
		URL:      os.Getenv("RABBITMQ_URL"),
		Exchange: rabbitmq.ExchangeOrderEvents,
	}
	rmqManager, err := rabbitmq.NewClient(rmqConfig)
	if err != nil {
		log.Fatal("Failed to initialize RabbitMQ:", err)
	}
	defer rmqManager.Close()

	// Initialize service
	orderService := service.NewOrderService(db, rmqManager)

	// Initialize auth event consumer
	authConsumer := consumer.NewAuthEventConsumer(rmqManager, orderService)
	go func() {
		if err := authConsumer.Start(); err != nil {
			log.Fatal("Failed to start auth consumer:", err)
		}
	}()

	// Initialize handler and router
	orderHandler := handler.NewOrderHandler(orderService)
	router := handler.NewRouter(orderHandler)

	// Start server
	server := &http.Server{
		Addr:    ":8082",
		Handler: router,
	}

	go func() {
		log.Println("Order service starting on :8082")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down order service...")

	if err := server.Close(); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}
}
