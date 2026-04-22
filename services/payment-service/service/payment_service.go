package service

import (
	"errors"
	"time"

	"github.com/rydex/payment-service/dto"
	"github.com/rydex/shared/models"
	"github.com/rydex/shared/utils"
	"gorm.io/gorm"
)

type PaymentService struct{ db *gorm.DB }

func NewPaymentService(db *gorm.DB) *PaymentService { return &PaymentService{db: db} }

func (s *PaymentService) Initiate(clientID uint, req dto.InitiatePaymentRequest) (*dto.PaymentResponse, error) {
	var order models.Order
	if err := s.db.First(&order, req.OrderID).Error; err != nil { return nil, errors.New("order not found") }
	if order.ClientID != clientID { return nil, errors.New("not your order") }
	if order.Status == models.OrderCancelled || order.Status == models.OrderDelivered {
		return nil, errors.New("order cannot be paid in current state")
	}
	var existing models.Payment
	if s.db.Where("order_id = ? AND status = ?", req.OrderID, models.PaymentSuccess).First(&existing).Error == nil {
		return nil, errors.New("order already paid")
	}
	payment := &models.Payment{
		OrderID: req.OrderID, ClientID: clientID, Amount: order.DeliveryFee,
		Method: req.Method, Status: models.PaymentPending, Reference: utils.GeneratePaymentRef(),
	}
	if req.Method == models.PaymentWallet {
		var profile models.ClientProfile
		if err := s.db.Where("user_id = ?", clientID).First(&profile).Error; err != nil {
			return nil, errors.New("client profile not found")
		}
		if profile.WalletBalance < order.DeliveryFee { return nil, errors.New("insufficient wallet balance") }
		profile.WalletBalance -= order.DeliveryFee
		s.db.Save(&profile)
		now := time.Now()
		payment.Status = models.PaymentSuccess
		payment.PaidAt = &now
	}
	if err := s.db.Create(payment).Error; err != nil { return nil, err }
	resp := dto.ToPaymentResponse(payment)
	return &resp, nil
}

func (s *PaymentService) Verify(req dto.VerifyPaymentRequest) (*dto.PaymentResponse, error) {
	var p models.Payment
	if err := s.db.Where("reference = ?", req.Reference).First(&p).Error; err != nil {
		return nil, errors.New("payment not found")
	}
	if p.Status == models.PaymentSuccess { return nil, errors.New("already verified") }
	now := time.Now()
	p.Status = models.PaymentSuccess
	p.GatewayRef = req.GatewayRef
	p.PaidAt = &now
	s.db.Save(&p)
	resp := dto.ToPaymentResponse(&p)
	return &resp, nil
}

func (s *PaymentService) GetByOrder(orderID uint) (*dto.PaymentResponse, error) {
	var p models.Payment
	if err := s.db.Where("order_id = ?", orderID).First(&p).Error; err != nil {
		return nil, errors.New("payment not found")
	}
	resp := dto.ToPaymentResponse(&p)
	return &resp, nil
}

func (s *PaymentService) GetWallet(clientID uint) (*dto.WalletResponse, error) {
	var profile models.ClientProfile
	if err := s.db.Where("user_id = ?", clientID).First(&profile).Error; err != nil {
		return nil, errors.New("client profile not found")
	}
	return &dto.WalletResponse{ClientID: clientID, WalletBalance: profile.WalletBalance}, nil
}

func (s *PaymentService) TopupWallet(clientID uint, req dto.WalletTopupRequest) (*dto.WalletResponse, error) {
	if req.Amount <= 0 { return nil, errors.New("amount must be positive") }
	var profile models.ClientProfile
	if err := s.db.Where("user_id = ?", clientID).First(&profile).Error; err != nil {
		return nil, errors.New("client profile not found")
	}
	profile.WalletBalance += req.Amount
	s.db.Save(&profile)
	return &dto.WalletResponse{ClientID: clientID, WalletBalance: profile.WalletBalance}, nil
}
