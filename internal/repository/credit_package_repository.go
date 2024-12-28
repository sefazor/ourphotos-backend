package repository

import (
	"github.com/sefazor/ourphotos-backend/internal/models"
	"gorm.io/gorm"
)

type CreditPackageRepository struct {
	db *gorm.DB
}

func NewCreditPackageRepository(db *gorm.DB) *CreditPackageRepository {
	return &CreditPackageRepository{
		db: db,
	}
}

func (r *CreditPackageRepository) GetByID(id uint) (*models.CreditPackage, error) {
	var creditPackage models.CreditPackage
	err := r.db.First(&creditPackage, id).Error
	return &creditPackage, err
}

func (r *CreditPackageRepository) GetAll() ([]models.CreditPackage, error) {
	var packages []models.CreditPackage
	err := r.db.Where("is_active = ?", true).Find(&packages).Error
	return packages, err
}
