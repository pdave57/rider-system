package service

import (
	"context"
	"time"

	"github.com/runns/shared/models"
	"github.com/runns/shared/rabbitmq"
)

type EventPublisher interface {
	PublishOrderCreated(ctx context.Context, order *models.Order) error
	PublishOrderStatusChanged(ctx context.Context, order *models.Order, oldStatus string, updatedBy uint) error
}

type eventPublisher struct {
	rabbitClient *rabbitmq.Client
}

func NewEventPublisher(client *rabbitmq.Client) EventPublisher {
	return &eventPublisher{
		rabbitClient: client,
	}
}

func (p *eventPublisher) PublishOrderCreated(ctx context.Context, order *models.Order) error {
	if p.rabbitClient == nil {
		return nil
	}

	event := rabbitmq.OrderCreatedEvent{
		OrderID:      order.ID,
		TrackingCode: order.TrackingCode,
		ClientID:     order.ClientID,
		PickupLat:    order.PickupLatitude,
		PickupLng:    order.PickupLongitude,
		DropoffLat:   order.DropoffLatitude,
		DropoffLng:   order.DropoffLongitude,
		DeliveryFee:  order.DeliveryFee,
		WeightKg:     order.WeightKg,
		Timestamp:    time.Now(),
	}

	return p.rabbitClient.PublishJSON(ctx, rabbitmq.ExchangeOrderEvents, rabbitmq.RoutingKeyOrderCreated, event)
}

func (p *eventPublisher) PublishOrderStatusChanged(ctx context.Context, order *models.Order, oldStatus string, updatedBy uint) error {
	if p.rabbitClient == nil {
		return nil
	}

	event := rabbitmq.OrderStatusChangedEvent{
		OrderID:      order.ID,
		TrackingCode: order.TrackingCode,
		OldStatus:    oldStatus,
		NewStatus:    string(order.Status),
		RiderID:      order.RiderID,
		UpdatedBy:    updatedBy,
		Timestamp:    time.Now(),
	}

	return p.rabbitClient.PublishJSON(ctx, rabbitmq.ExchangeOrderEvents, rabbitmq.RoutingKeyOrderStatusChanged, event)
}
