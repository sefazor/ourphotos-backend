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

func (r *CreditPackageRepository) Create(pkg *models.CreditPackage) error {
	return r.db.Create(pkg).Error
}

func (r *CreditPackageRepository) Update(pkg *models.CreditPackage) error {
	return r.db.Save(pkg).Error
}

func (r *CreditPackageRepository) Delete(id uint) error {
	return r.db.Delete(&models.CreditPackage{}, id).Error
}
