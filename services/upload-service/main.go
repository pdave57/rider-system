package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/runns/shared/models"
	"github.com/runns/shared/utils"
	"github.com/runns/upload-service/handler"
	"github.com/runns/upload-service/service"
)

func main() {
	utils.LoadEnvFile(".env", "../../deploy/env/.env")

	// Connect to the auth DB — that's where users and avatar fields live
	db := utils.NewPostgres(utils.GetEnv("DB_NAME", "rydex_auth"))

	// AutoMigrate will add avatar_public_id column if it doesn't exist
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("[upload-service] migration failed: %v", err)
	}

	svc, err := service.NewCloudinaryService(db)
	if err != nil {
		log.Fatalf("[upload-service] %v", err)
	}

	h := handler.NewUploadHandler(svc)
	router := handler.NewRouter(h)

	port := utils.GetEnv("UPLOAD_SERVICE_PORT", "8087")
	log.Printf("[upload-service] :%s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), router); err != nil {
		log.Fatal(err)
	}
}
