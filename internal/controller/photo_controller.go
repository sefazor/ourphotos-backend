package controller

import (
	"mime/multipart"

	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/service"
)

type PhotoController struct {
	photoService *service.PhotoService
}

func NewPhotoController(photoService *service.PhotoService) *PhotoController {
	return &PhotoController{
		photoService: photoService,
	}
}

func (c *PhotoController) UploadPhoto(eventID uint, userID uint, file *multipart.FileHeader) (*models.Photos, error) {
	return c.photoService.UploadPhoto(eventID, userID, file)
}

func (c *PhotoController) GetEventPhotos(eventID uint, userID uint) ([]models.Photos, error) {
	return c.photoService.GetEventPhotos(eventID, userID)
}

func (c *PhotoController) DeletePhoto(photoID uint, userID uint) error {
	return c.photoService.DeletePhoto(photoID, userID)
}

func (c *PhotoController) GetPublicEventPhotos(eventURL string) ([]models.Photos, error) {
	return c.photoService.GetPublicEventPhotos(eventURL)
}
