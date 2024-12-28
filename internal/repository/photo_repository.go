package repository

import (
	"encoding/json"

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
	// Variants'ı JSON string'e çevir
	variantsJSON, err := json.Marshal(photo.Variants)
	if err != nil {
		return err
	}

	// GORM'un raw SQL kullanmasını sağla
	return r.db.Exec(`
		INSERT INTO photos (
			event_id, user_id, file_name, file_size, mime_type, 
			path, image_id, variants, is_guest, created_at, updated_at,
			r2_key
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		photo.EventID, photo.UserID, photo.FileName, photo.FileSize,
		photo.MimeType, photo.Path, photo.ImageID, variantsJSON,
		photo.IsGuest, photo.CreatedAt, photo.UpdatedAt,
		photo.R2Key,
	).Error
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

	rows, err := r.db.Raw(`
		SELECT id, event_id, user_id, file_name, file_size, 
			   mime_type, path, image_id, variants, is_guest, 
			   created_at, updated_at, r2_key
		FROM photos 
		WHERE event_id = ? 
		ORDER BY created_at DESC`, eventID).Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var photo models.Photos
		var variantsJSON []byte

		err := rows.Scan(
			&photo.ID, &photo.EventID, &photo.UserID, &photo.FileName,
			&photo.FileSize, &photo.MimeType, &photo.Path, &photo.ImageID,
			&variantsJSON, &photo.IsGuest, &photo.CreatedAt, &photo.UpdatedAt,
			&photo.R2Key,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(variantsJSON, &photo.Variants); err != nil {
			return nil, err
		}

		photos = append(photos, photo)
	}

	return photos, nil
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
