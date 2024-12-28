package controller

import (
	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/service"
)

type AuthController struct {
	authService *service.AuthService
}

func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

func (c *AuthController) Register(req models.RegisterRequest) (*models.AuthResponse, error) {
	return c.authService.Register(req)
}

func (c *AuthController) Login(req models.LoginRequest) (*models.AuthResponse, error) {
	return c.authService.Login(req)
}

func (c *AuthController) ForgotPassword(email string) error {
	return c.authService.ForgotPassword(email)
}

func (c *AuthController) ResetPassword(token string, newPassword string) error {
	return c.authService.ResetPassword(token, newPassword)
}
