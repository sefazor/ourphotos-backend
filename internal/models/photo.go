package models

import (
	"time"
)

type Photos struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	EventID    uint      `json:"event_id"`
	UserID     uint      `json:"user_id"`
	FileName   string    `json:"file_name"`
	FileSize   int64     `json:"file_size"`
	MimeType   string    `json:"mime_type"`
	R2Key      string    `json:"r2_key"`
	ImageID    string    `json:"image_id"`
	PublicURL  string    `json:"public_url"`
	IsGuest    bool      `json:"is_guest"`
	UploadedAt time.Time `json:"uploaded_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreatePhotoRequest struct {
	EventID uint   `json:"event_id" validate:"required"`
	File    []byte `json:"-" validate:"required"` // Form-data'dan gelecek
}

type PhotoResponse struct {
	ID           uint      `json:"id"`
	EventID      uint      `json:"event_id"`
	UserID       uint      `json:"user_id,omitempty"`
	FileName     string    `json:"file_name"`
	FileSize     int64     `json:"file_size"`
	MimeType     string    `json:"mime_type"`
	PublicURL    string    `json:"public_url"`
	ThumbnailURL string    `json:"thumbnail_url"`
	MediumURL    string    `json:"medium_url"`
	LargeURL     string    `json:"large_url"`
	IsGuest      bool      `json:"is_guest"`
	CreatedAt    time.Time `json:"created_at"`
}
