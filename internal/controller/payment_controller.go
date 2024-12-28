package controller

import (
	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/service"
	"github.com/stripe/stripe-go/v74"
)

type PaymentController struct {
	paymentService *service.PaymentService
}

func NewPaymentController(paymentService *service.PaymentService) *PaymentController {
	return &PaymentController{
		paymentService: paymentService,
	}
}

func (c *PaymentController) CreateCheckoutSession(userID uint, packageID uint) (*models.CheckoutSession, error) {
	return c.paymentService.CreateCheckoutSession(userID, packageID)
}

func (c *PaymentController) HandleStripeWebhook(event *stripe.Event) error {
	return c.paymentService.HandleStripeWebhook(event)
}

func (c *PaymentController) GetCreditPackages() ([]models.CreditPackage, error) {
	return c.paymentService.GetCreditPackages()
}

func (c *PaymentController) GetUserPurchaseHistory(userID uint) ([]models.UserCreditPurchase, error) {
	return c.paymentService.GetUserPurchaseHistory(userID)
}
