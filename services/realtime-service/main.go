package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/runns/realtime-service/handler"
	"github.com/runns/realtime-service/hub"
	"github.com/runns/shared/rabbitmq"
	"github.com/runns/shared/utils"
)

func main() {
	utils.LoadEnvFile(".env", "../../deploy/env/.env")
	rdb := utils.NewRedis()

	rmqClient, err := rabbitmq.NewClient()
	if err != nil {
		log.Printf("[realtime-service] warning: rabbitmq connection failed: %v", err)
	} else {
		log.Printf("[realtime-service] connected to rabbitmq")
		defer rmqClient.Close()
		_ = rmqClient.SetupExchangeQueue("rydex_events", "topic", "realtime_events_queue", "order.*")
	}

	h := handler.NewRealtimeHandler(hub.NewHub(), rdb)
	h.StartRedisSub(context.Background())
	port := utils.GetEnv("REALTIME_SERVICE_PORT", "8085")
	log.Printf("[realtime-service] :%s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler.NewRouter(h)); err != nil {
		log.Fatal(err)
	}
}
