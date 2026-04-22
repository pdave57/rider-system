package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/rydex/payment-service/handler"
	"github.com/rydex/payment-service/service"
	"github.com/rydex/shared/models"
	"github.com/rydex/shared/utils"
)

func main() {
	_ = godotenv.Load("../../deploy/env/.env")
	db := utils.NewPostgres(utils.GetEnv("DB_NAME", "rydex_payments"))
	db.AutoMigrate(&models.Payment{}, &models.ClientProfile{})
	svc := service.NewPaymentService(db)
	h := handler.NewPaymentHandler(svc)
	port := utils.GetEnv("PAYMENT_SERVICE_PORT", "8084")
	log.Printf("[payment-service] :%s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler.NewRouter(h)); err != nil {
		log.Fatal(err)
	}
}
