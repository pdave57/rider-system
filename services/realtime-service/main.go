package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/rydex/realtime-service/handler"
	"github.com/rydex/realtime-service/hub"
	"github.com/rydex/shared/utils"
)

func main() {
	_ = godotenv.Load("../../deploy/env/.env")
	rdb := utils.NewRedis()
	h := handler.NewRealtimeHandler(hub.NewHub(), rdb)
	h.StartRedisSub(context.Background())
	port := utils.GetEnv("REALTIME_SERVICE_PORT", "8085")
	log.Printf("[realtime-service] :%s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler.NewRouter(h)); err != nil {
		log.Fatal(err)
	}
}
