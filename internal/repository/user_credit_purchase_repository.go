package repository

import (
	"github.com/sefazor/ourphotos-backend/internal/models"
	"gorm.io/gorm"
)

type UserCreditPurchaseRepository struct {
	db *gorm.DB
}

func NewUserCreditPurchaseRepository(db *gorm.DB) *UserCreditPurchaseRepository {
	return &UserCreditPurchaseRepository{
		db: db,
	}
}

func (r *UserCreditPurchaseRepository) Create(purchase *models.UserCreditPurchase) error {
	return r.db.Create(purchase).Error
}

func (r *UserCreditPurchaseRepository) GetBySessionID(sessionID string) (*models.UserCreditPurchase, error) {
	var purchase models.UserCreditPurchase
	err := r.db.Where("stripe_session_id = ?", sessionID).First(&purchase).Error
	return &purchase, err
}

func (r *UserCreditPurchaseRepository) Update(purchase *models.UserCreditPurchase) error {
	return r.db.Save(purchase).Error
}

func (r *UserCreditPurchaseRepository) GetUserPurchaseHistory(userID uint) ([]models.UserCreditPurchase, error) {
	var purchases []models.UserCreditPurchase
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&purchases).Error
	return purchases, err
}
