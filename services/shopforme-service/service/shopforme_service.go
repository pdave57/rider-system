package service

import (
	"context"
	"errors"
	"log"
	"math"
	"time"

	"github.com/runns/shopforme-service/dto"
	"github.com/runns/shared/models"
	"github.com/runns/shared/rabbitmq"
	"github.com/runns/shared/utils"
	"gorm.io/gorm"
)

type ShopForMeService struct {
	db         *gorm.DB
	ordersDB   *gorm.DB
	ridersDB   *gorm.DB
	paymentsDB *gorm.DB
	rmq        *rabbitmq.Client
}

func NewShopForMeService(db, ordersDB, ridersDB, paymentsDB *gorm.DB, rmq *rabbitmq.Client) *ShopForMeService {
	return &ShopForMeService{
		db:         db,
		ordersDB:   ordersDB,
		ridersDB:   ridersDB,
		paymentsDB: paymentsDB,
		rmq:        rmq,
	}
}

func (s *ShopForMeService) CreateRequest(clientID uint, req dto.CreateShopForMeRequest) (*dto.ShopForMeResponse, error) {
	if req.PickupAddress == "" || req.DropoffAddress == "" {
		return nil, errors.New("pickup_address and dropoff_address are required")
	}

	request := &models.ShopForMeRequest{
		ClientID:        clientID,
		Status:          models.ShopForMePending,
		PickupAddress:   req.PickupAddress,
		PickupLatitude:  req.PickupLatitude,
		PickupLongitude: req.PickupLongitude,
		DropoffAddress:  req.DropoffAddress,
		DropoffLatitude: req.DropoffLatitude,
		DropoffLongitude: req.DropoffLongitude,
		Notes:           req.Notes,
		ServiceFee:      500.00,
	}

	if err := s.db.Create(request).Error; err != nil {
		return nil, err
	}

	go s.findAndMatchRider(request.ID)

	if s.rmq != nil {
		event := rabbitmq.ShopForMeCreatedEvent{
			RequestID:  request.ID,
			ClientID:   clientID,
			PickupLat:  request.PickupLatitude,
			PickupLng:  request.PickupLongitude,
			DropoffLat: request.DropoffLatitude,
			DropoffLng: request.DropoffLongitude,
			Timestamp:  time.Now(),
		}
		_ = s.rmq.PublishJSON(context.Background(), "rydex_events", rabbitmq.RoutingKeyShopForMeCreated, event)
	}

	resp := dto.ToShopForMeResponse(request)
	return &resp, nil
}

func (s *ShopForMeService) findAndMatchRider(requestID uint) {
	var request models.ShopForMeRequest
	if err := s.db.First(&request, requestID).Error; err != nil {
		log.Printf("[shopforme] request %d not found for matching: %v", requestID, err)
		return
	}

	var profiles []models.RiderProfile
	s.ridersDB.Where("is_available = true AND is_verified = true AND nin_verified = true").Find(&profiles)

	if len(profiles) == 0 {
		log.Printf("[shopforme] no available verified riders for request %d", requestID)
		return
	}

	ids := make([]uint, len(profiles))
	profileMap := make(map[uint]*models.RiderProfile)
	for i, p := range profiles {
		ids[i] = p.UserID
		cp := profiles[i]
		profileMap[p.UserID] = &cp
	}

	var users []models.User
	s.ordersDB.Where("id IN ?", ids).Find(&users)

	var bestRider *models.User
	bestDist := math.MaxFloat64

	for _, u := range users {
		if p, ok := profileMap[u.ID]; ok {
			dist := haversine(request.PickupLatitude, request.PickupLongitude, p.CurrentLatitude, p.CurrentLongitude)
			if dist < bestDist && dist <= 20 {
				bestDist = dist
				bestRider = &u
			}
		}
	}

	if bestRider == nil {
		log.Printf("[shopforme] no riders within 20km for request %d", requestID)
		return
	}

	now := time.Now()
	request.RiderID = &bestRider.ID
	request.Status = models.ShopForMeMatched
	request.MatchedAt = &now
	if err := s.db.Save(&request).Error; err != nil {
		log.Printf("[shopforme] failed to update request %d with rider match: %v", requestID, err)
		return
	}

	if s.rmq != nil {
		event := rabbitmq.ShopForMeMatchedEvent{
			RequestID: request.ID,
			ClientID:  request.ClientID,
			RiderID:   bestRider.ID,
			Timestamp: time.Now(),
		}
		_ = s.rmq.PublishJSON(context.Background(), "rydex_events", rabbitmq.RoutingKeyShopForMeMatched, event)
	}

	log.Printf("[shopforme] request %d matched with rider %d", requestID, bestRider.ID)
}

