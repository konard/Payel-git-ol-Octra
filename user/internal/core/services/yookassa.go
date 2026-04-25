package services

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/rvinnie/yookassa-sdk-go/yookassa"
	yoocommon "github.com/rvinnie/yookassa-sdk-go/yookassa/common"
	yoopayment "github.com/rvinnie/yookassa-sdk-go/yookassa/payment"
)

type YooKassaService struct {
	handler *yookassa.PaymentHandler
}

func NewYooKassaService() *YooKassaService {
	shopID := os.Getenv("YOOKASSA_SHOP_ID")
	if shopID == "" {
		shopID = "1339826"
	}

	secretKey := os.Getenv("YOOKASSA_SECRET_KEY")
	if secretKey == "" {
		secretKey = "test_StL4_VJfVFbOJ7_BbolU2VhoR1zjIQ7Qf2gcwN3Gngw"
	}

	client := yookassa.NewClient(shopID, secretKey)

	return &YooKassaService{
		handler: yookassa.NewPaymentHandler(client),
	}
}

func (s *YooKassaService) CreatePayment(amount int64, currency, description, returnUrl, userID, planID string) (*yoopayment.Payment, error) {
	paymentID := uuid.New().String()

	payment := &yoopayment.Payment{
		Amount: &yoocommon.Amount{
			Value:    fmt.Sprintf("%.2f", float64(amount)/100),
			Currency: currency,
		},
		Description: description,
		Capture:     true,
		Confirmation: &yoopayment.Redirect{
			Type:      yoopayment.TypeRedirect,
			ReturnURL: returnUrl,
		},
		Metadata: map[string]interface{}{
			"payment_id": paymentID,
			"user_id":    userID,
			"plan_id":    planID,
		},
	}

	return s.handler.CreatePayment(context.Background(), payment)
}

func (s *YooKassaService) GetPayment(paymentID string) (*yoopayment.Payment, error) {
	return s.handler.FindPayment(context.Background(), paymentID)
}

func (s *YooKassaService) CancelPayment(paymentID string) error {
	_, err := s.handler.CancelPayment(context.Background(), paymentID)
	return err
}