package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/runns/shopforme-service/handler"
	"github.com/runns/shopforme-service/service"
	"github.com/runns/shared/models"
	"github.com/runns/shared/rabbitmq"
	"github.com/runns/shared/utils"
)

func main() {
	utils.LoadEnvFile(".env", "../../deploy/env/.env")
	db := utils.NewPostgres(utils.GetEnv("DB_NAME", "runns_shopforme"))
	db.AutoMigrate(&models.ShopForMeRequest{}, &models.ShopForMeItem{}, &models.ShopForMeOrder{})

	ordersDB := utils.NewPostgres(utils.GetEnv("ORDERS_DB_NAME", "runns_orders"))
	ridersDB := utils.NewPostgres(utils.GetEnv("RIDERS_DB_NAME", "runns_riders"))
	paymentsDB := utils.NewPostgres(utils.GetEnv("PAYMENTS_DB_NAME", "runns_payments"))

	rmqClient, err := rabbitmq.NewClient()
	if err != nil {
		log.Printf("[shopforme-service] warning: rabbitmq connection failed: %v", err)
	} else {
		log.Printf("[shopforme-service] connected to rabbitmq")
		defer rmqClient.Close()
		_ = rmqClient.SetupExchangeQueue("rydex_events", "topic", "shopforme_events_queue", "shopforme.*")
		_ = rmqClient.SetupExchangeQueue("rydex_events", "topic", "shopforme_events_queue", "order.*")
		_ = rmqClient.SetupExchangeQueue("rydex_events", "topic", "shopforme_events_queue", "rider.*")
	}

	svc := service.NewShopForMeService(db, ordersDB, ridersDB, paymentsDB, rmqClient)
	h := handler.NewShopForMeHandler(svc)
	port := utils.GetEnv("SHOPFORME_SERVICE_PORT", "8087")
	log.Printf("[shopforme-service] :%s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler.NewRouter(h)); err != nil {
		log.Fatal(err)
	}
}