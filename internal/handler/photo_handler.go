package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/service"
)

type PhotoHandler struct {
	photoService *service.PhotoService
}

func NewPhotoHandler(photoService *service.PhotoService) *PhotoHandler {
	return &PhotoHandler{
		photoService: photoService,
	}
}

func (h *PhotoHandler) GetEventPhotos(c *fiber.Ctx) error {
	eventID, err := strconv.ParseUint(c.Params("eventId"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid event ID"))
	}

	userID := c.Locals("userID").(uint)

	photos, err := h.photoService.GetEventPhotos(uint(eventID), userID)
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
			MediumURL:    h.photoService.ImgStorage.GetMediumURL(photo.ImageID),
			LargeURL:     h.photoService.ImgStorage.GetLargeURL(photo.ImageID),
			IsGuest:      photo.IsGuest,
			CreatedAt:    photo.CreatedAt,
		})
	}

	return c.JSON(models.SuccessResponse(responses, "Photos retrieved successfully"))
}

func (h *PhotoHandler) UploadPhoto(c *fiber.Ctx) error {
	eventID, err := strconv.ParseUint(c.Params("eventId"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid event ID"))
	}

	userID := c.Locals("userID").(uint)

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("No file uploaded"))
	}

	response, err := h.photoService.UploadPhoto(uint(eventID), userID, file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse(err.Error()))
	}

	return c.JSON(models.SuccessResponse(response, "Photo uploaded successfully"))
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
			MediumURL:    h.photoService.ImgStorage.GetMediumURL(photo.ImageID),
			LargeURL:     h.photoService.ImgStorage.GetLargeURL(photo.ImageID),
			IsGuest:      photo.IsGuest,
			CreatedAt:    photo.CreatedAt,
		})
	}

	return c.JSON(models.SuccessResponse(responses, "Photos retrieved successfully"))
}
