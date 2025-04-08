package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/service"
)

type PhotoHandler struct {
	photoService *service.PhotoService
	eventService *service.EventService
}

func NewPhotoHandler(photoService *service.PhotoService, eventService *service.EventService) *PhotoHandler {
	return &PhotoHandler{
		photoService: photoService,
		eventService: eventService,
	}
}

func (h *PhotoHandler) GetEventPhotos(c *fiber.Ctx) error {
	url := c.Params("url")

	userID := c.Locals("userID").(uint)

	// URL'den etkinlik ID'sini al
	event, err := h.eventService.GetEventByURL(url)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse("Event not found"))
	}

	photos, err := h.photoService.GetEventPhotos(event.ID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse(err.Error()))
	}

	var responses []models.PhotoResponse
	for _, photo := range photos {
		responses = append(responses, models.PhotoResponse{
			ID:           photo.ID,
			EventID:      photo.EventID,
			UserID:       photo.UserID,
			FileName:     photo.FileName,
			FileSize:     photo.FileSize,
			MimeType:     photo.MimeType,
			PublicURL:    photo.PublicURL,
			ThumbnailURL: h.photoService.ImgStorage.GetThumbnailURL(photo.ImageID),
			IsGuest:      photo.IsGuest,
			CreatedAt:    photo.CreatedAt,
		})
	}

	return c.JSON(models.SuccessResponse(responses, "Photos retrieved successfully"))
}

func (h *PhotoHandler) UploadPhoto(c *fiber.Ctx) error {
	url := c.Params("url")

	// Bu bir misafir yükleme endpoint'i olduğu için userID her zaman 0 (misafir)
	var userID uint = 0

	// URL'den etkinlik ID'sini al
	event, err := h.eventService.GetEventByURL(url)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse("Event not found"))
	}

	// Etkinlik misafir yüklemelerine izin veriyor mu kontrol et
	if !event.AllowGuestUploads {
		return c.Status(fiber.StatusForbidden).JSON(models.ErrorResponse("Guest uploads are not allowed for this event"))
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("No file uploaded"))
	}

	response, err := h.photoService.UploadPhoto(event.ID, userID, file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse(err.Error()))
	}

	return c.JSON(models.SuccessResponse(response, "Photo uploaded successfully as guest"))
}

func (h *PhotoHandler) DeletePhoto(c *fiber.Ctx) error {
	photoID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid photo ID"))
	}

	userID := c.Locals("userID").(uint)

	if err := h.photoService.DeletePhoto(uint(photoID), userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse(err.Error()))
	}

	return c.JSON(models.SuccessResponse(nil, "Photo deleted successfully"))
}

func (h *PhotoHandler) GetPublicEventPhotos(c *fiber.Ctx) error {
	eventURL := c.Params("url")

	photos, err := h.photoService.GetPublicEventPhotos(eventURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse(err.Error()))
	}

	var responses []models.PhotoResponse
	for _, photo := range photos {
		responses = append(responses, models.PhotoResponse{
			ID:           photo.ID,
			EventID:      photo.EventID,
			UserID:       photo.UserID,
			FileName:     photo.FileName,
			FileSize:     photo.FileSize,
			MimeType:     photo.MimeType,
			PublicURL:    photo.PublicURL,
			ThumbnailURL: h.photoService.ImgStorage.GetThumbnailURL(photo.ImageID),
			IsGuest:      photo.IsGuest,
			CreatedAt:    photo.CreatedAt,
		})
	}

	return c.JSON(models.SuccessResponse(responses, "Photos retrieved successfully"))
}
