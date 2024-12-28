package models

import (
	"time"
)

type User struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	FullName   string    `json:"full_name" gorm:"not null"`
	Email      string    `json:"email" gorm:"unique;not null"`
	Password   string    `json:"-" gorm:"not null"`
	EventLimit int       `json:"event_limit" gorm:"not null;default:1"`
	PhotoLimit int       `json:"photo_limit" gorm:"not null;default:20"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
