package service

import (
	"errors"
	"time"

	"os"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/repository"
	"github.com/sefazor/ourphotos-backend/pkg/bcrypt"
	"github.com/sefazor/ourphotos-backend/pkg/email"
)

type UserService struct {
	userRepo     *repository.UserRepository
	emailService *email.EmailService
}

func NewUserService(userRepo *repository.UserRepository, emailService *email.EmailService) *UserService {
	return &UserService{
		userRepo:     userRepo,
		emailService: emailService,
	}
}

func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	return s.userRepo.GetByEmail(email)
}

func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *UserService) ChangePassword(userID uint, req models.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	if err := bcrypt.ComparePassword(user.Password, req.CurrentPassword); err != nil {
		return errors.New("current password is incorrect")
	}

	hashedPassword, err := bcrypt.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(user.ID, hashedPassword)
}

func (s *UserService) InitiateEmailChange(userID uint, req models.ChangeEmailRequest) error {
	// Kullanıcıyı bul
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// Şifreyi doğrula
	if err := bcrypt.ComparePassword(user.Password, req.Password); err != nil {
		return errors.New("invalid password")
	}

	// Yeni email'in kullanımda olup olmadığını kontrol et
	exists, err := s.userRepo.EmailExists(req.NewEmail)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("email already in use")
	}

	// Email değişikliği için token oluştur
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   userID,
		"new_email": req.NewEmail,
		"exp":       time.Now().Add(TokenExpiryEmailChange).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return err
	}

	// Doğrulama emaili gönder
	return s.emailService.SendEmailChangeVerification(req.NewEmail, tokenString)
}

func (s *UserService) CompleteEmailChange(token string) error {
	// Token'ı doğrula
	claims := jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !parsedToken.Valid {
		return errors.New("invalid or expired token")
	}

	// Token'dan bilgileri al
	userID := uint(claims["user_id"].(float64))
	newEmail := claims["new_email"].(string)

	// Email'i güncelle
	return s.userRepo.UpdateEmail(userID, newEmail)
}

func (s *UserService) UpdateProfile(userID uint, req models.UpdateProfileRequest) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	user.FullName = req.FullName

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}
