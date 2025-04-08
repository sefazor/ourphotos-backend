package repository

import (
	"github.com/sefazor/ourphotos-backend/internal/models"
	"gorm.io/gorm"
)

type PhotoRepository struct {
	db *gorm.DB
}

func NewPhotoRepository(db *gorm.DB) *PhotoRepository {
	return &PhotoRepository{
		db: db,
	}
}

func (r *PhotoRepository) Create(photo *models.Photos) error {
	return r.db.Create(photo).Error
}

func (r *PhotoRepository) GetByID(id uint) (*models.Photos, error) {
	var photo models.Photos
	err := r.db.First(&photo, id).Error
	if err != nil {
		return nil, err
	}
	return &photo, nil
}

func (r *PhotoRepository) GetByEventID(eventID uint) ([]models.Photos, error) {
	var photos []models.Photos
	err := r.db.Where("event_id = ?", eventID).
		Order("created_at DESC").
		Find(&photos).Error
	return photos, err
}

func (r *PhotoRepository) Delete(id uint) error {
	return r.db.Delete(&models.Photos{}, id).Error
}

func (r *PhotoRepository) CountByEventID(eventID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Photos{}).Where("event_id = ?", eventID).Count(&count).Error
	return count, err
}

func (r *PhotoRepository) GetEventPhotoCount(eventID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Photos{}).Where("event_id = ?", eventID).Count(&count).Error
	return count, err
}

func (r *PhotoRepository) DeleteByEventID(eventID uint) error {
	return r.db.Where("event_id = ?", eventID).Delete(&models.Photos{}).Error
}
