package service

import (
	"errors"

	"github.com/rydex/order-service/dto"
	"github.com/rydex/shared/models"
	"github.com/rydex/shared/utils"
	"gorm.io/gorm"
)

type OrderService struct{ db *gorm.DB }

func NewOrderService(db *gorm.DB) *OrderService { return &OrderService{db: db} }

func (s *OrderService) Create(clientID uint, req dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	if req.PickupAddress == "" || req.DropoffAddress == "" {
		return nil, errors.New("pickup_address and dropoff_address are required")
	}
	dist := utils.HaversineKm(req.PickupLatitude, req.PickupLongitude, req.DropoffLatitude, req.DropoffLongitude)
	fee := utils.DeliveryFee(dist, req.WeightKg)
	order := &models.Order{
		TrackingCode: utils.GenerateTrackingCode(), 
		ClientID: clientID,
		Status: models.OrderPending, 
		PickupAddress: req.PickupAddress,
		PickupLatitude: req.PickupLatitude, 
		PickupLongitude: req.PickupLongitude,
		DropoffAddress: req.DropoffAddress, 
		DropoffLatitude: req.DropoffLatitude,
		DropoffLongitude: req.DropoffLongitude, 
		PackageDesc: req.PackageDesc,
		WeightKg: req.WeightKg, 
		DeliveryFee: fee, 
		Notes: req.Notes,
	}
	if err := s.db.Create(order).Error; err != nil { return nil, err }
	resp := dto.ToOrderResponse(order)
	return &resp, nil
}

func (s *OrderService) GetByID(id uint) (*dto.OrderResponse, error) {
	var o models.Order
	if err := s.db.First(&o, id).Error; err != nil { return nil, errors.New("order not found") }
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

func (s *OrderService) UpdateStatus(orderID uint, req dto.UpdateStatusRequest) (*dto.OrderResponse, error) {
	var o models.Order
	if err := s.db.First(&o, orderID).Error; err != nil { return nil, errors.New("order not found") }
	if !validTransition(o.Status, req.Status) { return nil, errors.New("invalid status transition") }
	o.Status = req.Status
	s.db.Save(&o)
	resp := dto.ToOrderResponse(&o)
	return &resp, nil
}

func (s *OrderService) Cancel(orderID, clientID uint) error {
	var o models.Order
	if err := s.db.First(&o, orderID).Error; err != nil { return errors.New("order not found") }
	if o.ClientID != clientID { return errors.New("not your order") }
	if o.Status != models.OrderPending { return errors.New("only pending orders can be cancelled") }
	return s.db.Model(&o).Update("status", models.OrderCancelled).Error
}

func toResponses(orders []models.Order) []dto.OrderResponse {
	result := make([]dto.OrderResponse, len(orders))
	for i := range orders { result[i] = dto.ToOrderResponse(&orders[i]) }
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
		if s == to { return true }
	}
	return false
}
