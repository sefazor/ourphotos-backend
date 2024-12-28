package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sefazor/ourphotos-backend/internal/controller"
	"github.com/sefazor/ourphotos-backend/internal/models"
)

type UserHandler struct {
	userController *controller.UserController
}

func NewUserHandler(userController *controller.UserController) *UserHandler {
	return &UserHandler{
		userController: userController,
	}
}

func (h *UserHandler) GetMyProfile(c *fiber.Ctx) error {
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

	user, err := h.userController.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    user,
	})
}

func (h *UserHandler) ChangePassword(c *fiber.Ctx) error {
	var req models.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid request body"))
	}

	userEmail := c.Locals("userEmail").(string)
	user, err := h.userController.GetUserByEmail(userEmail)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse("User not found"))
	}

	if err := h.userController.ChangePassword(user.ID, req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(err.Error()))
	}

	return c.JSON(models.SuccessResponse(nil, "Password changed successfully"))
}

func (h *UserHandler) InitiateEmailChange(c *fiber.Ctx) error {
	var req models.ChangeEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid request body"))
	}

	userEmail := c.Locals("userEmail").(string)
	user, err := h.userController.GetUserByEmail(userEmail)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse("User not found"))
	}

	if err := h.userController.InitiateEmailChange(user.ID, req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(err.Error()))
	}

	return c.JSON(models.SuccessResponse(nil, "Verification email sent to new email address"))
}

func (h *UserHandler) CompleteEmailChange(c *fiber.Ctx) error {
	var req models.VerifyEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid request body"))
	}

	if err := h.userController.CompleteEmailChange(req.Token); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(err.Error()))
	}

	return c.JSON(models.SuccessResponse(nil, "Email changed successfully"))
}

func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	var req models.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid request body"))
	}

	userEmail, ok := c.Locals("userEmail").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse("Unauthorized"))
	}

	user, err := h.userController.GetUserByEmail(userEmail)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse("User not found"))
	}

	updatedUser, err := h.userController.UpdateProfile(user.ID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(err.Error()))
	}

	return c.JSON(models.SuccessResponse(updatedUser, "Profile updated successfully"))
}
