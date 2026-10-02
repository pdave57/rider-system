package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/runns/payment-service/handler"
	"github.com/runns/payment-service/service"
	"github.com/runns/shared/rabbitmq"
	"github.com/runns/shared/utils"
)

func main() {
	utils.LoadEnvFile(".env", "../../deploy/env/.env")
	db := utils.NewPostgres(utils.GetEnv("DB_NAME", "runns_payments"))

	// Create tables manually to avoid foreign key issues with cross-database references
	createTablesSQL := []string{
		`CREATE TABLE IF NOT EXISTS client_profiles (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT UNIQUE NOT NULL,
			default_address TEXT,
			latitude DECIMAL(10,8),
			longitude DECIMAL(11,8),
			wallet_balance DECIMAL(12,2) DEFAULT 0.00,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,
		`CREATE TABLE IF NOT EXISTS rider_wallets (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT UNIQUE NOT NULL,
			balance DECIMAL(12,2) DEFAULT 0.00,
			pending_balance DECIMAL(12,2) DEFAULT 0.00,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,
		`CREATE TABLE IF NOT EXISTS rider_profiles (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT UNIQUE NOT NULL,
			vehicle_type VARCHAR(20),
			vehicle_plate VARCHAR(20),
			license_number VARCHAR(50),
			nin VARCHAR(20),
			nin_verified BOOLEAN DEFAULT FALSE,
			nin_verified_at TIMESTAMPTZ,
			is_available BOOLEAN DEFAULT FALSE,
			is_verified BOOLEAN DEFAULT FALSE,
			current_latitude DECIMAL(10,8),
			current_longitude DECIMAL(11,8),
			rating DECIMAL(3,2) DEFAULT 0.00,
			total_deliveries INT DEFAULT 0,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,
		`CREATE TABLE IF NOT EXISTS payments (
			id BIGSERIAL PRIMARY KEY,
			order_id BIGINT NOT NULL,
			client_id BIGINT NOT NULL,
			amount DECIMAL(12,2) NOT NULL,
			method VARCHAR(20) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			reference VARCHAR(100) UNIQUE,
			gateway_ref VARCHAR(200),
			failure_reason VARCHAR(255),
			paid_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,
		`CREATE INDEX IF NOT EXISTS idx_client_profiles_user_id ON client_profiles(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_rider_wallets_user_id ON rider_wallets(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_rider_profiles_user_id ON rider_profiles(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id)`,
		`CREATE INDEX IF NOT EXISTS idx_payments_client_id ON payments(client_id)`,
		`CREATE INDEX IF NOT EXISTS idx_payments_reference ON payments(reference)`,
	}

	for _, sql := range createTablesSQL {
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("[payment-service] warning: failed to create table: %v", err)
		}
	}

	rmqClient, err := rabbitmq.NewClient()
	if err != nil {
		log.Printf("[payment-service] warning: rabbitmq connection failed: %v", err)
	} else {
		log.Printf("[payment-service] connected to rabbitmq")
		defer rmqClient.Close()
		_ = rmqClient.SetupExchangeQueue("rydex_events", "topic", "payment_events_queue", "order.*")
		_ = rmqClient.SetupExchangeQueue("rydex_events", "topic", "payment_events_queue", "wallet.*")
		_ = rmqClient.SetupExchangeQueue("rydex_events", "topic", "payment_events_queue", "shopforme.*")
		_ = rmqClient.SetupExchangeQueue("rydex_events", "topic", "payment_events_queue", "rider.*")
		_ = rmqClient.SetupExchangeQueue("rydex_events", "topic", "payment_events_queue", "user.*")
	}

	svc := service.NewPaymentService(db, rmqClient)
	
	// Consume user registered events
	if rmqClient != nil {
		go func() {
			_ = rmqClient.Consume("payment_events_queue", "rydex_events", "user.registered", func(body []byte) error {
				var event rabbitmq.UserRegisteredEvent
				if err := json.Unmarshal(body, &event); err != nil {
					return err
				}
				return svc.SyncUser(event)
			})
		}()
	}

	h := handler.NewPaymentHandler(svc)
	port := utils.GetEnv("PAYMENT_SERVICE_PORT", "8084")
	log.Printf("[payment-service] :%s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler.NewRouter(h)); err != nil {
		log.Fatal(err)
	}
}