func (s *ShopForMeService) AcceptRequest(riderID uint, requestID uint) (*dto.ShopForMeResponse, error) {
	var request models.ShopForMeRequest
	if err := s.db.First(&request, requestID).Error; err != nil {
		return nil, errors.New("request not found")
	}
	if request.RiderID == nil || *request.RiderID != riderID {
		return nil, errors.New("not your request")
	}
	if request.Status != models.ShopForMeMatched {
		return nil, errors.New("request cannot be accepted in current state")
	}

	now := time.Now()
	request.Status = models.ShopForMeAccepted
	request.AcceptedAt = &now
	if err := s.db.Save(&request).Error; err != nil {
		return nil, err
	}

	if s.rmq != nil {
		event := rabbitmq.ShopForMeAcceptedEvent{
			RequestID: request.ID,
			RiderID:   riderID,
			Timestamp: time.Now(),
		}
		_ = s.rmq.PublishJSON(context.Background(), "rydex_events", rabbitmq.RoutingKeyShopForMeAccepted, event)
	}

	resp := dto.ToShopForMeResponse(&request)
	return &resp, nil
}

func (s *ShopForMeService) PlaceOrder(clientID uint, requestID uint, req dto.PlaceOrderRequest) (*dto.ShopForMeResponse, error) {
	var request models.ShopForMeRequest
	if err := s.db.Preload("Items").First(&request, requestID).Error; err != nil {
		return nil, errors.New("request not found")
	}
	if request.ClientID != clientID {
		return nil, errors.New("not your request")
	}
	if request.Status != models.ShopForMeAccepted {
		return nil, errors.New("order can only be placed after rider acceptance")
	}
	if len(req.Items) == 0 {
		return nil, errors.New("at least one item required")
	}

	var totalAmount float64
	for _, itemReq := range req.Items {
		if itemReq.Quantity <= 0 || itemReq.UnitPrice <= 0 {
			return nil, errors.New("invalid quantity or price")
		}
		item := models.ShopForMeItem{
			RequestID:   request.ID,
			Name:        itemReq.Name,
			Description: itemReq.Description,
			Quantity:    itemReq.Quantity,
			UnitPrice:   itemReq.UnitPrice,
			TotalPrice:  float64(itemReq.Quantity) * itemReq.UnitPrice,
		}
		if err := s.db.Create(&item).Error; err != nil {
			return nil, err
		}
		totalAmount += item.TotalPrice
	}

	request.Items = nil
	s.db.Preload("Items").First(&request, requestID)

	request.TotalAmount = totalAmount + request.ServiceFee
	request.Status = models.ShopForMeOrdering
	now := time.Now()
	request.OrderPlacedAt = &now
	if err := s.db.Save(&request).Error; err != nil {
		return nil, err
	}

	if s.rmq != nil {
		event := rabbitmq.ShopForMeOrderPlacedEvent{
			RequestID:   request.ID,
			ClientID:    clientID,
			RiderID:     *request.RiderID,
			TotalAmount: request.TotalAmount,
			ServiceFee:  request.ServiceFee,
			Timestamp:   time.Now(),
		}
		_ = s.rmq.PublishJSON(context.Background(), "rydex_events", rabbitmq.RoutingKeyShopForMeOrderPlaced, event)
	}

	resp := dto.ToShopForMeResponse(&request)
	return &resp, nil
}

func (s *ShopForMeService) VerifyItems(riderID uint, requestID uint, req dto.VerifyItemsRequest) (*dto.ShopForMeResponse, error) {
	var request models.ShopForMeRequest
	if err := s.db.Preload("Order").First(&request, requestID).Error; err != nil {
		return nil, errors.New("request not found")
	}
	if request.RiderID == nil || *request.RiderID != riderID {
		return nil, errors.New("not your request")
	}
	if request.Status != models.ShopForMeOrdering && request.Status != models.ShopForMeShopping {
		return nil, errors.New("items can only be verified in ordering/shopping state")
	}

	now := time.Now()
	if req.Verified {
		request.Status = models.ShopForMeVerifying

		order := &models.ShopForMeOrder{
			RequestID:   request.ID,
			RiderID:     riderID,
			ClientID:    request.ClientID,
			TotalAmount: request.TotalAmount,
			ServiceFee:  request.ServiceFee,
			WalletAmount: request.TotalAmount,
			Status:      "pending_payment",
		}
		if err := s.db.Create(order).Error; err != nil {
			return nil, err
		}
		request.Order = order

		order.VerifiedAt = &now
		s.db.Save(order)
	} else {
		request.Status = models.ShopForMeCancelled
		request.CancelledAt = &now
	}

	if err := s.db.Save(&request).Error; err != nil {
		return nil, err
	}

	if s.rmq != nil {
		event := rabbitmq.ShopForMeVerifiedEvent{
			RequestID: request.ID,
			RiderID:   riderID,
			Timestamp: time.Now(),
		}
		_ = s.rmq.PublishJSON(context.Background(), "rydex_events", rabbitmq.RoutingKeyShopForMeVerified, event)
	}

	resp := dto.ToShopForMeResponse(&request)
	return &resp, nil
}

