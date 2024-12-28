package service

import (
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/repository"
	"github.com/sefazor/ourphotos-backend/pkg/storage"
)

type PhotoService struct {
	photoRepo  *repository.PhotoRepository
	eventRepo  *repository.EventRepository
	r2Storage  *storage.CloudflareStorage
	imgStorage *storage.CloudflareImages
}

func NewPhotoService(
	photoRepo *repository.PhotoRepository,
	eventRepo *repository.EventRepository,
	r2Storage *storage.CloudflareStorage,
	imgStorage *storage.CloudflareImages,
) *PhotoService {
	return &PhotoService{
		photoRepo:  photoRepo,
		eventRepo:  eventRepo,
		r2Storage:  r2Storage,
		imgStorage: imgStorage,
	}
}

func (s *PhotoService) UploadPhoto(eventID uint, userID uint, file *multipart.FileHeader) (*models.Photos, error) {
	// Event'i kontrol et
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	// Yetki kontrolü
	if userID == 0 { // Misafir kullanıcı
		if !event.IsPublic || !event.AllowGuestUploads {
			return nil, errors.New("unauthorized")
		}
	} else if event.UserID != userID { // Giriş yapmış ama event sahibi değil
		if !event.IsPublic {
			return nil, errors.New("unauthorized")
		}
		if !event.AllowGuestUploads {
			return nil, errors.New("this event does not allow guest uploads")
		}
	}

	// Fotoğraf limit kontrolü
	photoCount, err := s.photoRepo.GetEventPhotoCount(eventID)
	if err != nil {
		return nil, err
	}

	if event.PhotoLimit > 0 && int(photoCount) >= event.PhotoLimit {
		return nil, errors.New("photo limit reached for this event")
	}

	// Dosya tipini kontrol et
	if !isValidImageType(file.Header.Get("Content-Type")) {
		return nil, errors.New("invalid file type")
	}

	// Dosya boyutunu kontrol et (örn: 10MB)
	if file.Size > 10*1024*1024 {
		return nil, errors.New("file size too large")
	}

	// Dosyayı aç
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// R2'ye yükle
	r2Key := fmt.Sprintf("events/%d/%s", eventID, file.Filename)
	if err := s.r2Storage.Upload(r2Key, src); err != nil {
		return nil, err
	}

	// Cloudflare Images'a yükle
	imageID, variants, err := s.imgStorage.UploadImage(file)
	if err != nil {
		// R2'den sil (cleanup)
		_ = s.r2Storage.Delete(r2Key)
		return nil, err
	}

	// DB'ye kaydet
	photo := &models.Photos{
		EventID:    eventID,
		UserID:     userID,
		R2Key:      r2Key,
		ImageID:    imageID,
		Variants:   variants,
		IsUploaded: true,
		UploadedAt: time.Now(),
	}

	if err := s.photoRepo.Create(photo); err != nil {
		// Cleanup
		_ = s.r2Storage.Delete(r2Key)
		_ = s.imgStorage.Delete(imageID)
		return nil, err
	}

	return photo, nil
}

func (s *PhotoService) GetEventPhotos(eventID uint, userID uint) ([]models.Photos, error) {
	// Önce event'i kontrol et
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	// Event'in public olup olmadığını kontrol et
	if !event.IsPublic {
		// Event public değilse, sadece event sahibi görebilir
		if event.UserID != userID {
			return nil, errors.New("unauthorized")
		}
	}

	// Fotoğrafları getir
	photos, err := s.photoRepo.GetByEventID(eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get photos: %v", err)
	}

	return photos, nil
}

func (s *PhotoService) DeletePhoto(photoID uint, userID uint) error {
	photo, err := s.photoRepo.GetByID(photoID)
	if err != nil {
		return fmt.Errorf("photo not found: %w", err)
	}

	// Yetki kontrolü
	if photo.UserID != userID {
		return errors.New("unauthorized")
	}

	// Önce storage'dan sil
	if err := s.r2Storage.Delete(photo.R2Key); err != nil {
		return fmt.Errorf("failed to delete from storage: %w", err)
	}

	if err := s.imgStorage.Delete(photo.ImageID); err != nil {
		return fmt.Errorf("failed to delete from image service: %w", err)
	}

	// Veritabanından sil
	return s.photoRepo.Delete(photoID)
}

func (s *PhotoService) GetPublicEventPhotos(eventURL string) ([]models.Photos, error) {
	// Önce event'i bul
	event, err := s.eventRepo.GetByURL(eventURL)
	if err != nil {
		return nil, errors.New("event not found")
	}

	// Event public değilse hata dön
	if !event.IsPublic {
		return nil, errors.New("event is not public")
	}

	// Event'in fotoğraflarını getir
	return s.photoRepo.GetByEventID(event.ID)
}

func isValidImageType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	return validTypes[contentType]
}
