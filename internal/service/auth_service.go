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

const (
	// Token süreleri
	TokenExpiryLogin       = 7 * 24 * time.Hour // 7 gün
	TokenExpiryReset       = 15 * time.Minute   // 15 dakika
	TokenExpiryEmailChange = 15 * time.Minute   // 15 dakika
	TokenExpiryEmailVerify = 24 * time.Hour     // 24 saat
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
		FullName:   req.FullName,
		Email:      req.Email,
		Password:   string(hashedPassword),
		EventLimit: 1,     // Default 1 event
		PhotoLimit: 20,    // Default 20 photos
		IsVerified: false, // Doğrulama gerekli
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Email doğrulama tokeni oluştur
	verificationToken, err := s.generateVerificationToken(user.Email)
	if err != nil {
		return nil, err
	}

	// Doğrulama emaili gönder
	go s.emailService.SendVerificationEmail(user.Email, user.FullName, verificationToken)

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

func (s *AuthService) VerifyEmail(token string) error {
	// Token doğrula
	claims, err := jwtPkg.ValidateToken(token)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return errors.New("invalid token claims")
	}

	// Kullanıcıyı bul
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return errors.New("user not found")
	}

	// Kullanıcı zaten doğrulanmış mı?
	if user.IsVerified {
		return errors.New("email already verified")
	}

	// Kullanıcıyı doğrulanmış olarak işaretle
	user.IsVerified = true
	if err := s.userRepo.Update(user); err != nil {
		return errors.New("failed to verify email")
	}

	return nil
}

func (s *AuthService) ResendVerificationEmail(email string) error {
	// Kullanıcıyı bul
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return errors.New("user not found")
	}

	// Kullanıcı zaten doğrulanmış mı?
	if user.IsVerified {
		return errors.New("email already verified")
	}

	// Yeni doğrulama tokeni oluştur
	verificationToken, err := s.generateVerificationToken(email)
	if err != nil {
		return err
	}

	// Doğrulama emaili gönder
	return s.emailService.SendVerificationEmail(user.Email, user.FullName, verificationToken)
}

func (s *AuthService) generateVerificationToken(email string) (string, error) {
	// Email doğrulama tokeni oluştur
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(TokenExpiryEmailVerify).Unix(),
		"iat":   time.Now().Unix(),
		"iss":   s.jwtIssuer,
		"type":  "email_verification",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) ForgotPassword(email string) error {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil // Güvenlik için hata dönme
	}

	// Reset token oluştur
	claims := jwt.MapClaims{
		"sub": user.Email,
		"exp": time.Now().Add(TokenExpiryReset).Unix(),
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

// Reset token ile şifre değiştirme
func (s *AuthService) ResetPassword(token string, newPassword string) error {
	claims, err := jwtPkg.ValidateToken(token)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	email, ok := claims["sub"].(string)
	if !ok {
		return errors.New("invalid token claims")
	}

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return err
	}

	// Yeni şifreyi hashle
	hashedPassword, err := bcrypt.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Şifreyi güncelle
	if err := s.userRepo.UpdatePassword(user.ID, hashedPassword); err != nil {
		return err
	}

	return nil
}
