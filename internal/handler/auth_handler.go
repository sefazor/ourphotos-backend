package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sefazor/ourphotos-backend/internal/controller"
	"github.com/sefazor/ourphotos-backend/internal/models"
)

type AuthHandler struct {
	authController *controller.AuthController
}

func NewAuthHandler(authController *controller.AuthController) *AuthHandler {
	return &AuthHandler{
		authController: authController,
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid request body"))
	}

	user, err := h.authController.Register(req)
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

	token, err := h.authController.Login(req)
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

	if err := h.authController.ForgotPassword(req.Email); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(err.Error()))
	}

	return c.JSON(models.SuccessResponse(nil, "Password reset email sent"))
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req models.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid request body"))
	}

	if err := h.authController.ResetPassword(req.Token, req.NewPassword); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(err.Error()))
	}

	return c.JSON(models.SuccessResponse(nil, "Password reset successful"))
}
