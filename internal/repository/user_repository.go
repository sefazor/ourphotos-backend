package repository

import (
	"errors"
	"fmt"

	"github.com/sefazor/ourphotos-backend/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) EmailExists(email string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) UpdatePassword(userID uint, hashedPassword string) error {
	result := r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("password", hashedPassword)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("no rows affected")
	}

	// Debug için ekliyoruz
	var user models.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return err
	}
	fmt.Printf("Password updated in DB. New hash: %s\n", user.Password)

	return nil
}

func (r *UserRepository) GetByID(userID uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) UpdateEmail(userID uint, newEmail string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("email", newEmail).Error
}
