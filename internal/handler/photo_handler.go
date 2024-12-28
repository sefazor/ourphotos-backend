package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sefazor/ourphotos-backend/internal/controller"
	"github.com/sefazor/ourphotos-backend/internal/models"
)

type PhotoHandler struct {
	photoController *controller.PhotoController
	userController  *controller.UserController
	eventController *controller.EventController
}

func NewPhotoHandler(photoController *controller.PhotoController, userController *controller.UserController, eventController *controller.EventController) *PhotoHandler {
	return &PhotoHandler{
		photoController: photoController,
		userController:  userController,
		eventController: eventController,
	}
}

func (h *PhotoHandler) UploadPhoto(c *fiber.Ctx) error {
	// Get event_id from form
	eventIDStr := c.FormValue("event_id")
	if eventIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Event ID is required",
		})
	}

	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid event ID",
		})
	}

	// Get user_id from JWT (optional for public events)
	var userID uint = 0
	userIDInterface := c.Locals("user_id")
	if userIDInterface != nil {
		userID, _ = userIDInterface.(uint)
	}

	// Get file from form
	file, err := c.FormFile("photo")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Photo is required",
		})
	}

	// Check if event allows guest uploads if no user is authenticated
	if userID == 0 {
		event, err := h.eventController.GetEvent(uint(eventID))
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"error":   "Event not found",
			})
		}

		if !event.IsPublic || !event.AllowGuestUploads {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "This event does not allow guest uploads",
			})
		}
	}

	// Upload photo
	photo, err := h.photoController.UploadPhoto(uint(eventID), userID, file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Photo uploaded successfully",
		"data":    photo,
	})
}

func (h *PhotoHandler) GetEventPhotos(c *fiber.Ctx) error {
	eventID, err := strconv.ParseUint(c.Params("eventId"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid event ID",
		})
	}

	// User ID'yi JWT'den al ve nil kontrolü yap
	userIDInterface := c.Locals("user_id")
	if userIDInterface == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid user ID format",
		})
	}

	photos, err := h.photoController.GetEventPhotos(uint(eventID), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Photos retrieved successfully",
		"data":    photos,
	})
}

// DeletePhotoRequest struct'ı ekleyelim
type DeletePhotoRequest struct {
	PhotoID uint `json:"photo_id"`
}

func (h *PhotoHandler) DeletePhoto(c *fiber.Ctx) error {
	// Get photo_id from URL parameter
	photoIDStr := c.Params("id")
	photoID, err := strconv.ParseUint(photoIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid photo ID",
		})
	}

	userIDInterface := c.Locals("user_id")
	if userIDInterface == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse("Unauthorized"))
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse("Invalid user ID format"))
	}

	// Delete photo
	if err := h.photoController.DeletePhoto(uint(photoID), userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Photo deleted successfully",
	})
}

func (h *PhotoHandler) GetPublicEventPhotos(c *fiber.Ctx) error {
	// Get event URL from params
	eventURL := c.Params("url")
	if eventURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Event URL is required",
		})
	}

	// Get photos
	photos, err := h.photoController.GetPublicEventPhotos(eventURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Photos retrieved successfully",
		"data":    photos,
	})
}
