package models

import (
	"time"
)

type Event struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	UserID            uint      `json:"user_id" gorm:"not null"`
	Title             string    `json:"title" gorm:"not null"`
	Description       string    `json:"description"`
	URL               string    `json:"url" gorm:"unique;not null"`
	IsPublic          bool      `json:"is_public" gorm:"default:true"`
	HasPassword       bool      `json:"has_password" gorm:"default:false"`
	Password          string    `json:"-" gorm:"type:varchar(255)"`
	AllowGuestUploads bool      `json:"allow_guest_uploads" gorm:"default:true"`
	PhotoLimit        int       `json:"photo_limit" gorm:"default:0"`
	ExpiresAt         time.Time `json:"expires_at"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	PhotoCount        int       `json:"photo_count" gorm:"default:0"`
}

type EventPasswordRequest struct {
	Password string `json:"password" validate:"required"`
}

type EventRequest struct {
	Title            string     `json:"title" validate:"required"`
	Description      string     `json:"description"`
	HasPassword      bool       `json:"has_password"`
	Password         string     `json:"password"`
	IsPublic         bool       `json:"is_public"`
	AllowGuestUpload bool       `json:"allow_guest_uploads"`
	PhotoLimit       int        `json:"photo_limit" validate:"required"`
	ExpiresAt        *time.Time `json:"expires_at"`
}

type UpdateEventRequest struct {
	Title             *string    `json:"title"`
	Description       *string    `json:"description"`
	URL               *string    `json:"url"`
	HasPassword       *bool      `json:"has_password"`
	Password          *string    `json:"password"`
	IsPublic          *bool      `json:"is_public"`
	AllowGuestUploads *bool      `json:"allow_guest_uploads"`
	PhotoLimit        *int       `json:"photo_limit"`
	ExpiresAt         *time.Time `json:"expires_at"`
}

type EventResponse struct {
	ID                      uint      `json:"id"`
	Title                   string    `json:"title"`
	Description             string    `json:"description"`
	URL                     string    `json:"url"`
	IsPublic                bool      `json:"is_public"`
	HasPassword             bool      `json:"has_password"`
	AllowGuestUploads       bool      `json:"allow_guest_uploads"`
	PhotoLimit              int       `json:"photo_limit"`
	PhotoCount              int       `json:"photo_count"`
	ExpiresAt               time.Time `json:"expires_at"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
	RemainingUserPhotoLimit int       `json:"remaining_user_photo_limit"`
	TotalAllocatedPhotos    int       `json:"total_allocated_photos"`
}
