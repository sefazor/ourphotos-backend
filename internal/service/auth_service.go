package service

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/repository"
	"github.com/sefazor/ourphotos-backend/pkg/email"
	jwtPkg "github.com/sefazor/ourphotos-backend/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo     *repository.UserRepository
	emailService *email.EmailService
	jwtSecret    []byte
	jwtIssuer    string
}

func NewAuthService(userRepo *repository.UserRepository, emailService *email.EmailService) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		emailService: emailService,
		jwtSecret:    []byte(os.Getenv("JWT_SECRET")),
		jwtIssuer:    os.Getenv("JWT_ISSUER"),
	}
}

func (s *AuthService) Register(req models.RegisterRequest) (*models.AuthResponse, error) {
	// Email kontrolü
	exists, err := s.userRepo.EmailExists(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	// Şifreyi hashle
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Kullanıcıyı oluştur
	user := &models.User{
		FullName: req.FullName,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// JWT token oluştur
	token, err := jwtPkg.GenerateToken(user.Email)
	if err != nil {
		return nil, err
	}

	// Welcome email gönder
	go s.emailService.SendWelcomeEmail(user.Email, user.FullName)

	return &models.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *AuthService) Login(req models.LoginRequest) (*models.AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate token with user_id
	token, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	// Return AuthResponse
	return &models.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *AuthService) ForgotPassword(email string) error {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil // Güvenlik için hata dönme
	}

	// Reset token oluştur
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   user.Email,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    s.jwtIssuer,
	})

	resetToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return err
	}

	// Reset email gönder
	return s.emailService.SendPasswordResetEmail(user.Email, resetToken)
}

func (s *AuthService) generateToken(user *models.User) (string, error) {
	// Token süresi: 7 gün
	expirationTime := time.Now().Add(7 * 24 * time.Hour)

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     expirationTime.Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(s.jwtSecret)
}

// Reset token ile şifre değiştirme
func (s *AuthService) ResetPassword(token string, newPassword string) error {
	// Token'ı doğrula
	claims := &jwt.RegisteredClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil || !parsedToken.Valid {
		return errors.New("invalid or expired token")
	}

	// Token'dan email'i al
	email := claims.Subject
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return errors.New("user not found")
	}

	// Yeni şifreyi hashle
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Şifreyi güncelle
	return s.userRepo.UpdatePassword(user.ID, string(hashedPassword))
}
