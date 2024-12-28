package repository

import (
	"github.com/sefazor/ourphotos-backend/internal/models"
	"gorm.io/gorm"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(event *models.Event) (*models.Event, error) {
	result := r.db.Create(event)
	if result.Error != nil {
		return nil, result.Error
	}
	return event, nil
}

func (r *EventRepository) GetByID(id uint) (*models.Event, error) {
	var event models.Event
	err := r.db.First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *EventRepository) GetUserEvents(userID uint) ([]models.Event, error) {
	var events []models.Event
	err := r.db.Where("user_id = ?", userID).Find(&events).Error
	return events, err
}

func (r *EventRepository) Update(event *models.Event) error {
	return r.db.Save(event).Error
}

func (r *EventRepository) Delete(id uint) error {
	return r.db.Delete(&models.Event{}, id).Error
}

func (r *EventRepository) GetByURL(url string) (*models.Event, error) {
	var event models.Event
	err := r.db.Where("url = ?", url).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *EventRepository) URLExists(url string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Event{}).Where("url = ?", url).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}