package models

import (
	"time"
)

// Duration türleri için enum tanımı
type DurationType string

const (
	Duration7Days   DurationType = "7days"
	Duration14Days  DurationType = "14days"
	Duration21Days  DurationType = "21days"
	Duration30Days  DurationType = "30days"
	Duration3Months DurationType = "3months"
)

type Event struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	UserID            uint      `json:"user_id" gorm:"not null"`
	Title             string    `json:"title" gorm:"not null"`
	Description       string    `json:"description"`
	Location          string    `json:"location"` // Etkinlik lokasyonu
	URL               string    `json:"url" gorm:"unique;not null"`
	IsPublic          bool      `json:"is_public" gorm:"default:true"`
	HasPassword       bool      `json:"has_password" gorm:"default:false"`
	Password          string    `json:"-" gorm:"type:varchar(255)"`
	AllowGuestUploads bool      `json:"allow_guest_uploads" gorm:"default:true"`
	ExpiresAt         time.Time `json:"expires_at"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	PhotoCount        int       `json:"photo_count" gorm:"default:0"`
}

type EventPasswordRequest struct {
	Password string `json:"password" validate:"required"`
}

type EventRequest struct {
	Title             string       `json:"title" validate:"required"`
	Description       string       `json:"description"`
	Location          string       `json:"location"` // Lokasyon alanı
	HasPassword       bool         `json:"has_password"`
	Password          string       `json:"password"`
	IsPublic          bool         `json:"is_public"`
	AllowGuestUploads bool         `json:"allow_guest_uploads"`
	Duration          DurationType `json:"duration" validate:"required"` // ExpiresAt yerine Duration alanı
}

type UpdateEventRequest struct {
	Title             *string       `json:"title"`
	Description       *string       `json:"description"`
	Location          *string       `json:"location"` // Lokasyon güncelleme için alan
	URL               *string       `json:"url"`
	HasPassword       *bool         `json:"has_password"`
	Password          *string       `json:"password"`
	IsPublic          *bool         `json:"is_public"`
	AllowGuestUploads *bool         `json:"allow_guest_uploads"`
	Duration          *DurationType `json:"duration"`
}

type EventResponse struct {
	ID                      uint      `json:"id"`
	Title                   string    `json:"title"`
	Description             string    `json:"description"`
	Location                string    `json:"location"` // Lokasyon bilgisi için yanıt alanı
	URL                     string    `json:"url"`
	IsPublic                bool      `json:"is_public"`
	HasPassword             bool      `json:"has_password"`
	AllowGuestUploads       bool      `json:"allow_guest_uploads"`
	PhotoCount              int       `json:"photo_count"`
	ExpiresAt               time.Time `json:"expires_at"`
	Duration                string    `json:"duration"` // Kullanıcı dostu gösterim için
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
	RemainingUserPhotoLimit int       `json:"remaining_user_photo_limit"`
	TotalAllocatedPhotos    int       `json:"total_allocated_photos"`
}
