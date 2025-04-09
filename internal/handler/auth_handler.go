package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/service"
	"github.com/sefazor/ourphotos-backend/pkg/captcha"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid request body"))
	}

	// Turnstile doğrulama
	isValid, err := captcha.VerifyTurnstile(req.TurnstileToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse("Error verifying CAPTCHA: " + err.Error()))
	}

	if !isValid {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("CAPTCHA verification failed"))
	}

	user, err := h.authService.Register(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(models.SuccessResponse(user, "User registered successfully"))
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid request body"))
	}

	// Turnstile doğrulama
	isValid, err := captcha.VerifyTurnstile(req.TurnstileToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse("Error verifying CAPTCHA: " + err.Error()))
	}

	if !isValid {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("CAPTCHA verification failed"))
	}

	token, err := h.authService.Login(req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse(err.Error()))
	}

	return c.JSON(models.SuccessResponse(fiber.Map{
		"token": token,
	}, "Login successful"))
}

func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req models.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid request body"))
	}

	if err := h.authService.ForgotPassword(req.Email); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(err.Error()))
	}

	return c.JSON(models.SuccessResponse(nil, "Password reset email sent"))
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req models.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid request body"))
	}

	// Validation
	if req.Token == "" || req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Token and new password are required"))
	}

	if err := h.authService.ResetPassword(req.Token, req.NewPassword); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(err.Error()))
	}

	return c.JSON(models.SuccessResponse(nil, "Password reset successful"))
}

func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	var req models.VerifyEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid request body"))
	}

	if req.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Token is required"))
	}

	if err := h.authService.VerifyEmail(req.Token); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(err.Error()))
	}

	return c.JSON(models.SuccessResponse(nil, "Email verified successfully"))
}

func (h *AuthHandler) ResendVerificationEmail(c *fiber.Ctx) error {
	var req models.ForgotPasswordRequest // Aynı modeli kullanabiliriz, sadece email gerekli
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid request body"))
	}

	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Email is required"))
	}

	if err := h.authService.ResendVerificationEmail(req.Email); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(err.Error()))
	}

	return c.JSON(models.SuccessResponse(nil, "Verification email sent"))
}
