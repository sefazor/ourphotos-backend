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
	StartDate         time.Time `json:"start_date"`
	EndDate           time.Time `json:"end_date"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type EventPasswordRequest struct {
	Password string `json:"password" validate:"required"`
}

type EventRequest struct {
	Title             string    `json:"title" validate:"required,min=3,max=100"`
	Description       string    `json:"description" validate:"max=500"`
	StartDate         time.Time `json:"start_date" validate:"required,gt=now" time_format:"2006-01-02T15:04:05.000Z"`
	EndDate           time.Time `json:"end_date" validate:"required,gtfield=StartDate" time_format:"2006-01-02T15:04:05.000Z"`
	IsPublic          bool      `json:"is_public"`
	Password          string    `json:"password,omitempty"`
	AllowGuestUploads bool      `json:"allow_guest_uploads"`
	PhotoLimit        int       `json:"photo_limit" validate:"required,min=1,max=1000"`
}

type UpdateEventRequest struct {
	Title             string    `json:"title" validate:"required"`
	Description       string    `json:"description"`
	StartDate         time.Time `json:"start_date" validate:"required"`
	EndDate           time.Time `json:"end_date" validate:"required,gtfield=StartDate"`
	IsPublic          bool      `json:"is_public"`
	AllowGuestUploads bool      `json:"allow_guest_uploads"`
	PhotoLimit        int       `json:"photo_limit" validate:"required,min=1"`
}
