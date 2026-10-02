package service

import (
	"context"
	"errors"
	"time"

	"github.com/runns/payment-service/dto"
	"github.com/runns/shared/models"
	"github.com/runns/shared/rabbitmq"
	"github.com/runns/shared/utils"
	"gorm.io/gorm"
)

type PaymentService struct {
	db  *gorm.DB
	rmq *rabbitmq.Client
}

func NewPaymentService(db *gorm.DB, rmq *rabbitmq.Client) *PaymentService {
	return &PaymentService{db: db, rmq: rmq}
}

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

	if payment.Status == models.PaymentSuccess && s.rmq != nil {
		_ = s.rmq.PublishJSON(context.Background(), "rydex_events", "payment.successful", resp)
	}

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

	if s.rmq != nil {
		_ = s.rmq.PublishJSON(context.Background(), "rydex_events", "payment.successful", resp)
	}

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

func (s *PaymentService) GetRiderWallet(riderID uint) (*dto.RiderWalletResponse, error) {
	var wallet models.RiderWallet
	if err := s.db.Where("user_id = ?", riderID).First(&wallet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			wallet = models.RiderWallet{UserID: riderID, Balance: 0, PendingBalance: 0}
			if err := s.db.Create(&wallet).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("rider wallet not found")
		}
	}
	return &dto.RiderWalletResponse{RiderID: riderID, Balance: wallet.Balance, PendingBalance: wallet.PendingBalance}, nil
}

func (s *PaymentService) TransferToRiderWallet(clientID uint, req dto.WalletTransferRequest) (*dto.WalletTransferResponse, error) {
	if req.Amount <= 0 { return nil, errors.New("amount must be positive") }
	if req.RiderID == 0 { return nil, errors.New("rider ID required") }

	var clientProfile models.ClientProfile
	if err := s.db.Where("user_id = ?", clientID).First(&clientProfile).Error; err != nil {
		return nil, errors.New("client profile not found")
	}
	if clientProfile.WalletBalance < req.Amount { return nil, errors.New("insufficient wallet balance") }

	var riderProfile models.RiderProfile
	if err := s.db.Where("user_id = ?", req.RiderID).First(&riderProfile).Error; err != nil {
		return nil, errors.New("rider not found")
	}
	if !riderProfile.NINVerified { return nil, errors.New("rider NIN not verified") }

	var riderWallet models.RiderWallet
	if err := s.db.Where("user_id = ?", req.RiderID).First(&riderWallet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			riderWallet = models.RiderWallet{UserID: req.RiderID, Balance: 0, PendingBalance: 0}
			if err := s.db.Create(&riderWallet).Error; err != nil { return nil, err }
		} else { return nil, err }
	}

	tx := s.db.Begin()
	clientProfile.WalletBalance -= req.Amount
	if err := tx.Save(&clientProfile).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	riderWallet.PendingBalance += req.Amount
	if err := tx.Save(&riderWallet).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	ref := utils.GeneratePaymentRef()
	payment := &models.Payment{
		OrderID: 0, ClientID: clientID, Amount: req.Amount,
		Method: models.PaymentWallet, Status: models.PaymentSuccess,
		Reference: ref, PaidAt: func() *time.Time { t := time.Now(); return &t }(),
	}
	if err := tx.Create(payment).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	if s.rmq != nil {
		event := dto.WalletTransferEvent{
			PaymentID:  payment.ID,
			ClientID:   clientID,
			RiderID:    req.RiderID,
			Amount:     req.Amount,
			Reference:  ref,
			Timestamp:  time.Now(),
		}
		_ = s.rmq.PublishJSON(context.Background(), "rydex_events", "wallet.transferred", event)
	}

	return &dto.WalletTransferResponse{
		Reference: ref, Amount: req.Amount,
		ClientNewBalance: clientProfile.WalletBalance,
		RiderNewBalance:  riderWallet.Balance + req.Amount,
	}, nil
}

func (s *PaymentService) ReleaseRiderFunds(riderID uint, amount float64, reference string) error {
	if amount <= 0 { return errors.New("amount must be positive") }
	var wallet models.RiderWallet
	if err := s.db.Where("user_id = ?", riderID).First(&wallet).Error; err != nil {
		return errors.New("rider wallet not found")
	}
	if wallet.PendingBalance < amount { return errors.New("insufficient pending balance") }
	wallet.PendingBalance -= amount
	wallet.Balance += amount
	return s.db.Save(&wallet).Error
}

func (s *PaymentService) VerifyNIN(riderID uint, nin string) error {
	var profile models.RiderProfile
	if err := s.db.Where("user_id = ?", riderID).First(&profile).Error; err != nil {
		return errors.New("rider profile not found")
	}
	if profile.NINVerified { return errors.New("NIN already verified") }
	profile.NIN = nin
	profile.NINVerified = true
	now := time.Now()
	profile.NINVerifiedAt = &now
	return s.db.Save(&profile).Error
}

func (s *PaymentService) GetNINStatus(riderID uint) (*dto.NINStatusResponse, error) {
	var profile models.RiderProfile
	if err := s.db.Where("user_id = ?", riderID).First(&profile).Error; err != nil {
		return nil, errors.New("rider profile not found")
	}
	return &dto.NINStatusResponse{
		RiderID:     riderID,
		NINVerified: profile.NINVerified,
		NIN:         profile.NIN,
		VerifiedAt:  profile.NINVerifiedAt,
	}, nil
}

func (s *PaymentService) SyncUser(event rabbitmq.UserRegisteredEvent) error {
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
			if err := s.db.Create(&user).Error; err != nil {
				return err
			}
			// Create profile based on role
			if event.Data.Role == string(models.RoleClient) {
				profile := models.ClientProfile{UserID: event.Data.UserID}
				return s.db.Create(&profile).Error
			} else if event.Data.Role == string(models.RoleRider) {
				profile := models.RiderProfile{
					UserID:       event.Data.UserID,
					VehicleType:  models.VehicleType(event.Data.VehicleType),
					VehiclePlate: event.Data.VehiclePlate,
					LicenseNumber: event.Data.LicenseNumber,
				}
				return s.db.Create(&profile).Error
			}
			return nil
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
