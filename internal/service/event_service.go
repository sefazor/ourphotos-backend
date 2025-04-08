package service

import (
	cryptorand "crypto/rand"
	"errors"
	"fmt"
	"math/rand"
	"mime/multipart"
	"time"

	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/repository"
	"github.com/sefazor/ourphotos-backend/pkg/bcrypt"
	"github.com/sefazor/ourphotos-backend/pkg/qrcode"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

type EventService struct {
	eventRepo    *repository.EventRepository
	userRepo     *repository.UserRepository
	photoService *PhotoService
	qrService    *qrcode.QRService
}

func NewEventService(
	eventRepo *repository.EventRepository,
	userRepo *repository.UserRepository,
	photoService *PhotoService,
	qrService *qrcode.QRService,
) *EventService {
	return &EventService{
		eventRepo:    eventRepo,
		userRepo:     userRepo,
		photoService: photoService,
		qrService:    qrService,
	}
}

func generateEventURL() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	length := 8

	// Rastgele bayt dizisi oluştur
	randomBytes := make([]byte, length)
	_, err := cryptorand.Read(randomBytes)
	if err != nil {
		// Hata durumunda fallback olarak eski yöntemi kullan
		b := make([]byte, length)
		for i := range b {
			b[i] = charset[r.Intn(len(charset))]
		}
		return string(b)
	}

	// Baytları karakter setine dönüştür
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		// Rastgele baytları charset'in indekslerine dönüştür
		result[i] = charset[randomBytes[i]%byte(len(charset))]
	}

	return string(result)
}

// Süreye göre son geçerlilik tarihini hesaplayan yardımcı fonksiyon
func calculateExpiryDate(duration models.DurationType) time.Time {
	now := time.Now()

	switch duration {
	case models.Duration7Days:
		return now.AddDate(0, 0, 7)
	case models.Duration14Days:
		return now.AddDate(0, 0, 14)
	case models.Duration21Days:
		return now.AddDate(0, 0, 21)
	case models.Duration30Days:
		return now.AddDate(0, 0, 30)
	case models.Duration3Months:
		return now.AddDate(0, 3, 0)
	default:
		// Varsayılan olarak 7 gün
		return now.AddDate(0, 0, 7)
	}
}

