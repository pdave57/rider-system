package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/runns/dispatch-service/dto"
	"github.com/runns/shared/models"
	"gorm.io/gorm"
)

type DispatchService struct {
	db       *gorm.DB       // runns_dispatch — Dispatch records
	ordersDB *gorm.DB       // runns_orders  — Order records
	ridersDB *gorm.DB       // runns_riders  — RiderProfile records
	rdb      *redis.Client
}

func NewDispatchService(db, ordersDB, ridersDB *gorm.DB, rdb *redis.Client) *DispatchService {
	return &DispatchService{db: db, ordersDB: ordersDB, ridersDB: ridersDB, rdb: rdb}
}

func (s *DispatchService) Assign(req dto.AssignRequest) (*dto.DispatchResponse, error) {
	// Load order from the orders database
	var order models.Order
	if err := s.ordersDB.First(&order, req.OrderID).Error; err != nil {
		return nil, errors.New("order not found")
	}
	if order.Status != models.OrderPending {
		return nil, errors.New("order is not pending")
	}

	riderID := req.RiderID
	if riderID == 0 {
		var err error
		riderID, err = s.nearestRider(order.PickupLatitude, order.PickupLongitude)
		if err != nil {
			return nil, errors.New("no available riders nearby")
		}
	}

	// Verify rider availability from riders database
	var profile models.RiderProfile
	if err := s.ridersDB.Where("user_id = ? AND is_available = true", riderID).First(&profile).Error; err != nil {
		return nil, errors.New("rider is not available")
	}

	// Create dispatch record in dispatch database
	dispatch := &models.Dispatch{OrderID: req.OrderID, RiderID: riderID, AssignedAt: time.Now()}
	if err := s.db.Create(dispatch).Error; err != nil {
		return nil, err
	}

	// Update order status in orders database
	s.ordersDB.Model(&order).Updates(map[string]interface{}{"status": models.OrderAssigned, "rider_id": riderID})

	// Mark rider as unavailable in riders database
	s.ridersDB.Model(&models.RiderProfile{}).Where("user_id = ?", riderID).Update("is_available", false)

	resp := dto.ToDispatchResponse(dispatch)
	return &resp, nil
}

func (s *DispatchService) Respond(dispatchID, riderID uint, req dto.RespondRequest) (*dto.DispatchResponse, error) {
	var d models.Dispatch
	if err := s.db.First(&d, dispatchID).Error; err != nil {
		return nil, errors.New("dispatch not found")
	}
	if d.RiderID != riderID {
		return nil, errors.New("not your dispatch")
	}
	if d.AcceptedAt != nil || d.RejectedAt != nil {
		return nil, errors.New("already responded")
	}
	now := time.Now()
	if req.Accept {
		d.AcceptedAt = &now
	} else {
		d.RejectedAt = &now
		// Revert order status in orders database
		s.ordersDB.Model(&models.Order{}).Where("id = ?", d.OrderID).
			Updates(map[string]interface{}{"status": models.OrderPending, "rider_id": nil})
		// Make rider available again in riders database
		s.ridersDB.Model(&models.RiderProfile{}).Where("user_id = ?", riderID).
			Update("is_available", true)
	}
	s.db.Save(&d)
	resp := dto.ToDispatchResponse(&d)
	return &resp, nil
}

func (s *DispatchService) GetByOrder(orderID uint) (*dto.DispatchResponse, error) {
	var d models.Dispatch
	if err := s.db.Where("order_id = ?", orderID).Order("created_at desc").First(&d).Error; err != nil {
		return nil, errors.New("no dispatch found")
	}
	resp := dto.ToDispatchResponse(&d)
	return &resp, nil
}

func (s *DispatchService) ListPendingForRider(riderID uint) ([]dto.DispatchResponse, error) {
	var dispatches []models.Dispatch
	s.db.Where("rider_id = ? AND accepted_at IS NULL AND rejected_at IS NULL", riderID).Find(&dispatches)
	result := make([]dto.DispatchResponse, len(dispatches))
	for i := range dispatches {
		result[i] = dto.ToDispatchResponse(&dispatches[i])
	}
	return result, nil
}

func (s *DispatchService) UpdateRiderLocation(riderID uint, lat, lon float64) error {
	ctx := context.Background()
	event := models.GPSEvent{RiderID: riderID, Latitude: lat, Longitude: lon, Timestamp: time.Now()}
	b, _ := json.Marshal(event)
	s.rdb.Set(ctx, fmt.Sprintf("gps:rider:%d", riderID), b, 10*time.Minute)
	return s.rdb.GeoAdd(ctx, "rider:locations", &redis.GeoLocation{
		Name: fmt.Sprintf("%d", riderID), Latitude: lat, Longitude: lon,
	}).Err()
}

func (s *DispatchService) nearestRider(lat, lon float64) (uint, error) {
	ctx := context.Background()
	results, err := s.rdb.GeoRadius(ctx, "rider:locations", lon, lat,
		&redis.GeoRadiusQuery{Radius: 50, Unit: "km", WithCoord: true, Sort: "ASC", Count: 10}).Result()
	if err != nil || len(results) == 0 {
		return 0, errors.New("no riders found in geo index")
	}
	bestDist := math.MaxFloat64
	var bestID uint
	for _, loc := range results {
		var id uint
		fmt.Sscanf(loc.Name, "%d", &id)
		// Check availability in riders database
		var profile models.RiderProfile
		if err := s.ridersDB.Where("user_id = ? AND is_available = true", id).First(&profile).Error; err != nil {
			continue
		}
		dist := haversine(lat, lon, loc.Latitude, loc.Longitude)
		if dist < bestDist {
			bestDist = dist
			bestID = id
		}
	}
	if bestID == 0 {
		return 0, errors.New("no available riders")
	}
	return bestID, nil
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}