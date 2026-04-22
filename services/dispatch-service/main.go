package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/rydex/dispatch-service/handler"
	"github.com/rydex/dispatch-service/service"
	"github.com/rydex/shared/models"
	"github.com/rydex/shared/utils"
)

func main() {
	_ = godotenv.Load("../../deploy/env/.env")
	db := utils.NewPostgres(utils.GetEnv("DB_NAME", "rydex_dispatch"))
	db.AutoMigrate(&models.Dispatch{}, &models.RiderProfile{})
	rdb := utils.NewRedis()
	svc := service.NewDispatchService(db, rdb)
	h := handler.NewDispatchHandler(svc)
	port := utils.GetEnv("DISPATCH_SERVICE_PORT", "8083")
	log.Printf("[dispatch-service] :%s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler.NewRouter(h)); err != nil {
		log.Fatal(err)
	}
}
