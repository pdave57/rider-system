package consumer

import (
	"encoding/json"
	"log"

	"github.com/runns/order-service/service"
	"github.com/runns/shared/rabbitmq"
)

type AuthEventConsumer struct {
	rmqManager   *rabbitmq.Client
	orderService *service.OrderService
}

func NewAuthEventConsumer(rmqManager *rabbitmq.Client, orderService *service.OrderService) *AuthEventConsumer {
	return &AuthEventConsumer{
		rmqManager:   rmqManager,
		orderService: orderService,
	}
}

func (c *AuthEventConsumer) Start() error {
	// IMPORTANT: Each binding must use its OWN unique queue name.
	// If all bindings share one queue, RabbitMQ round-robins messages
	// across all consumer goroutines — so the wrong handler processes
	// the wrong event type.
	bindings := []struct {
		queue      string
		routingKey string
		handler    func([]byte) error
	}{
		// Specific client/rider registration events (published by auth-service)
		{"order_auth_user_registered_client", "user.registered.client", c.handleUserRegistrationBytes},
		{"order_auth_user_registered_rider", "user.registered.rider", c.handleUserRegistrationBytes},
		// Generic registration fallback (routing key "user.registered")
		{"order_auth_user_registered", rabbitmq.RoutingKeyUserRegistered, c.handleUserRegistrationBytes},
		// Role changes and deactivations
		{"order_auth_role_changed", rabbitmq.RoutingKeyUserRoleChanged, c.handleUserRoleChangedBytes},
		{"order_auth_user_deactivated", rabbitmq.RoutingKeyUserDeactivated, c.handleUserDeactivatedBytes},
	}

	for _, b := range bindings {
		if err := c.rmqManager.Consume(b.queue, rabbitmq.ExchangeAuthEvents, b.routingKey, b.handler); err != nil {
			return err
		}
	}

	return nil
}

func (c *AuthEventConsumer) handleUserRegistrationBytes(body []byte) error {
	var event rabbitmq.UserRegisteredEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Failed to unmarshal user.registered: %v", err)
		return err
	}
	log.Printf("User registered: UserID=%d, Role=%s, Email=%s", event.Data.UserID, event.Data.Role, event.Data.Email)
	return c.handleUserRegistration(event)
}

func (c *AuthEventConsumer) handleUserRoleChangedBytes(body []byte) error {
	var event rabbitmq.UserRoleChangedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Failed to unmarshal user.role_changed: %v", err)
		return err
	}
	log.Printf("User role changed: UserID=%d, NewRole=%s", event.Data.UserID, event.Data.NewRole)
	return c.handleUserRoleChanged(event)
}

func (c *AuthEventConsumer) handleUserDeactivatedBytes(body []byte) error {
	var event rabbitmq.UserDeactivatedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Failed to unmarshal user.deactivated: %v", err)
		return err
	}
	log.Printf("User deactivated: UserID=%d", event.Data.UserID)
	return c.handleUserDeactivated(event)
}

func (c *AuthEventConsumer) handleUserRegistration(event rabbitmq.UserRegisteredEvent) error {
	log.Printf("Initializing data for user %d in order-service", event.Data.UserID)
	return c.orderService.SyncUser(event)
}

func (c *AuthEventConsumer) handleUserRoleChanged(event rabbitmq.UserRoleChangedEvent) error {
	log.Printf("Updating user role for user %d in order-service", event.Data.UserID)
	return nil
}

func (c *AuthEventConsumer) handleUserDeactivated(event rabbitmq.UserDeactivatedEvent) error {
	log.Printf("Handling user deactivation for user %d in order-service", event.Data.UserID)
	return nil
}