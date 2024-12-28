package handler

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sefazor/ourphotos-backend/internal/controller"
	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/utils"
)

type EventHandler struct {
	eventController *controller.EventController
	userController  *controller.UserController
}

func NewEventHandler(eventController *controller.EventController, userController *controller.UserController) *EventHandler {
	return &EventHandler{
		eventController: eventController,
		userController:  userController,
	}
}

func (h *EventHandler) CreateEvent(c *fiber.Ctx) error {
	// Önce raw body'yi yazdıralım
	fmt.Printf("Raw body: %s\n", string(c.Body()))

	var req models.EventRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(fmt.Sprintf(
			"Invalid request body: %+v\nRaw body: %s\nError: %v",
			req,
			string(c.Body()),
			err,
		)))
	}

	// Parsed request'i de yazdıralım
	fmt.Printf("Parsed request: %+v\n", req)

	if err := utils.Validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(fmt.Sprintf("Validation error: %+v", err)))
	}

	userIDInterface := c.Locals("user_id")
	if userIDInterface == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse("Unauthorized"))
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse("Invalid user ID format"))
	}

	event, err := h.eventController.CreateEvent(userID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(fmt.Sprintf("Create event error: %v", err)))
	}

	return c.Status(fiber.StatusCreated).JSON(models.SuccessResponse(event, "Event created successfully"))
}

func (h *EventHandler) GetEvent(c *fiber.Ctx) error {
	eventID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid event ID"))
	}

	event, err := h.eventController.GetEvent(uint(eventID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse("Event not found"))
	}

	return c.JSON(models.SuccessResponse(event, "Event retrieved successfully"))
}

func (h *EventHandler) GetUserEvents(c *fiber.Ctx) error {
	// Güvenli type assertion
	userIDRaw := c.Locals("user_id")
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

	// Kullanıcının eventlerini getir
	events, err := h.eventController.GetUserEvents(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    events,
	})
}

func (h *EventHandler) UpdateEvent(c *fiber.Ctx) error {
	eventID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid event ID"))
	}

	var req models.UpdateEventRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid request body"))
	}

	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse("Unauthorized"))
	}

	userID, ok := userIDRaw.(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse("Invalid user ID format"))
	}

	event, err := h.eventController.UpdateEvent(uint(eventID), userID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(err.Error()))
	}

	return c.JSON(models.SuccessResponse(event, "Event updated successfully"))
}

func (h *EventHandler) DeleteEvent(c *fiber.Ctx) error {
	eventID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid event ID"))
	}

	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse("Unauthorized"))
	}

	userID, ok := userIDRaw.(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse("Invalid user ID format"))
	}

	if err := h.eventController.DeleteEvent(uint(eventID), userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(err.Error()))
	}

	return c.JSON(models.SuccessResponse(nil, "Event successfully deleted"))
}

func (h *EventHandler) GetEventByURL(c *fiber.Ctx) error {
	url := c.Params("url")

	event, err := h.eventController.GetEventByURL(url)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse("Event not found"))
	}

	// Public event kontrolü
	if !event.IsPublic {
		// Eğer event public değilse, kullanıcı giriş yapmış olmalı
		userEmail, ok := c.Locals("userEmail").(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse("Unauthorized"))
		}

		user, err := h.userController.GetUserByEmail(userEmail)
		if err != nil || user.ID != event.UserID {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse("Unauthorized"))
		}
	}

	return c.JSON(models.SuccessResponse(event, "Event retrieved successfully"))
}

func (h *EventHandler) CheckEventPassword(c *fiber.Ctx) error {
	eventID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid event ID",
		})
	}

	var req models.EventPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if err := h.eventController.CheckEventPassword(uint(eventID), req.Password); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Incorrect password",
		})
	}

	// Başarılı giriş için cookie oluştur
	cookie := new(fiber.Cookie)
	cookie.Name = fmt.Sprintf("event_%d_access", eventID)
	cookie.Value = "true"
	cookie.Expires = time.Now().Add(24 * time.Hour)
	c.Cookie(cookie)

	return c.JSON(fiber.Map{
		"success": true,
	})
}
