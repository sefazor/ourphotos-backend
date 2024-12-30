package service

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/repository"
	"github.com/sefazor/ourphotos-backend/pkg/bcrypt"
	"github.com/sefazor/ourphotos-backend/pkg/email"
	jwtPkg "github.com/sefazor/ourphotos-backend/pkg/jwt"
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
	hashedPassword, err := bcrypt.HashPassword(req.Password)
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
	token, err := jwtPkg.GenerateToken(user.Email, user.ID)
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
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	fmt.Printf("\nLogin Debug:\n")
	fmt.Printf("User Email: %s\n", req.Email)
	fmt.Printf("Stored Hash: %s\n", user.Password)

	// Şifre karşılaştırma - ComparePassword kullanıyoruz
	if err := bcrypt.ComparePassword(user.Password, req.Password); err != nil {
		fmt.Printf("Login failed: %v\n", err)
		return nil, errors.New("invalid email or password")
	}

	// JWT token oluştur
	token, err := jwtPkg.GenerateToken(user.Email, user.ID)
	if err != nil {
		return nil, fmt.Errorf("token generation failed: %v", err)
	}

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
	claims := jwt.MapClaims{
		"sub": user.Email,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
		"iss": s.jwtIssuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

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
	claims, err := jwtPkg.ValidateToken(token)
	if err != nil {
		return fmt.Errorf("token validation failed: %v", err)
	}

	email, ok := claims["sub"].(string)
	if !ok {
		return errors.New("invalid token claims")
	}

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return err
	}

	fmt.Printf("\nReset Password Debug:\n")
	fmt.Printf("User Email: %s\n", email)
	fmt.Printf("New Password: %s\n", newPassword)

	// Yeni şifreyi hashle
	hashedPassword, err := bcrypt.HashPassword(newPassword)
	if err != nil {
		return err
	}
	fmt.Printf("Generated Hash: %s\n", hashedPassword)

	// Debug için
	bcrypt.DebugHashAndCompare(newPassword)

	// Şifreyi güncelle
	if err := s.userRepo.UpdatePassword(user.ID, hashedPassword); err != nil {
		return err
	}

	// Doğrulama
	updatedUser, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return err
	}
	fmt.Printf("Stored Hash: %s\n", updatedUser.Password)

	// Test et - ComparePassword kullanıyoruz
	if err := bcrypt.ComparePassword(updatedUser.Password, newPassword); err != nil {
		fmt.Printf("Verification Error: %v\n", err)
		return fmt.Errorf("password verification failed: %v", err)
	}

	fmt.Printf("Password reset successful!\n")
	return nil
}
