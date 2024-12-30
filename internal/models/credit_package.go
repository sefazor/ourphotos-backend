package models

import "time"

type CreditPackage struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Credits     int       `json:"credits" gorm:"not null"`
	EventLimit  int       `json:"event_limit" gorm:"not null"`
	PhotoLimit  int       `json:"photo_limit" gorm:"not null"`
	Price       float64   `json:"price" gorm:"not null"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Kullanıcının satın aldığı paketleri takip etmek için
type UserCreditPurchase struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	UserID          uint      `json:"user_id" gorm:"not null"`
	PackageID       uint      `json:"package_id" gorm:"not null"`
	EventLimit      int       `json:"event_limit" gorm:"not null"`
	PhotoLimit      int       `json:"photo_limit" gorm:"not null"`
	Price           float64   `json:"price" gorm:"not null"`
	StripeSessionID string    `json:"stripe_session_id" gorm:"unique;not null"`
	Status          string    `json:"status" gorm:"not null;default:'pending'"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
