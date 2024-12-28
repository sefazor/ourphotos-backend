package payment

import (
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
)

type StripeService struct {
	secretKey string
}

func NewStripeService(secretKey string) *StripeService {
	stripe.Key = secretKey
	return &StripeService{
		secretKey: secretKey,
	}
}

func (s *StripeService) CreateCheckoutSession(userEmail string, priceID string, metadata map[string]string) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionParams{
		CustomerEmail: &userEmail,
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String("http://localhost:3000/payment/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String("http://localhost:3000/payment/cancel"),
	}

	params.AddMetadata("user_id", metadata["user_id"])
	params.AddMetadata("package_id", metadata["package_id"])

	session, err := session.New(params)
	if err != nil {
		return nil, err
	}

	return session, nil
}
