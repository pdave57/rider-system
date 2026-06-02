package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/runns/rider-service/handler"
	"github.com/runns/rider-service/service"
	"github.com/runns/shared/models"
	"github.com/runns/shared/rabbitmq"
	"github.com/runns/shared/utils"
)

func main() {
	utils.LoadEnvFile(".env", "../../deploy/env/.env")

	// Riders DB — stores RiderProfile records
	db := utils.NewPostgres(utils.GetEnv("DB_NAME", "runns_riders"))
	db.AutoMigrate(&models.RiderProfile{})

	// Auth DB — needed to fetch User records for profiles
	authDB := utils.NewPostgres("runns_auth")

	// Redis — for geo index used by dispatch service
	rdb := utils.NewRedis()

	rmqClient, err := rabbitmq.NewClient()
	if err != nil {
		log.Printf("[rider-service] warning: rabbitmq connection failed: %v", err)
	} else {
		log.Printf("[rider-service] connected to rabbitmq")
		defer rmqClient.Close()
		_ = rmqClient.SetupExchangeQueue("rydex_events", "topic", "rider_events_queue", "rider.*")
	}

	svc := service.NewRiderService(db, authDB, rdb)
	h := handler.NewRiderHandler(svc)
	port := utils.GetEnv("RIDER_SERVICE_PORT", "8086")
	log.Printf("[rider-service] :%s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler.NewRouter(h)); err != nil {
		log.Fatal(err)
	}
}
