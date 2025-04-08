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
			Name:        "100 Credits",
			Description: "100 image uploads, Unlimited events, 3 months hosting",
			Credits:     100,
			EventLimit:  999, // Unlimited events
			PhotoLimit:  100,
			Price:       19.99,
			IsActive:    true,
		},
		{
			Name:        "300 Credits",
			Description: "300 image uploads, Unlimited events, 3 months hosting",
			Credits:     300,
			EventLimit:  999, // Unlimited events
			PhotoLimit:  300,
			Price:       39.99,
			IsActive:    true,
		},
		{
			Name:        "500 Credits",
			Description: "500 image uploads, Unlimited events, 3 months hosting",
			Credits:     500,
			EventLimit:  999, // Unlimited events
			PhotoLimit:  500,
			Price:       59.99,
			IsActive:    true,
		},
		{
			Name:        "1500 Credits",
			Description: "1500 image uploads, Unlimited events, 3 months hosting, Priority support",
			Credits:     1500,
			EventLimit:  999, // Unlimited events
			PhotoLimit:  1500,
			Price:       149.99,
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
			Name:        "100 Credits",
			Description: "100 image uploads, Unlimited events, 3 months hosting",
			Credits:     100,
			EventLimit:  999, // Unlimited events
			PhotoLimit:  100,
			Price:       19.99,
			IsActive:    true,
		},
		{
			Name:        "300 Credits",
			Description: "300 image uploads, Unlimited events, 3 months hosting",
			Credits:     300,
			EventLimit:  999, // Unlimited events
			PhotoLimit:  300,
			Price:       39.99,
			IsActive:    true,
		},
		{
			Name:        "500 Credits",
			Description: "500 image uploads, Unlimited events, 3 months hosting",
			Credits:     500,
			EventLimit:  999, // Unlimited events
			PhotoLimit:  500,
			Price:       59.99,
			IsActive:    true,
		},
		{
			Name:        "1500 Credits",
			Description: "1500 image uploads, Unlimited events, 3 months hosting, Priority support",
			Credits:     1500,
			EventLimit:  999, // Unlimited events
			PhotoLimit:  1500,
			Price:       149.99,
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
