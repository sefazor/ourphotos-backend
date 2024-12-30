package service

import (
	"errors"
	"fmt"
	"math/rand"
	"mime/multipart"
	"time"

	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/repository"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

type EventService struct {
	eventRepo    *repository.EventRepository
	userRepo     *repository.UserRepository
	photoService *PhotoService
}

func NewEventService(
	eventRepo *repository.EventRepository,
	userRepo *repository.UserRepository,
	photoService *PhotoService,
) *EventService {
	return &EventService{
		eventRepo:    eventRepo,
		userRepo:     userRepo,
		photoService: photoService,
	}
}

func generateEventURL() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	length := 6
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
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

	// Kullanıcının tüm eventlerindeki photo_limit toplamını hesapla
	totalAllocatedPhotos, err := s.eventRepo.GetTotalPhotoLimitsByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Kalan limiti hesapla
	remainingLimit := user.PhotoLimit - totalAllocatedPhotos

	if req.PhotoLimit > remainingLimit {
		return nil, fmt.Errorf(
			"photo limit allocation exceeded. Available: %d, Requested: %d",
			remainingLimit,
			req.PhotoLimit,
		)
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

	// Event modelini oluştur
	event := &models.Event{
		UserID:            userID,
		Title:             req.Title,
		Description:       req.Description,
		URL:               url,
		IsPublic:          req.IsPublic,
		HasPassword:       req.HasPassword,
		Password:          req.Password,
		AllowGuestUploads: req.AllowGuestUpload,
		PhotoLimit:        req.PhotoLimit,
		ExpiresAt:         *req.ExpiresAt,
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

	// Response oluştur
	response := &models.EventResponse{
		ID:                      createdEvent.ID,
		Title:                   createdEvent.Title,
		Description:             createdEvent.Description,
		URL:                     createdEvent.URL,
		IsPublic:                createdEvent.IsPublic,
		HasPassword:             createdEvent.HasPassword,
		AllowGuestUploads:       createdEvent.AllowGuestUploads,
		PhotoLimit:              createdEvent.PhotoLimit,
		PhotoCount:              createdEvent.PhotoCount,
		ExpiresAt:               createdEvent.ExpiresAt,
		CreatedAt:               createdEvent.CreatedAt,
		UpdatedAt:               createdEvent.UpdatedAt,
		RemainingUserPhotoLimit: remainingLimit - req.PhotoLimit,
		TotalAllocatedPhotos:    totalAllocatedPhotos + req.PhotoLimit,
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

	// Production'da hash karşılaştırması yapılmalı
	if event.Password != password {
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

	totalAllocatedPhotos, err := s.eventRepo.GetTotalPhotoLimitsByUserID(userID)
	if err != nil {
		return nil, err
	}

	remainingLimit := user.PhotoLimit - totalAllocatedPhotos

	var response []models.EventResponse
	for _, event := range events {
		response = append(response, models.EventResponse{
			ID:                      event.ID,
			Title:                   event.Title,
			Description:             event.Description,
			URL:                     event.URL,
			IsPublic:                event.IsPublic,
			HasPassword:             event.HasPassword,
			AllowGuestUploads:       event.AllowGuestUploads,
			PhotoLimit:              event.PhotoLimit,
			PhotoCount:              event.PhotoCount,
			ExpiresAt:               event.ExpiresAt,
			CreatedAt:               event.CreatedAt,
			UpdatedAt:               event.UpdatedAt,
			RemainingUserPhotoLimit: remainingLimit,
			TotalAllocatedPhotos:    totalAllocatedPhotos,
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
	if req.ExpiresAt != nil {
		event.ExpiresAt = *req.ExpiresAt
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
			event.Password = *req.Password
		}
		updated = true
	}
	if req.AllowGuestUploads != nil {
		event.AllowGuestUploads = *req.AllowGuestUploads
		updated = true
	}
	if req.PhotoLimit != nil {
		event.PhotoLimit = *req.PhotoLimit
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
