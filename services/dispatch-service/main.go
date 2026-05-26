package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/runns/dispatch-service/dto"
	"github.com/runns/dispatch-service/handler"
	"github.com/runns/dispatch-service/service"
	"github.com/runns/shared/models"
	"github.com/runns/shared/rabbitmq"
	"github.com/runns/shared/utils"
)

func main() {
	utils.LoadEnvFile(".env", "../../deploy/env/.env")
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

		err := rmqClient.Consume("dispatch_events_queue", "rydex_events", "payment.*", func(body []byte) error {
			log.Printf("[dispatch-service] Received event: %s", string(body))

			var event struct {
				OrderID uint `json:"order_id"`
			}
			if err := json.Unmarshal(body, &event); err != nil {
				return err
			}
			if event.OrderID == 0 {
				return nil
			}

			_, assignErr := svc.Assign(dto.AssignRequest{OrderID: event.OrderID})
			if assignErr != nil {
				log.Printf("[dispatch-service] Failed to auto-assign for order %d: %v", event.OrderID, assignErr)
				return assignErr
			}
			log.Printf("[dispatch-service] Successfully auto-assigned rider for order %d", event.OrderID)
			return nil
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