func (s *EventService) CreateEvent(userID uint, req models.EventRequest) (*models.EventResponse, error) {
	// Kullanıcıyı getir
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Event limit kontrolü
	if user.EventLimit <= 0 {
		return nil, errors.New("event limit exceeded")
	}

	// URL generate et
	url := generateEventURL()

	// URL unique olana kadar dene
	for {
		exists, _ := s.eventRepo.URLExists(url)
		if !exists {
			break
		}
		url = generateEventURL()
	}

	// Şifre varsa hashle
	var hashedPassword string
	if req.HasPassword && req.Password != "" {
		var err error
		hashedPassword, err = bcrypt.HashPassword(req.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
	}

	// Süreye göre son geçerlilik tarihini hesapla
	expiresAt := calculateExpiryDate(req.Duration)

	// Event modelini oluştur
	event := &models.Event{
		UserID:            userID,
		Title:             req.Title,
		Description:       req.Description,
		Location:          req.Location,
		URL:               url,
		IsPublic:          req.IsPublic,
		HasPassword:       req.HasPassword,
		Password:          hashedPassword,
		AllowGuestUploads: req.AllowGuestUploads,
		ExpiresAt:         expiresAt,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Eventi oluştur
	createdEvent, err := s.eventRepo.Create(event)
	if err != nil {
		return nil, err
	}

	// Event başarıyla oluşturulduysa limiti düş
	user.EventLimit--
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	// Kullanıcının kalan fotoğraf limitini hesapla
	remainingPhotoLimit := user.PhotoLimit

	// Response oluştur
	response := &models.EventResponse{
		ID:                      createdEvent.ID,
		Title:                   createdEvent.Title,
		Description:             createdEvent.Description,
		Location:                createdEvent.Location,
		URL:                     createdEvent.URL,
		IsPublic:                createdEvent.IsPublic,
		HasPassword:             createdEvent.HasPassword,
		AllowGuestUploads:       createdEvent.AllowGuestUploads,
		PhotoCount:              createdEvent.PhotoCount,
		ExpiresAt:               createdEvent.ExpiresAt,
		Duration:                string(req.Duration),
		CreatedAt:               createdEvent.CreatedAt,
		UpdatedAt:               createdEvent.UpdatedAt,
		RemainingUserPhotoLimit: remainingPhotoLimit,
		TotalAllocatedPhotos:    0,
	}

	return response, nil
}

func (s *EventService) CheckEventPassword(eventURL string, password string) error {
	event, err := s.eventRepo.GetByURL(eventURL)
	if err != nil {
		return err
	}

	if !event.HasPassword {
		return nil
	}

	// Hash karşılaştırması yap
	if err := bcrypt.ComparePassword(event.Password, password); err != nil {
		return errors.New("incorrect password")
	}

	return nil
}

func (s *EventService) GetEvent(eventID uint) (*models.Event, error) {
	return s.eventRepo.GetByID(eventID)
}

func (s *EventService) GetUserEvents(userID uint) ([]models.EventResponse, error) {
	events, err := s.eventRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	remainingLimit := user.PhotoLimit

	var response []models.EventResponse
	for _, event := range events {
		response = append(response, models.EventResponse{
			ID:                      event.ID,
			Title:                   event.Title,
			Description:             event.Description,
			Location:                event.Location,
			URL:                     event.URL,
			IsPublic:                event.IsPublic,
			HasPassword:             event.HasPassword,
			AllowGuestUploads:       event.AllowGuestUploads,
			PhotoCount:              event.PhotoCount,
			ExpiresAt:               event.ExpiresAt,
			CreatedAt:               event.CreatedAt,
			UpdatedAt:               event.UpdatedAt,
			RemainingUserPhotoLimit: remainingLimit,
			TotalAllocatedPhotos:    0,
		})
	}

	return response, nil
}

func (s *EventService) UpdateEvent(eventID uint, userID uint, req models.UpdateEventRequest) (*models.Event, error) {
	// Önce eventi bul
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return nil, err
	}

	// Yetki kontrolü
	if event.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	// Değişiklikleri kontrol et
	updated := false

	if req.Title != nil {
		event.Title = *req.Title
		updated = true
	}
	if req.Description != nil {
		event.Description = *req.Description
		updated = true
	}
	if req.Location != nil {
		event.Location = *req.Location
		updated = true
	}
	if req.Duration != nil {
		// Süreye göre yeni son geçerlilik tarihini hesapla
		event.ExpiresAt = calculateExpiryDate(*req.Duration)
		updated = true
	}
	if req.IsPublic != nil {
		event.IsPublic = *req.IsPublic
		updated = true
	}
	if req.HasPassword != nil {
		event.HasPassword = *req.HasPassword
		// Şifre durumunu güncelle
		if !event.HasPassword {
			event.Password = ""
		} else if req.Password != nil && *req.Password != "" {
			// Şifreyi hashle
			hashedPassword, err := bcrypt.HashPassword(*req.Password)
			if err != nil {
				return nil, fmt.Errorf("failed to hash password: %w", err)
			}
			event.Password = hashedPassword
		}
		updated = true
	}
	if req.AllowGuestUploads != nil {
		event.AllowGuestUploads = *req.AllowGuestUploads
		updated = true
	}

	// Değişiklik yoksa güncelleme yapma
	if !updated {
		return event, nil
	}

	// Eventi güncelle
	if err := s.eventRepo.Update(event); err != nil {
		return nil, err
	}

	return event, nil
}

func (s *EventService) DeleteEvent(eventID uint, userID uint) error {
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return err
	}

	// Event sahibi kontrolü
	if event.UserID != userID {
		return errors.New("unauthorized")
	}

	// Önce etkinliğe ait fotoğrafları bul
	photos, err := s.photoService.photoRepo.GetByEventID(eventID)
	if err != nil {
		return fmt.Errorf("failed to retrieve event photos: %w", err)
	}

	// Her bir fotoğrafı depolama servislerinden sil
	for _, photo := range photos {
		// Cloudflare Images'dan sil
		if err := s.photoService.ImgStorage.Delete(photo.ImageID); err != nil {
			// Hata loga kaydedilmeli ama işleme devam edilmeli
			fmt.Printf("Error deleting photo %s from Cloudflare Images: %v\n", photo.ImageID, err)
		}
	}

	// Veritabanından tüm fotoğrafları sil
	if err := s.photoService.photoRepo.DeleteByEventID(eventID); err != nil {
		return fmt.Errorf("failed to delete event photos from database: %w", err)
	}

	// Event'i sil
	if err := s.eventRepo.Delete(eventID); err != nil {
		return err
	}

	// Kullanıcının event limitini geri ver
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	user.EventLimit++
	return s.userRepo.Update(user)
}

