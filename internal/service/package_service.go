package service

import (
	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/repository"
)

type PackageService struct {
	packageRepo *repository.CreditPackageRepository
}

func NewPackageService(packageRepo *repository.CreditPackageRepository) *PackageService {
	return &PackageService{
		packageRepo: packageRepo,
	}
}

func (s *PackageService) GetAllPackages() ([]models.CreditPackage, error) {
	return s.packageRepo.GetAll()
}

func (s *PackageService) GetPackageByID(id uint) (*models.CreditPackage, error) {
	return s.packageRepo.GetByID(id)
}
