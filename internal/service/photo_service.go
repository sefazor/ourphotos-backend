package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/repository"
	"github.com/sefazor/ourphotos-backend/pkg/storage"
)

type PhotoService struct {
	photoRepo  *repository.PhotoRepository
	eventRepo  *repository.EventRepository
	userRepo   *repository.UserRepository
	r2Storage  *storage.CloudflareStorage
	ImgStorage *storage.CloudflareImages
}

func NewPhotoService(
	photoRepo *repository.PhotoRepository,
	eventRepo *repository.EventRepository,
	r2Storage *storage.CloudflareStorage,
	ImgStorage *storage.CloudflareImages,
	userRepo *repository.UserRepository,
) *PhotoService {
	return &PhotoService{
		photoRepo:  photoRepo,
		eventRepo:  eventRepo,
		r2Storage:  r2Storage,
		ImgStorage: ImgStorage,
		userRepo:   userRepo,
	}
}

func (s *PhotoService) UploadPhoto(eventID uint, userID uint, file *multipart.FileHeader) (*models.PhotoResponse, error) {
	fmt.Printf("UploadPhoto called - EventID: %d, UserID: %d\n", eventID, userID)

	// Event'i bul
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		fmt.Printf("Error getting event: %v\n", err)
		return nil, err
	}

	// Event'in photo limitini kontrol et
	if event.PhotoLimit <= 0 {
		return nil, errors.New("event photo limit exceeded")
	}

	// Event sahibinin limitini kontrol et
	eventOwner, err := s.userRepo.GetByID(event.UserID)
	if err != nil {
		fmt.Printf("Error getting event owner: %v\n", err)
		return nil, err
	}

	fmt.Printf("Current photo limit for event owner (ID: %d): %d\n", eventOwner.ID, eventOwner.PhotoLimit)
	if eventOwner.PhotoLimit <= 0 {
		return nil, errors.New("event owner's photo limit exceeded")
	}

	// Eğer userID 0 ise (guest upload) ve event guest upload'a izin vermiyorsa hata dön
	if userID == 0 && !event.AllowGuestUploads {
		return nil, errors.New("guest uploads are not allowed for this event")
	}

	// Eğer giriş yapmış kullanıcı ise limit kontrolü yap
	if userID > 0 {
		fmt.Printf("Checking limits for user: %d\n", userID)

		user, err := s.userRepo.GetByID(userID)
		if err != nil {
			fmt.Printf("Error getting user: %v\n", err)
			return nil, err
		}

		fmt.Printf("Current photo limit for user: %d\n", user.PhotoLimit)

		// Photo limit kontrolü
		if user.PhotoLimit <= 0 {
			return nil, errors.New("photo limit exceeded")
		}

		// Fotoğraf başarıyla yüklendiyse limiti düş
		user.PhotoLimit--
		fmt.Printf("Decreasing photo limit to: %d\n", user.PhotoLimit)

		if err := s.userRepo.Update(user); err != nil {
			fmt.Printf("Error updating user: %v\n", err)
			return nil, err
		}
		fmt.Printf("Successfully updated user photo limit\n")
	} else {
		fmt.Printf("Guest upload - no limit check needed\n")
	}

	// Dosyayı aç
	fileContent, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer fileContent.Close()

	// Dosya içeriğini byte array'e oku
	fileBytes, err := io.ReadAll(fileContent)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Dosya adı oluştur
	fileName := generateUniqueFileName()

	// R2'ye yükle
	if err := s.r2Storage.Upload(fileName, bytes.NewReader(fileBytes)); err != nil {
		return nil, fmt.Errorf("failed to upload to R2: %w", err)
	}

	// Cloudflare Images'a yükle
	imageID, _, err := s.ImgStorage.Upload(bytes.NewReader(fileBytes))
	if err != nil {
		// R2'den dosyayı sil
		_ = s.r2Storage.Delete(fileName)
		return nil, fmt.Errorf("failed to upload to Cloudflare Images: %w", err)
	}

	// Fotoğraf kaydı oluştur
	photo := &models.Photos{
		EventID:    eventID,
		UserID:     userID,
		FileName:   file.Filename,
		FileSize:   file.Size,
		MimeType:   file.Header.Get("Content-Type"),
		R2Key:      fileName,
		ImageID:    imageID,
		PublicURL:  s.ImgStorage.GetPublicURL(imageID),
		IsGuest:    userID == 0,
		UploadedAt: time.Now(),
	}

	// Veritabanına kaydet
	err = s.photoRepo.Create(photo)
	if err != nil {
		// Hata durumunda yüklenen dosyaları sil
		_ = s.r2Storage.Delete(fileName)
		_ = s.ImgStorage.Delete(imageID)
		return nil, err
	}

	// Response için URL'leri oluştur
	response := &models.PhotoResponse{
		ID:           photo.ID,
		EventID:      photo.EventID,
		UserID:       photo.UserID,
		FileName:     photo.FileName,
		FileSize:     photo.FileSize,
		MimeType:     photo.MimeType,
		PublicURL:    s.ImgStorage.GetPublicURL(photo.ImageID),
		ThumbnailURL: s.ImgStorage.GetThumbnailURL(photo.ImageID),
		MediumURL:    s.ImgStorage.GetMediumURL(photo.ImageID),
		LargeURL:     s.ImgStorage.GetLargeURL(photo.ImageID),
		IsGuest:      photo.IsGuest,
		CreatedAt:    photo.UploadedAt,
	}

	// Fotoğraf başarıyla yüklendiyse limitleri düş
	eventOwner.PhotoLimit--
	if err := s.userRepo.Update(eventOwner); err != nil {
		return nil, err
	}

	// Event'in photo limitini düş
	event.PhotoLimit--
	event.PhotoCount++
	if err := s.eventRepo.Update(event); err != nil {
		return nil, err
	}

	return response, nil
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

	if err := s.ImgStorage.Delete(photo.ImageID); err != nil {
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

func (s *PhotoService) GetEventPhotoCount(eventID uint) (int64, error) {
	return s.photoRepo.GetEventPhotoCount(eventID)
}

// Yardımcı fonksiyonlar
func generateUniqueFileName() string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), uuid.New().String())
}

type fileInfo struct {
	mimeType string
	// Gerekirse başka bilgiler eklenebilir
}

func getFileInfo(fileBytes []byte) fileInfo {
	// http.DetectContentType kullanarak mime type'ı belirle
	mimeType := http.DetectContentType(fileBytes)
	return fileInfo{
		mimeType: mimeType,
	}
}
