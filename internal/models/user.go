package models

import (
	"time"
)

type User struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	FullName   string    `json:"full_name" gorm:"not null"`
	Email      string    `json:"email" gorm:"unique;not null"`
	Password   string    `json:"-" gorm:"not null"`
	EventLimit int       `json:"event_limit" gorm:"default:1"`
	PhotoLimit int       `json:"photo_limit" gorm:"default:20"`
	IsVerified bool      `json:"is_verified" gorm:"default:false"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
