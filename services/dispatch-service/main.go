package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/rabbitmq/amqp091-go"
	"github.com/runns/dispatch-service/dto"
	"github.com/runns/dispatch-service/handler"
	"github.com/runns/dispatch-service/service"
	"github.com/runns/shared/models"
	"github.com/runns/shared/rabbitmq"
	"github.com/runns/shared/utils"
)

func main() {
	_ = godotenv.Load("../../deploy/env/.env")
	db := utils.NewPostgres(utils.GetEnv("DB_NAME", "runns_dispatch"))
	db.AutoMigrate(&models.Dispatch{}, &models.RiderProfile{})
	rdb := utils.NewRedis()
	
	svc := service.NewDispatchService(db, rdb)

	rmqClient, err := rabbitmq.NewClient()
	if err != nil {
		log.Printf("[dispatch-service] warning: rabbitmq connection failed: %v", err)
	} else {
		log.Printf("[dispatch-service] connected to rabbitmq")
		defer rmqClient.Close()
		_ = rmqClient.SetupExchangeQueue("rydex_events", "topic", "dispatch_events_queue", "order.*")
		_ = rmqClient.Channel.QueueBind("dispatch_events_queue", "payment.*", "rydex_events", false, nil)
		
		err := rmqClient.Consume("dispatch_events_queue", func(d amqp091.Delivery) {
			log.Printf("[dispatch-service] Received event %s: %s", d.RoutingKey, string(d.Body))
			
			if d.RoutingKey == "payment.successful" {
				var event struct {
					OrderID uint `json:"order_id"`
				}
				if err := json.Unmarshal(d.Body, &event); err == nil && event.OrderID != 0 {
					_, assignErr := svc.Assign(dto.AssignRequest{OrderID: event.OrderID})
					if assignErr != nil {
						log.Printf("[dispatch-service] Failed to auto-assign for order %d: %v", event.OrderID, assignErr)
					} else {
						log.Printf("[dispatch-service] Successfully auto-assigned rider for order %d", event.OrderID)
					}
				}
			}
		})
		if err != nil {
			log.Printf("[dispatch-service] warning: failed to start consumer: %v", err)
		}
	}

	h := handler.NewDispatchHandler(svc)
	port := utils.GetEnv("DISPATCH_SERVICE_PORT", "8083")
	log.Printf("[dispatch-service] :%s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler.NewRouter(h)); err != nil {
		log.Fatal(err)
	}
}
