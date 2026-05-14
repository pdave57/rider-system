package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/runns/auth-service/handler"
	"github.com/runns/auth-service/service"
	"github.com/runns/shared/models"
	"github.com/runns/shared/utils"
)

func main() {
	_ = godotenv.Load("../../deploy/env/.env")
	db := utils.NewPostgres(utils.GetEnv("DB_NAME", "runns_auth"))
	db.AutoMigrate(&models.User{}, &models.RiderProfile{}, &models.ClientProfile{})	
	svc := service.NewAuthService(db)
	h := handler.NewAuthHandler(svc)
	port := utils.GetEnv("AUTH_SERVICE_PORT", "8081")
	log.Printf("[auth-service] :%s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler.NewRouter(h)); err != nil {
		log.Fatal(err)
	}
}