func (s *ShopForMeService) PayForOrder(clientID uint, requestID uint) (*dto.ShopForMeResponse, error) {
	var request models.ShopForMeRequest
	if err := s.db.Preload("Order").First(&request, requestID).Error; err != nil {
		return nil, errors.New("request not found")
	}
	if request.ClientID != clientID {
		return nil, errors.New("not your request")
	}
	if request.Order == nil {
		return nil, errors.New("order not created yet")
	}
	if request.Status != models.ShopForMeVerifying {
		return nil, errors.New("payment can only be made after verification")
	}

	var clientProfile models.ClientProfile
	if err := s.paymentsDB.Where("user_id = ?", clientID).First(&clientProfile).Error; err != nil {
		return nil, errors.New("client profile not found")
	}
	if clientProfile.WalletBalance < request.TotalAmount {
		return nil, errors.New("insufficient wallet balance")
	}

	var riderWallet models.RiderWallet
	if err := s.paymentsDB.Where("user_id = ?", *request.RiderID).First(&riderWallet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			riderWallet = models.RiderWallet{UserID: *request.RiderID, Balance: 0, PendingBalance: 0}
			if err := s.paymentsDB.Create(&riderWallet).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	tx := s.paymentsDB.Begin()
	clientProfile.WalletBalance -= request.TotalAmount
	if err := tx.Save(&clientProfile).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	riderWallet.PendingBalance += request.TotalAmount
	if err := tx.Save(&riderWallet).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	ref := utils.GeneratePaymentRef()
	payment := &models.Payment{
		OrderID: 0, ClientID: clientID, Amount: request.TotalAmount,
		Method: models.PaymentWallet, Status: models.PaymentSuccess,
		Reference: ref, PaidAt: func() *time.Time { t := time.Now(); return &t }(),
	}
	if err := tx.Create(payment).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	now := time.Now()
	request.Order.PaymentRef = ref
	request.Order.Status = "paid"
	request.Order.PaidAt = &now
	request.Status = models.ShopForMePaid
	request.PaidAt = &now
	if err := s.db.Save(&request).Error; err != nil {
		return nil, err
	}
	s.db.Save(request.Order)

	if s.rmq != nil {
		event := rabbitmq.ShopForMePaidEvent{
			RequestID:    request.ID,
			ClientID:     clientID,
			RiderID:      *request.RiderID,
			Amount:       request.TotalAmount,
			WalletAmount: request.TotalAmount,
			PaymentRef:   ref,
			Timestamp:    time.Now(),
		}
		_ = s.rmq.PublishJSON(context.Background(), "rydex_events", rabbitmq.RoutingKeyShopForMePaid, event)
	}

	resp := dto.ToShopForMeResponse(&request)
	return &resp, nil
}

func (s *ShopForMeService) DispatchOrder(riderID uint, requestID uint) (*dto.ShopForMeResponse, error) {
	var request models.ShopForMeRequest
	if err := s.db.Preload("Order").First(&request, requestID).Error; err != nil {
		return nil, errors.New("request not found")
	}
	if request.RiderID == nil || *request.RiderID != riderID {
		return nil, errors.New("not your request")
	}
	if request.Status != models.ShopForMePaid {
		return nil, errors.New("can only dispatch paid orders")
	}

	now := time.Now()
	request.Status = models.ShopForMeDispatched
	request.DispatchedAt = &now
	request.Order.Status = "dispatched"
	if err := s.db.Save(&request).Error; err != nil {
		return nil, err
	}
	s.db.Save(request.Order)

	if s.rmq != nil {
		event := rabbitmq.ShopForMeDispatchedEvent{
			RequestID: request.ID,
			RiderID:   riderID,
			Timestamp: time.Now(),
		}
		_ = s.rmq.PublishJSON(context.Background(), "rydex_events", rabbitmq.RoutingKeyShopForMeDispatched, event)
	}

	resp := dto.ToShopForMeResponse(&request)
	return &resp, nil
}

func (s *ShopForMeService) DeliverOrder(riderID uint, requestID uint) (*dto.ShopForMeResponse, error) {
	var request models.ShopForMeRequest
	if err := s.db.Preload("Order").First(&request, requestID).Error; err != nil {
		return nil, errors.New("request not found")
	}
	if request.RiderID == nil || *request.RiderID != riderID {
		return nil, errors.New("not your request")
	}
	if request.Status != models.ShopForMeDispatched {
		return nil, errors.New("can only deliver dispatched orders")
	}

	now := time.Now()
	request.Status = models.ShopForMeDelivered
	request.DeliveredAt = &now
	request.Order.Status = "delivered"
	if err := s.db.Save(&request).Error; err != nil {
		return nil, err
	}
	s.db.Save(request.Order)

	var wallet models.RiderWallet
	if err := s.paymentsDB.Where("user_id = ?", riderID).First(&wallet).Error; err == nil {
		if wallet.PendingBalance >= request.TotalAmount {
			wallet.PendingBalance -= request.TotalAmount
			wallet.Balance += request.TotalAmount
			s.paymentsDB.Save(&wallet)
		}
	}

	if s.rmq != nil {
		event := rabbitmq.ShopForMeDeliveredEvent{
			RequestID: request.ID,
			RiderID:   riderID,
			Timestamp: time.Now(),
		}
		_ = s.rmq.PublishJSON(context.Background(), "rydex_events", rabbitmq.RoutingKeyShopForMeDelivered, event)
	}

	resp := dto.ToShopForMeResponse(&request)
	return &resp, nil
}

func (s *ShopForMeService) CancelRequest(userID uint, requestID uint, isClient bool) (*dto.ShopForMeResponse, error) {
	var request models.ShopForMeRequest
	if err := s.db.First(&request, requestID).Error; err != nil {
		return nil, errors.New("request not found")
	}

	if isClient {
		if request.ClientID != userID {
			return nil, errors.New("not your request")
		}
	} else {
		if request.RiderID == nil || *request.RiderID != userID {
			return nil, errors.New("not your request")
		}
	}

	if request.Status == models.ShopForMeDelivered || request.Status == models.ShopForMeCancelled {
		return nil, errors.New("request cannot be cancelled")
	}

	now := time.Time{}
	request.Status = models.ShopForMeCancelled
	request.CancelledAt = &now
	if request.Order != nil {
		request.Order.Status = "cancelled"
		s.db.Save(request.Order)
	}
	if err := s.db.Save(&request).Error; err != nil {
		return nil, err
	}

	var riderID *uint
	if request.RiderID != nil {
		riderID = request.RiderID
	}

	if s.rmq != nil {
		event := rabbitmq.ShopForMeCancelledEvent{
			RequestID: request.ID,
			ClientID:  request.ClientID,
			RiderID:   riderID,
			Reason:    "cancelled by user",
			Timestamp: time.Now(),
		}
		_ = s.rmq.PublishJSON(context.Background(), "rydex_events", rabbitmq.RoutingKeyShopForMeCancelled, event)
	}

	resp := dto.ToShopForMeResponse(&request)
	return &resp, nil
}

func (s *ShopForMeService) GetRequest(requestID uint) (*dto.ShopForMeResponse, error) {
	var request models.ShopForMeRequest
	if err := s.db.Preload("Items").Preload("Order").First(&request, requestID).Error; err != nil {
		return nil, errors.New("request not found")
	}
	resp := dto.ToShopForMeResponse(&request)
	return &resp, nil
}

func (s *ShopForMeService) ListByClient(clientID uint) ([]dto.ShopForMeResponse, error) {
	var requests []models.ShopForMeRequest
	s.db.Preload("Items").Preload("Order").Where("client_id = ?", clientID).Order("created_at desc").Find(&requests)
	result := make([]dto.ShopForMeResponse, len(requests))
	for i := range requests {
		result[i] = dto.ToShopForMeResponse(&requests[i])
	}
	return result, nil
}

func (s *ShopForMeService) ListByRider(riderID uint) ([]dto.ShopForMeResponse, error) {
	var requests []models.ShopForMeRequest
	s.db.Preload("Items").Preload("Order").Where("rider_id = ?", riderID).Order("created_at desc").Find(&requests)
	result := make([]dto.ShopForMeResponse, len(requests))
	for i := range requests {
		result[i] = dto.ToShopForMeResponse(&requests[i])
	}
	return result, nil
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}