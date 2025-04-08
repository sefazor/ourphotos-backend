package service

import (
	"bytes"
	"context"
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
	ImgStorage *storage.CloudflareImages
}

func NewPhotoService(
	photoRepo *repository.PhotoRepository,
	eventRepo *repository.EventRepository,
	ImgStorage *storage.CloudflareImages,
	userRepo *repository.UserRepository,
) *PhotoService {
	return &PhotoService{
		photoRepo:  photoRepo,
		eventRepo:  eventRepo,
		userRepo:   userRepo,
		ImgStorage: ImgStorage,
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
	var user *models.User
	if userID > 0 {
		fmt.Printf("Checking limits for user: %d\n", userID)

		var err error
		user, err = s.userRepo.GetByID(userID)
		if err != nil {
			fmt.Printf("Error getting user: %v\n", err)
			return nil, err
		}

		fmt.Printf("Current photo limit for user: %d\n", user.PhotoLimit)

		// Photo limit kontrolü
		if user.PhotoLimit <= 0 {
			return nil, errors.New("photo limit exceeded")
		}

		// ÖNEMLİ: Limiti henüz düşürme, yükleme başarılı olduktan sonra düşür
	} else {
		fmt.Printf("Guest upload - no limit check needed\n")
	}

	// Dosyayı aç
	fileContent, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer fileContent.Close()

	// MIME type algılama için sadece başlangıç kısmını oku
	headerBytes := make([]byte, 512)
	_, err = fileContent.Read(headerBytes)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read file header: %w", err)
	}

	// Dosyayı başa sar
	if _, err = fileContent.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek file: %w", err)
	}

	// MIME type'ı belirle
	mimeType := http.DetectContentType(headerBytes)

	// Dosya içeriğini okuyalım - MultipartFileHeader'ı bir kez kullanma sınırlamasını aşmak için
	fmt.Printf("Dosya içeriğini okuyoruz...\n")
	fileData, err := io.ReadAll(fileContent)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}
	fmt.Printf("Dosya içeriği okundu, boyut: %d bytes\n", len(fileData))

	// Cloudflare Images'a yükleme yap
	cfReader := bytes.NewReader(fileData)

	// Yükleme işlemi için timeout ekle
	uploadCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Yükleme işlemini başlat
	uploadDone := make(chan struct {
		imageID string
		err     error
	}, 1)

	go func() {
		// Cloudflare Images'a yükleme
		fmt.Printf("Starting Cloudflare Images upload\n")
		imageID, _, err := s.ImgStorage.Upload(cfReader)
		uploadDone <- struct {
			imageID string
			err     error
		}{imageID, err}
	}()

	// Yükleme işleminin tamamlanmasını bekle veya timeout
	var imageID string
	var cfErr error
	select {
	case <-uploadCtx.Done():
		return nil, fmt.Errorf("upload timed out after 60 seconds")
	case result := <-uploadDone:
		imageID = result.imageID
		cfErr = result.err
	}

	// Hata kontrolü
	if cfErr != nil {
		fmt.Printf("Cloudflare Images upload failed: %v\n", cfErr)
		return nil, cfErr
	}

	fmt.Printf("Cloudflare Images upload completed successfully, image ID: %s\n", imageID)

	// Fotoğraf kaydı oluştur
	photo := &models.Photos{
		EventID:    eventID,
		UserID:     userID,
		FileName:   file.Filename,
		FileSize:   file.Size,
		MimeType:   mimeType,
		ImageID:    imageID,
		PublicURL:  s.ImgStorage.GetPublicURL(imageID),
		IsGuest:    userID == 0,
		UploadedAt: time.Now(),
	}

	// Veritabanına kaydet
	err = s.photoRepo.Create(photo)
	if err != nil {
		// Hata durumunda yüklenen resmi sil
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
		IsGuest:      photo.IsGuest,
		CreatedAt:    photo.UploadedAt,
	}

	// ÖNEMLİ: Şimdi başarılı yüklemeden sonra limitleri düşür

	// 1. Kullanıcı limiti güncelleme (eğer giriş yapmış bir kullanıcıysa)
	if userID > 0 && user != nil {
		// Fotoğraf başarıyla yüklendiyse limiti düş
		user.PhotoLimit--
		fmt.Printf("Decreasing user photo limit to: %d\n", user.PhotoLimit)

		if err := s.userRepo.Update(user); err != nil {
			// Bu noktada fotoğraf zaten yüklendi, sadece kredi düşürme başarısız oldu
			// Log atıp devam edebiliriz, ama ideal olan işlemi geri almak olurdu
			fmt.Printf("Warning: Failed to update user photo limit: %v\n", err)
		} else {
			fmt.Printf("Successfully updated user photo limit\n")
		}
	}

	// 2. Event owner limiti güncelleme
	eventOwner.PhotoLimit--
	if err := s.userRepo.Update(eventOwner); err != nil {
		fmt.Printf("Warning: Failed to update event owner photo limit: %v\n", err)
	}

	// 3. Event'in fotoğraf sayısını artır
	event.PhotoCount++
	if err := s.eventRepo.Update(event); err != nil {
		fmt.Printf("Warning: Failed to update event photo count: %v\n", err)
	}

	return response, nil
}

func (s *PhotoService) GetEventPhotos(eventID uint, userID uint) ([]models.Photos, error) {
	// Önce event'in var olup olmadığını kontrol et
	_, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return nil, errors.New("event not found")
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

	// Cloudflare Images'dan sil
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
