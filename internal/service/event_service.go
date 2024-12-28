package service

import (
	"errors"

	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/repository"
	"github.com/sefazor/ourphotos-backend/pkg/utils"
)

type EventService struct {
	eventRepo *repository.EventRepository
	userRepo  *repository.UserRepository
}

func NewEventService(eventRepo *repository.EventRepository, userRepo *repository.UserRepository) *EventService {
	return &EventService{
		eventRepo: eventRepo,
		userRepo:  userRepo,
	}
}

func (s *EventService) CreateEvent(userID uint, req models.EventRequest) (*models.Event, error) {
	// URL oluştur
	url := utils.GenerateRandomString(10)

	event := &models.Event{
		UserID:            userID,
		Title:             req.Title,
		Description:       req.Description,
		URL:               url,
		IsPublic:          req.IsPublic,
		AllowGuestUploads: req.AllowGuestUploads,
		PhotoLimit:        req.PhotoLimit,
		StartDate:         req.StartDate,
		EndDate:           req.EndDate,
	}

	// Şifre varsa ekle
	if req.Password != "" {
		event.HasPassword = true
		event.Password = req.Password
	}

	// Create metodunun dönüş değerlerini doğru şekilde kullan
	createdEvent, err := s.eventRepo.Create(event)
	if err != nil {
		return nil, err
	}

	return createdEvent, nil
}

func (s *EventService) CheckEventPassword(eventID uint, password string) error {
	event, err := s.eventRepo.GetByID(eventID)
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

func (s *EventService) GetUserEvents(userID uint) ([]models.Event, error) {
	return s.eventRepo.GetUserEvents(userID)
}

func (s *EventService) UpdateEvent(eventID uint, userID uint, req models.UpdateEventRequest) (*models.Event, error) {
	// Event'i bul
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return nil, err
	}

	// Yetki kontrolü
	if event.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	// Event bilgilerini güncelle
	event.Title = req.Title
	event.Description = req.Description
	event.StartDate = req.StartDate
	event.EndDate = req.EndDate
	event.IsPublic = req.IsPublic
	event.AllowGuestUploads = req.AllowGuestUploads
	event.PhotoLimit = req.PhotoLimit

	// İsim değiştiyse URL'i güncelle
	if event.Title != req.Title {
		// URL oluştur
		url := utils.GenerateRandomString(10)
		event.URL = url
	}

	// Veritabanında güncelle
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
