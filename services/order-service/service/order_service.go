package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/runns/order-service/dto"
	"github.com/runns/shared/models"
	"github.com/runns/shared/rabbitmq"
	"github.com/runns/shared/utils"
	"gorm.io/gorm"
)

type OrderService struct {
	db         *gorm.DB
	rmqManager *rabbitmq.Client
}

func NewOrderService(db *gorm.DB, rmqManager *rabbitmq.Client) *OrderService {
	return &OrderService{db: db, rmqManager: rmqManager}
}

func (s *OrderService) SyncUser(event rabbitmq.UserRegisteredEvent) error {
	user := models.User{
		ID:       event.Data.UserID,
		FullName: event.Data.FullName,
		Email:    event.Data.Email,
		Phone:    event.Data.Phone,
		Role:     models.Role(event.Data.Role),
		IsActive: true,
	}

	var existing models.User
	if err := s.db.First(&existing, event.Data.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return s.db.Create(&user).Error
		}
		return err
	}

	return s.db.Model(&existing).Updates(map[string]interface{}{
		"full_name": user.FullName,
		"email":     user.Email,
		"phone":     user.Phone,
		"role":      user.Role,
		"is_active": user.IsActive,
	}).Error
}

func (s *OrderService) Create(clientID uint, req dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	if req.PickupAddress == "" || req.DropoffAddress == "" {
		return nil, errors.New("pickup_address and dropoff_address are required")
	}
	dist := utils.HaversineKm(req.PickupLatitude, req.PickupLongitude, req.DropoffLatitude, req.DropoffLongitude)
	fee := utils.DeliveryFee(dist, req.WeightKg)
	order := &models.Order{
		TrackingCode:      utils.GenerateTrackingCode(),
		ClientID:          clientID,
		Status:            models.OrderPending,
		PickupAddress:     req.PickupAddress,
		PickupLatitude:    req.PickupLatitude,
		PickupLongitude:   req.PickupLongitude,
		DropoffAddress:    req.DropoffAddress,
		DropoffLatitude:   req.DropoffLatitude,
		DropoffLongitude:  req.DropoffLongitude,
		PackageDesc:       req.PackageDesc,
		WeightKg:          req.WeightKg,
		DeliveryFee:       fee,
		Notes:             req.Notes,
	}
	if err := s.db.Create(order).Error; err != nil {
		return nil, err
	}
	resp := dto.ToOrderResponse(order)

	// Publish order.created event
	go func() {
		event := rabbitmq.OrderCreatedEvent{
			OrderID:          order.ID,
			TrackingCode:     order.TrackingCode,
			ClientID:         order.ClientID,
			PickupLat:        order.PickupLatitude,
			PickupLng:        order.PickupLongitude,
			DropoffLat:       order.DropoffLatitude,
			DropoffLng:       order.DropoffLongitude,
			DeliveryFee:      order.DeliveryFee,
			WeightKg:         order.WeightKg,
			Timestamp:        order.CreatedAt,
		}
		if err := s.publishEvent(rabbitmq.ExchangeOrderEvents, rabbitmq.RoutingKeyOrderCreated, event); err != nil {
			log.Printf("Failed to publish order.created event: %v", err)
		}
	}()

	return &resp, nil
}

func (s *OrderService) GetByID(id uint) (*dto.OrderResponse, error) {
	var o models.Order
	if err := s.db.First(&o, id).Error; err != nil {
		return nil, errors.New("order not found")
	}
	resp := dto.ToOrderResponse(&o)
	return &resp, nil
}

func (s *OrderService) GetByTracking(code string) (*dto.OrderResponse, error) {
	var o models.Order
	if err := s.db.Where("tracking_code = ?", code).First(&o).Error; err != nil {
		return nil, errors.New("order not found")
	}
	resp := dto.ToOrderResponse(&o)
	return &resp, nil
}

