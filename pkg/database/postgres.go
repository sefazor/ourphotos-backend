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
		&models.CreditPackage{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Örnek paketleri ekle
	packages := []models.CreditPackage{
		{
			Name:        "Mini Package",
			Description: "Perfect for small events",
			Credits:     250,
			EventLimit:  1,
			PhotoLimit:  250,
			Price:       7.99,
			IsActive:    true,
		},
		{
			Name:        "Pro Package",
			Description: "Ideal for medium-sized events",
			Credits:     1000,
			EventLimit:  3,
			PhotoLimit:  1000,
			Price:       13.99,
			IsActive:    true,
		},
		{
			Name:        "Enterprise Package",
			Description: "Best for large events",
			Credits:     2000,
			EventLimit:  5,
			PhotoLimit:  2000,
			Price:       49.99,
			IsActive:    true,
		},
	}

	// Paketleri veritabanına ekle (eğer yoksa)
	for _, pkg := range packages {
		var count int64
		DB.Model(&models.CreditPackage{}).Where("name = ?", pkg.Name).Count(&count)
		if count == 0 {
			if err := DB.Create(&pkg).Error; err != nil {
				log.Fatalf("Failed to add credit package: %v", err)
			}
		}
	}

	return DB
}

func RunMigrations(db *gorm.DB) error {
	// Mevcut migrationlar
	err := db.AutoMigrate(
		&models.User{},
		&models.Event{},
		&models.Photos{},
		&models.CreditPackage{},
	)
	if err != nil {
		return err
	}

	// Örnek paketleri ekle
	packages := []models.CreditPackage{
		{
			Name:        "Mini Package",
			Description: "Perfect for small events",
			Credits:     250,
			EventLimit:  1,
			PhotoLimit:  250,
			Price:       7.99,
			IsActive:    true,
		},
		{
			Name:        "Pro Package",
			Description: "Ideal for medium-sized events",
			Credits:     1000,
			EventLimit:  3,
			PhotoLimit:  1000,
			Price:       13.99,
			IsActive:    true,
		},
		{
			Name:        "Enterprise Package",
			Description: "Best for large events",
			Credits:     2000,
			EventLimit:  5,
			PhotoLimit:  2000,
			Price:       49.99,
			IsActive:    true,
		},
	}

	// Paketleri veritabanına ekle (eğer yoksa)
	for _, pkg := range packages {
		var count int64
		db.Model(&models.CreditPackage{}).Where("name = ?", pkg.Name).Count(&count)
		if count == 0 {
			if err := db.Create(&pkg).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
