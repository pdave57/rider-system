package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/rydex/payment-service/handler"
	"github.com/rydex/payment-service/service"
	"github.com/rydex/shared/models"
	"github.com/rydex/shared/rabbitmq"
	"github.com/rydex/shared/utils"
)

func main() {
	_ = godotenv.Load("../../deploy/env/.env")
	db := utils.NewPostgres(utils.GetEnv("DB_NAME", "runns_payments"))
	db.AutoMigrate(&models.Payment{}, &models.ClientProfile{})

	rmqURL := utils.GetEnv("RABBITMQ_URL", "amqp://ryder:ryder_pass@localhost:5672/")
	rmqClient, err := rabbitmq.Connect(rmqURL)
	if err != nil {
		log.Printf("[payment-service] warning: rabbitmq connection failed: %v", err)
	} else {
		log.Printf("[payment-service] connected to rabbitmq")
		defer rmqClient.Close()
		_ = rmqClient.SetupExchangeQueue("rydex_events", "topic", "payment_events_queue", "order.*")
	}

	svc := service.NewPaymentService(db, rmqClient)
	h := handler.NewPaymentHandler(svc)
	port := utils.GetEnv("PAYMENT_SERVICE_PORT", "8084")
	log.Printf("[payment-service] :%s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler.NewRouter(h)); err != nil {
		log.Fatal(err)
	}
}