func (s *OrderService) ListByClient(clientID uint) ([]dto.OrderResponse, error) {
	var orders []models.Order
	s.db.Where("client_id = ?", clientID).Order("created_at desc").Find(&orders)
	return toResponses(orders), nil
}

func (s *OrderService) ListByRider(riderID uint) ([]dto.OrderResponse, error) {
	var orders []models.Order
	s.db.Where("rider_id = ?", riderID).Order("created_at desc").Find(&orders)
	return toResponses(orders), nil
}

func (s *OrderService) ListPending() ([]dto.OrderResponse, error) {
	var orders []models.Order
	s.db.Where("status = ?", models.OrderPending).Order("created_at asc").Find(&orders)
	return toResponses(orders), nil
}

func (s *OrderService) UpdateStatus(orderID uint, req dto.UpdateStatusRequest, userID uint) (*dto.OrderResponse, error) {
	var o models.Order
	if err := s.db.First(&o, orderID).Error; err != nil {
		return nil, errors.New("order not found")
	}
	if !validTransition(o.Status, req.Status) {
		return nil, errors.New("invalid status transition")
	}

	oldStatus := string(o.Status)
	o.Status = req.Status
	if err := s.db.Save(&o).Error; err != nil {
		return nil, err
	}

	resp := dto.ToOrderResponse(&o)

	// Publish order.status_changed event
	go func() {
		event := rabbitmq.OrderStatusChangedEvent{
			OrderID:      o.ID,
			TrackingCode: o.TrackingCode,
			OldStatus:    oldStatus,
			NewStatus:    string(o.Status),
			RiderID:      o.RiderID,
			UpdatedBy:    userID,
			Timestamp:    o.UpdatedAt,
		}
		if err := s.publishEvent(rabbitmq.ExchangeOrderEvents, rabbitmq.RoutingKeyOrderStatusChanged, event); err != nil {
			log.Printf("Failed to publish order.status_changed event: %v", err)
		}
	}()

	return &resp, nil
}

func (s *OrderService) Cancel(orderID, clientID uint) error {
	var o models.Order
	if err := s.db.First(&o, orderID).Error; err != nil {
		return errors.New("order not found")
	}
	if o.ClientID != clientID {
		return errors.New("not your order")
	}
	if o.Status != models.OrderPending {
		return errors.New("only pending orders can be cancelled")
	}

	if err := s.db.Model(&o).Update("status", models.OrderCancelled).Error; err != nil {
		return err
	}

	// Publish order.cancelled event
	go func() {
		event := rabbitmq.OrderStatusChangedEvent{
			OrderID:      o.ID,
			TrackingCode: o.TrackingCode,
			OldStatus:    string(models.OrderPending),
			NewStatus:    string(models.OrderCancelled),
			RiderID:      o.RiderID,
			UpdatedBy:    clientID,
			Timestamp:    o.UpdatedAt,
		}
		if err := s.publishEvent(rabbitmq.ExchangeOrderEvents, rabbitmq.RoutingKeyOrderStatusChanged, event); err != nil {
			log.Printf("Failed to publish order.status_changed event: %v", err)
		}
	}()

	return nil
}

func (s *OrderService) publishEvent(exchange, routingKey string, event interface{}) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return s.rmqManager.Publish(context.Background(), exchange, routingKey, body)
}

func toResponses(orders []models.Order) []dto.OrderResponse {
	result := make([]dto.OrderResponse, len(orders))
	for i := range orders {
		result[i] = dto.ToOrderResponse(&orders[i])
	}
	return result
}

func validTransition(from, to models.OrderStatus) bool {
	allowed := map[models.OrderStatus][]models.OrderStatus{
		models.OrderPending:   {models.OrderAssigned, models.OrderCancelled},
		models.OrderAssigned:  {models.OrderPickedUp, models.OrderCancelled},
		models.OrderPickedUp:  {models.OrderDelivered},
		models.OrderDelivered: {},
		models.OrderCancelled: {},
	}
	for _, s := range allowed[from] {
		if s == to {
			return true
		}
	}
	return false
}