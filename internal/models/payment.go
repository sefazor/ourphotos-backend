package models

type CreateCheckoutSessionRequest struct {
	PlanID string `json:"plan_id" validate:"required"`
}

type CheckoutSession struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	CustomerID string `json:"customer_id"`
	PaymentID  string `json:"payment_id"`
	Status     string `json:"status"`
}
