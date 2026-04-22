package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/rydex/order-service/handler"
	"github.com/rydex/order-service/service"
	"github.com/rydex/shared/models"
	"github.com/rydex/shared/utils"
)

func main() {
	_ = godotenv.Load("../../deploy/env/.env")
	db := utils.NewPostgres(utils.GetEnv("DB_NAME", "rydex_orders"))
	db.AutoMigrate(&models.Order{})
	svc := service.NewOrderService(db)
	h := handler.NewOrderHandler(svc)
	port := utils.GetEnv("ORDER_SERVICE_PORT", "8082")
	log.Printf("[order-service] :%s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler.NewRouter(h)); err != nil {
		log.Fatal(err)
	}
}
