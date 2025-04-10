package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/service"
	"github.com/sefazor/ourphotos-backend/pkg/utils"
)

type EventHandler struct {
	eventService *service.EventService
	userService  *service.UserService
	validator    *utils.Validator
}

func NewEventHandler(eventService *service.EventService, userService *service.UserService, validator *utils.Validator) *EventHandler {
	return &EventHandler{
		eventService: eventService,
		userService:  userService,
		validator:    validator,
	}
}

func (h *EventHandler) CreateEvent(c *fiber.Ctx) error {
	var req models.EventRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid request body"))
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(err.Error()))
	}

	// Get user ID from context
	userID := c.Locals("userID").(uint)

	// Create event
	event, err := h.eventService.CreateEvent(userID, req)
	if err != nil {
		if strings.Contains(err.Error(), "not allowed") {
			return c.Status(fiber.StatusForbidden).JSON(models.ErrorResponse(err.Error()))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse(err.Error()))
	}

	// Event başarıyla oluşturuldu, URL ile birlikte dön
	return c.JSON(models.SuccessResponse(event, "Event created successfully"))
}

func (h *EventHandler) GetEvent(c *fiber.Ctx) error {
	url := c.Params("url")

	// Güvenli type assertion
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		fmt.Printf("userID is nil in context\n")
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse("User not authenticated"))
	}

	userID, ok := userIDRaw.(uint)
	if !ok {
		fmt.Printf("userID type assertion failed. Type: %T, Value: %v\n", userIDRaw, userIDRaw)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse("Invalid user ID format"))
	}

	fmt.Printf("Getting event with URL: %s for userID: %d\n", url, userID)

	event, err := h.eventService.GetEventByURL(url)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse("Event not found"))
	}

	if event.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(models.ErrorResponse("You don't have permission to view this event"))
	}

	return c.JSON(models.SuccessResponse(event, "Event retrieved successfully"))
}

func (h *EventHandler) GetUserEvents(c *fiber.Ctx) error {
	// Güvenli type assertion
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		fmt.Printf("userID is nil in context\n")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "User not authenticated",
		})
	}

	userID, ok := userIDRaw.(uint)
	if !ok {
		fmt.Printf("userID type assertion failed. Type: %T, Value: %v\n", userIDRaw, userIDRaw)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid user ID format",
		})
	}

	fmt.Printf("Getting events for userID: %d\n", userID)

	// Kullanıcının eventlerini getir
	events, err := h.eventService.GetUserEvents(userID)
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
	url := c.Params("url")

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

	// URL'ye göre event'i al
	event, err := h.eventService.GetEventByURL(url)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse("Event not found"))
	}

	// Kullanıcı bu event'i güncelleme yetkisine sahip mi?
	if event.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(models.ErrorResponse("You don't have permission to update this event"))
	}

	updatedEvent, err := h.eventService.UpdateEvent(event.ID, userID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(err.Error()))
	}

	return c.JSON(models.SuccessResponse(updatedEvent, "Event updated successfully"))
}

func (h *EventHandler) DeleteEvent(c *fiber.Ctx) error {
	url := c.Params("url")

	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse("Unauthorized"))
	}

	userID, ok := userIDRaw.(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse("Invalid user ID format"))
	}

	// URL'ye göre event'i al
	event, err := h.eventService.GetEventByURL(url)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse("Event not found"))
	}

	// Kullanıcı bu event'i silme yetkisine sahip mi?
	if event.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(models.ErrorResponse("You don't have permission to delete this event"))
	}

	if err := h.eventService.DeleteEvent(event.ID, userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(err.Error()))
	}

	return c.JSON(models.SuccessResponse(nil, "Event successfully deleted"))
}

func (h *EventHandler) GetEventByURL(c *fiber.Ctx) error {
	url := c.Params("url")

	event, err := h.eventService.GetEventByURL(url)
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

		user, err := h.userService.GetUserByEmail(userEmail)
		if err != nil || user.ID != event.UserID {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse("Unauthorized"))
		}
	}

	return c.JSON(models.SuccessResponse(event, "Event retrieved successfully"))
}

