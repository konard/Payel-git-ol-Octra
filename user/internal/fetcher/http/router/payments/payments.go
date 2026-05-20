package payments

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	yoopayment "github.com/rvinnie/yookassa-sdk-go/yookassa/payment"

	"user/internal/core/services"
	"user/internal/fetcher/http/oauth/middleware"
)

func RegisterRoutes(r *gin.Engine) {
	yookassaService := services.NewYooKassaService()

	r.POST("/payments/create", middleware.AuthMiddleware(), func(c *gin.Context) {
		var req struct {
			Amount      int64  `json:"amount"`
			Currency    string `json:"currency"`
			Description string `json:"description"`
			ReturnURL   string `json:"return_url"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "Invalid request body"})
			return
		}

		payment, err := yookassaService.CreatePayment(
			req.Amount,
			req.Currency,
			req.Description,
			req.ReturnURL,
			"",
			"",
		)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "error": "Failed to create payment: " + err.Error()})
			return
		}

		redirectURL := ""
		if redirect, ok := payment.Confirmation.(*yoopayment.Redirect); ok {
			redirectURL = redirect.ConfirmationURL
		}

		c.JSON(200, gin.H{
			"status":       "success",
			"payment_id":   payment.ID,
			"amount":       payment.Amount,
			"payment_status": payment.Status,
			"redirect_url":   redirectURL,
		})
	})

	r.POST("/payments/webhook", func(c *gin.Context) {
		var notification map[string]interface{}
		if err := c.ShouldBindJSON(&notification); err != nil {
			c.JSON(400, gin.H{"status": "error", "error": err.Error()})
			return
		}

		event, _ := notification["event"].(string)
		if event == "payment.succeeded" {
			paymentObj := notification["object"].(map[string]interface{})
			paymentID := paymentObj["id"].(string)
			metadata := paymentObj["metadata"].(map[string]interface{})

			userIDStr := metadata["user_id"].(string)
			planID := metadata["plan_id"].(string)

			_, _ = uuid.Parse(userIDStr)
			// TODO: activate subscription
			_ = planID
			_ = paymentID
		}

		c.JSON(200, gin.H{"status": "ok"})
	})

	r.POST("/payments/simulate-success", middleware.AuthMiddleware(), func(c *gin.Context) {
		_ = c.MustGet("userID").(uuid.UUID)

		var req struct {
			PaymentID string `json:"payment_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "Invalid request body"})
			return
		}

		// TODO: activate by payment ID
		_ = req.PaymentID
		err := error(nil)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": "success"})
	})
}