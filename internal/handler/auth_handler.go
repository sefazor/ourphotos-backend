package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/service"
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
		fmt.Printf("Body parse error: %v\n", err)
		fmt.Printf("Raw body: %s\n", string(c.Body()))
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid request body"))
	}

	// Validation
	if req.Token == "" || req.NewPassword == "" {
		fmt.Printf("Invalid request: Token=%v, NewPassword=%v\n", req.Token != "", req.NewPassword != "")
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Token and new password are required"))
	}

	fmt.Printf("Reset Password Request:\n")
	fmt.Printf("Token length: %d\n", len(req.Token))
	fmt.Printf("New Password length: %d\n", len(req.NewPassword))

	if err := h.authService.ResetPassword(req.Token, req.NewPassword); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(err.Error()))
	}

	return c.JSON(models.SuccessResponse(nil, "Password reset successful"))
}
