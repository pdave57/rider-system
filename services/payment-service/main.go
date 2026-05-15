package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/runns/payment-service/handler"
	"github.com/runns/payment-service/service"
	"github.com/runns/shared/models"
	"github.com/runns/shared/rabbitmq"
	"github.com/runns/shared/utils"
)

func main() {
	_ = godotenv.Load("../../deploy/env/.env")
	db := utils.NewPostgres(utils.GetEnv("DB_NAME", "runns_payments"))
	db.AutoMigrate(&models.Payment{}, &models.ClientProfile{})

	rmqClient, err := rabbitmq.NewClient()
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
