package service

import (
	"context"
	"time"

	"github.com/runns/auth-service/dto"
	"github.com/google/uuid"
	events "github.com/runns/shared/rabbitmq"
)

type EventPublisher interface {
    PublishUserRegistered(ctx context.Context, userID uint, req interface{}) error
    PublishUserRoleChanged(ctx context.Context, userID uint, oldRole, newRole string, changedBy uint) error
    PublishUserDeactivated(ctx context.Context, userID uint, reason string, deactivatedBy uint) error
}

type eventPublisher struct {
    rabbitClient *events.Client
}

func NewEventPublisher(client *events.Client) EventPublisher {
    return &eventPublisher{
        rabbitClient: client,
    }
}

func (p *eventPublisher) PublishUserRegistered(ctx context.Context, userID uint, req interface{}) error {
    // Type assert to get the actual request
    registerReq, ok := req.(dto.RegisterRequest)
    if !ok {
        // Handle if needed
        return nil
    }

    eventData := events.UserRegisteredData{
        UserID:       userID,
        Email:        registerReq.Email,
        Phone:        registerReq.Phone,
        FullName:     registerReq.FullName,
        Role:         string(registerReq.Role),
        RegisteredAt: time.Now(),
    }

    // Add role-specific data
    switch registerReq.Role {
    case "rider":
        eventData.VehicleType = string(registerReq.VehicleType)
        eventData.VehiclePlate = registerReq.VehiclePlate
        eventData.LicenseNumber = registerReq.LicenseNumber
    case "client":
        eventData.DefaultAddress = registerReq.DefaultAddress
        eventData.Latitude = registerReq.Latitude
        eventData.Longitude = registerReq.Longitude
    }

    event := events.UserRegisteredEvent{
        BaseEvent: events.BaseEvent{
            EventID:   uuid.New().String(),
            EventType: events.UserRegistered,
            Timestamp: time.Now(),
            Service:   "auth-service",
            Version:   "1.0",
        },
        Data: eventData,
    }

    // Choose routing key based on role
    routingKey := events.UserRegistered
    switch registerReq.Role {
    case "rider":
        routingKey = "user.registered.rider"
    case "client":
        routingKey = "user.registered.client"
    }

    return p.rabbitClient.PublishEvent(ctx, routingKey, event)
}

func (p *eventPublisher) PublishUserRoleChanged(ctx context.Context, userID uint, oldRole, newRole string, changedBy uint) error {
    event := events.UserRoleChangedEvent{
        BaseEvent: events.BaseEvent{
            EventID:   uuid.New().String(),
            EventType: events.UserRoleChanged,
            Timestamp: time.Now(),
            Service:   "auth-service",
            Version:   "1.0",
        },
        Data: events.UserRoleChangedData{
            UserID:    userID,
            OldRole:   oldRole,
            NewRole:   newRole,
            ChangedBy: changedBy,
            ChangedAt: time.Now(),
        },
    }

    routingKey := events.UserRoleChanged
    if newRole == "rider" {
        routingKey = "user.role_changed.to_rider"
    }

    return p.rabbitClient.PublishEvent(ctx, routingKey, event)
}

func (p *eventPublisher) PublishUserDeactivated(ctx context.Context, userID uint, reason string, deactivatedBy uint) error {
    event := events.UserDeactivatedEvent{
        BaseEvent: events.BaseEvent{
            EventID:   uuid.New().String(),
            EventType: events.UserDeactivated,
            Timestamp: time.Now(),
            Service:   "auth-service",
            Version:   "1.0",
        },
        Data: events.UserDeactivatedData{
            UserID:        userID,
            Reason:        reason,
            DeactivatedBy: deactivatedBy,
            DeactivatedAt: time.Now(),
        },
    }

    return p.rabbitClient.PublishEvent(ctx, events.UserDeactivated, event)
}