package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/runns/realtime-service/handler"
	"github.com/runns/realtime-service/hub"
	"github.com/runns/shared/rabbitmq"
	"github.com/runns/shared/utils"
)

func main() {
	_ = godotenv.Load("../../deploy/env/.env")
	rdb := utils.NewRedis()

	rmqURL := utils.GetEnv("RABBITMQ_URL", "amqp://ryder:ryder_pass@localhost:5672/")
	rmqClient, err := rabbitmq.Connect(rmqURL)
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