func (h *EventHandler) CheckEventPassword(c *fiber.Ctx) error {
	url := c.Params("url")

	var req models.EventPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if err := h.eventService.CheckEventPassword(url, req.Password); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Incorrect password",
		})
	}

	// Başarılı giriş için cookie oluştur
	cookie := new(fiber.Cookie)
	cookie.Name = fmt.Sprintf("event_%s_access", url)
	cookie.Value = "true"
	cookie.Expires = time.Now().Add(24 * time.Hour)
	c.Cookie(cookie)

	return c.JSON(fiber.Map{
		"success": true,
	})
}

func (h *EventHandler) UploadEventPhotos(c *fiber.Ctx) error {
	url := c.Params("url")

	// Kullanıcı varsa al, yoksa 0 (misafir)
	var userID uint = 0
	if userIDRaw := c.Locals("userID"); userIDRaw != nil {
		if id, ok := userIDRaw.(uint); ok {
			userID = id
			fmt.Printf("Authenticated user upload - UserID: %d\n", userID)
		}
	} else {
		fmt.Printf("Guest upload - UserID: 0\n")
	}

	// Event kontrolü
	event, err := h.eventService.GetEventByURL(url)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse("Event not found"))
	}

	// Yetki kontrolü - sadece AllowGuestUploads false ise ve misafirse engelle
	if userID == 0 && !event.AllowGuestUploads {
		return c.Status(fiber.StatusForbidden).JSON(models.ErrorResponse("Guest uploads are not allowed for this event"))
	}

	// Multipart form
	form, err := c.MultipartForm()
	if err != nil {
		fmt.Printf("Error parsing multipart form: %v\n", err)
		fmt.Printf("Request Content-Type: %s\n", c.Get("Content-Type"))
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid form data"))
	}

	files := form.File["files"]
	fmt.Printf("Number of files received: %d\n", len(files))
	if len(files) == 0 {
		fmt.Printf("No files found in form. Available fields: %v\n", form.Value)
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("No files uploaded"))
	}

	var uploadedPhotos []models.PhotoResponse
	for i, file := range files {
		fmt.Printf("Uploading file %d/%d for userID: %d\n", i+1, len(files), userID)
		photo, err := h.eventService.UploadEventPhoto(event.ID, userID, file)
		if err != nil {
			fmt.Printf("Error uploading file %s: %v\n", file.Filename, err)
			return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse(err.Error()))
		}
		fmt.Printf("Successfully uploaded file %s\n", file.Filename)
		uploadedPhotos = append(uploadedPhotos, *photo)
	}

	fmt.Printf("Successfully uploaded %d photos\n", len(uploadedPhotos))
	return c.JSON(models.SuccessResponse(uploadedPhotos, "Photos uploaded successfully"))
}

// GetEventQRCode, belirtilen etkinlik için QR kodu oluşturur ve döndürür
func (h *EventHandler) GetEventQRCode(c *fiber.Ctx) error {
	url := c.Params("url")

	// Kullanıcı kimliğini kontrol et
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse("Unauthorized"))
	}

	userID, ok := userIDRaw.(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse("Invalid user ID format"))
	}

	// Event'i al
	event, err := h.eventService.GetEventByURL(url)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse("Event not found"))
	}

	// Kullanıcı bu event'e erişim yetkisine sahip mi?
	if event.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(models.ErrorResponse("You don't have permission to access this event"))
	}

	// QR kod boyutunu al, varsayılan 256px
	sizeStr := c.Query("size", "256")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 || size > 1024 {
		size = 256 // Geçersiz boyut için varsayılan değer
	}

	// QR kodu oluştur
	qrCode, err := h.eventService.GetEventQRCode(event.ID, size)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse(err.Error()))
	}

	// QR kodu PNG olarak döndür
	c.Set("Content-Type", "image/png")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=event-%s-qr.png", url))
	return c.Send(qrCode)
}
