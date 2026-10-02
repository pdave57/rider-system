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

	// Dispatch DB — stores Dispatch records only
	db := utils.NewPostgres(utils.GetEnv("DB_NAME", "runns_dispatch"))
	db.Config.DisableForeignKeyConstraintWhenMigrating = true
	db.AutoMigrate(&models.Dispatch{})

	// Orders DB — needed to look up and update Order records
	ordersDB := utils.NewPostgres("runns_orders")

	// Riders DB — needed to check and update RiderProfile availability
	ridersDB := utils.NewPostgres("runns_riders")

	rdb := utils.NewRedis()

	svc := service.NewDispatchService(db, ordersDB, ridersDB, rdb)

	rmqClient, err := rabbitmq.NewClient()
	if err != nil {
		log.Printf("[dispatch-service] warning: rabbitmq connection failed: %v", err)
	} else {
		log.Printf("[dispatch-service] connected to rabbitmq")
		defer rmqClient.Close()

		// Listen to order.created events from order-service
		err := rmqClient.Consume("dispatch_order_events_queue", rabbitmq.ExchangeOrderEvents, rabbitmq.RoutingKeyOrderCreated, func(body []byte) error {
			log.Printf("[dispatch-service] Received order.created event: %s", string(body))

			var event struct {
				OrderID uint `json:"order_id"`
			}
			if err := json.Unmarshal(body, &event); err != nil {
				log.Printf("[dispatch-service] Failed to unmarshal event: %v", err)
				return err
			}
			if event.OrderID == 0 {
				log.Printf("[dispatch-service] Received event with no order_id, skipping")
				return nil
			}

			log.Printf("[dispatch-service] Auto-assigning order %d", event.OrderID)
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

		// Listen to shopforme.order_placed events from shopforme-service
		err = rmqClient.Consume("dispatch_shopforme_events_queue", "rydex_events", rabbitmq.RoutingKeyShopForMeOrderPlaced, func(body []byte) error {
			log.Printf("[dispatch-service] Received shopforme.order_placed event: %s", string(body))

			var event rabbitmq.ShopForMeOrderPlacedEvent
			if err := json.Unmarshal(body, &event); err != nil {
				log.Printf("[dispatch-service] Failed to unmarshal shopforme event: %v", err)
				return err
			}
			if event.RequestID == 0 {
				log.Printf("[dispatch-service] Received shopforme event with no request_id, skipping")
				return nil
			}

			log.Printf("[dispatch-service] Shopforme request %d placed for client %d with total amount %.2f", event.RequestID, event.ClientID, event.TotalAmount)
			// Here you could add logic to create a dispatch record for shopforme orders if needed
			return nil
		})
		if err != nil {
			log.Printf("[dispatch-service] warning: failed to start shopforme consumer: %v", err)
		}
	}

	h := handler.NewDispatchHandler(svc)
	port := utils.GetEnv("DISPATCH_SERVICE_PORT", "8083")
	log.Printf("[dispatch-service] :%s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler.NewRouter(h)); err != nil {
		log.Fatal(err)
	}
}
