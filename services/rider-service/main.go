package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/rydex/rider-service/handler"
	"github.com/rydex/rider-service/service"
	"github.com/rydex/shared/models"
	"github.com/rydex/shared/utils"
)

func main() {
	_ = godotenv.Load("../../deploy/env/.env")
	db := utils.NewPostgres(utils.GetEnv("DB_NAME", "rydex_riders"))
	db.AutoMigrate(&models.RiderProfile{})
	svc := service.NewRiderService(db)
	h := handler.NewRiderHandler(svc)
	port := utils.GetEnv("RIDER_SERVICE_PORT", "8086")
	log.Printf("[rider-service] :%s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler.NewRouter(h)); err != nil {
		log.Fatal(err)
	}
}
