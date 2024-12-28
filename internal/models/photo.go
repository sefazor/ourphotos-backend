package models

import (
	"time"
)

type Photos struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	EventID    uint      `json:"event_id" gorm:"not null"`
	UserID     uint      `json:"user_id"`
	FileName   string    `json:"file_name" gorm:"not null"`
	FileSize   int64     `json:"file_size" gorm:"not null"`
	MimeType   string    `json:"mime_type" gorm:"not null"`
	Path       string    `json:"path" gorm:"not null"`
	ImageID    string    `json:"image_id" gorm:"not null"`
	Variants   []string  `json:"variants" gorm:"type:json;serializer:json"`
	R2Key      string    `json:"r2_key" gorm:"not null"`
	IsGuest    bool      `json:"is_guest" gorm:"default:false"`
	IsUploaded bool      `json:"is_uploaded" gorm:"default:false"`
	UploadedAt time.Time `json:"uploaded_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreatePhotoRequest struct {
	EventID uint   `json:"event_id" validate:"required"`
	File    []byte `json:"-" validate:"required"` // Form-data'dan gelecek
}

type PhotoResponse struct {
	ID        uint      `json:"id"`
	EventID   uint      `json:"event_id"`
	UserID    uint      `json:"user_id,omitempty"`
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
	MimeType  string    `json:"mime_type"`
	Path      string    `json:"path"`
	IsGuest   bool      `json:"is_guest"`
	CreatedAt time.Time `json:"created_at"`
}
