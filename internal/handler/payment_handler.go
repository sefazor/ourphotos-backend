package handler

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sefazor/ourphotos-backend/internal/controller"
	"github.com/stripe/stripe-go/v74/webhook"
)

type PaymentHandler struct {
	paymentController *controller.PaymentController
}

func NewPaymentHandler(paymentController *controller.PaymentController) *PaymentHandler {
	return &PaymentHandler{
		paymentController: paymentController,
	}
}

func (h *PaymentHandler) CreateCheckoutSession(c *fiber.Ctx) error {
	packageID, err := strconv.ParseUint(c.Params("packageId"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid package ID",
		})
	}

	// Güvenli type assertion
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "User not authenticated",
		})
	}

	userID, ok := userIDRaw.(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid user ID format",
		})
	}

	session, err := h.paymentController.CreateCheckoutSession(userID, uint(packageID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    session,
	})
}

func (h *PaymentHandler) HandleStripeWebhook(c *fiber.Ctx) error {
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	payload := c.Body()
	signatureHeader := c.Get("Stripe-Signature")

	// Debug için loglar
	fmt.Printf("Webhook Secret: %s\n", webhookSecret)
	fmt.Printf("Signature Header: %s\n", signatureHeader)
	fmt.Printf("Payload Length: %d\n", len(payload))

	// API version mismatch'i ignore et
	event, err := webhook.ConstructEventWithOptions(payload, signatureHeader, webhookSecret,
		webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true,
		})
	if err != nil {
		fmt.Printf("Webhook Error: %v\n", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fmt.Sprintf("Webhook error: %v", err),
		})
	}

	if err := h.paymentController.HandleStripeWebhook(&event); err != nil {
		fmt.Printf("Controller Error: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *PaymentHandler) GetCreditPackages(c *fiber.Ctx) error {
	packages, err := h.paymentController.GetCreditPackages()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    packages,
	})
}

func (h *PaymentHandler) GetPurchaseHistory(c *fiber.Ctx) error {
	fmt.Println("GetPurchaseHistory handler called") // Debug log

	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "User not authenticated",
		})
	}

	userID, ok := userIDRaw.(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid user ID format",
		})
	}

	purchases, err := h.paymentController.GetUserPurchaseHistory(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    purchases,
	})
}
