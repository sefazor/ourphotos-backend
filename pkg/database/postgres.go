package database

import (
	"log"
	"os"

	"github.com/sefazor/ourphotos-backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Global DB değişkeni
var DB *gorm.DB

func NewDatabase() *gorm.DB {
	// Doğrudan DATABASE_URL'i kullan
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	var err error
	// Global DB değişkenini başlat
	DB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto Migrate
	err = DB.AutoMigrate(
		&models.User{},
		&models.Event{},
		&models.Photos{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	return DB
}