func (s *EventService) GetEventByURL(url string) (*models.Event, error) {
	return s.eventRepo.GetByURL(url)
}

func (s *EventService) GetEventPhotoCount(eventID uint) (int64, error) {
	return s.eventRepo.GetPhotoCount(eventID)
}

func (s *EventService) UploadEventPhoto(eventID uint, userID uint, file *multipart.FileHeader) (*models.PhotoResponse, error) {
	// Event kontrolü
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return nil, err
	}

	// Yetki kontrolü
	if event.UserID != userID && !event.AllowGuestUploads {
		return nil, errors.New("unauthorized")
	}

	// Fotoğraf yükleme işlemi
	// Bu kısmı PhotoService'e delege edebiliriz
	return s.photoService.UploadPhoto(eventID, userID, file)
}

// Süresi dolmuş etkinlikleri temizleme metodu
func (s *EventService) CleanupExpiredEvents() error {
	// Şu anki tarihten önceki ExpiresAt değerine sahip eventleri bul
	expiredEvents, err := s.eventRepo.FindExpiredEvents(time.Now())
	if err != nil {
		return fmt.Errorf("failed to find expired events: %w", err)
	}

	if len(expiredEvents) == 0 {
		return nil // Silinecek etkinlik yok
	}

	fmt.Printf("Found %d expired events to clean up\n", len(expiredEvents))

	// Her etkinlik için silme işlemini gerçekleştir
	for _, event := range expiredEvents {
		// Önce ilişkili fotoğrafları sil
		photos, err := s.photoService.photoRepo.GetByEventID(event.ID)
		if err != nil {
			fmt.Printf("Error retrieving photos for event %d: %v\n", event.ID, err)
			continue
		}

		// Her bir fotoğrafı sil
		for _, photo := range photos {
			// Cloudflare Images'dan sil
			if err := s.photoService.ImgStorage.Delete(photo.ImageID); err != nil {
				fmt.Printf("Error deleting photo %s from Cloudflare Images: %v\n", photo.ImageID, err)
			}
		}

		// Veritabanından tüm fotoğrafları sil
		if err := s.photoService.photoRepo.DeleteByEventID(event.ID); err != nil {
			fmt.Printf("Error deleting photos for event %d: %v\n", event.ID, err)
		}

		// Etkinliği sil
		if err := s.eventRepo.Delete(event.ID); err != nil {
			fmt.Printf("Error deleting event %d: %v\n", event.ID, err)
			continue
		}

		fmt.Printf("Successfully deleted expired event %d (%s) with %d photos\n",
			event.ID, event.Title, len(photos))

		// Event sahibinin event limitini geri ver
		user, err := s.userRepo.GetByID(event.UserID)
		if err == nil { // Kullanıcı hala mevcutsa
			user.EventLimit++
			if err := s.userRepo.Update(user); err != nil {
				fmt.Printf("Error returning event limit to user %d: %v\n", user.ID, err)
			}
		}
	}

	return nil
}

// GetEventQRCode, belirtilen etkinlik için QR kod PNG formatında döndürür
func (s *EventService) GetEventQRCode(eventID uint, size int) ([]byte, error) {
	// Etkinliği getir
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	// QR kodu oluştur
	qrCode, err := s.qrService.GenerateQRCode(event.URL, size)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	return qrCode, nil
}
