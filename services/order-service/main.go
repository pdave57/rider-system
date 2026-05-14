package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/runns/order-service/handler"
	"github.com/runns/order-service/service"
	"github.com/runns/shared/models"
	"github.com/runns/shared/rabbitmq"
	"github.com/runns/shared/utils"
)

func main() {
	_ = godotenv.Load("../../deploy/env/.env")
	db := utils.NewPostgres(utils.GetEnv("DB_NAME", "runns_orders"))
	db.AutoMigrate(&models.Order{})
	
	rmqURL := utils.GetEnv("RABBITMQ_URL", "amqp://ryder:ryder_pass@localhost:5672/")
	rmqClient, err := rabbitmq.Connect(rmqURL)
	if err != nil {
		log.Printf("[order-service] warning: rabbitmq connection failed: %v", err)
	} else {
		log.Printf("[order-service] connected to rabbitmq")
		defer rmqClient.Close()
		// Setup exchange
		_ = rmqClient.SetupExchangeQueue("rydex_events", "topic", "order_events_queue", "order.*")
	}

	svc := service.NewOrderService(db, rmqClient)
	h := handler.NewOrderHandler(svc)
	port := utils.GetEnv("ORDER_SERVICE_PORT", "8082")
	log.Printf("[order-service] :%s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler.NewRouter(h)); err != nil {
		log.Fatal(err)
	}
}
