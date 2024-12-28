package controller

import (
	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/service"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController(userService *service.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

func (c *UserController) GetUserByEmail(email string) (*models.User, error) {
	return c.userService.GetUserByEmail(email)
}

func (c *UserController) GetUserByID(id uint) (*models.User, error) {
	return c.userService.GetUserByID(id)
}

func (c *UserController) ChangePassword(userID uint, req models.ChangePasswordRequest) error {
	return c.userService.ChangePassword(userID, req)
}

func (c *UserController) InitiateEmailChange(userID uint, req models.ChangeEmailRequest) error {
	return c.userService.InitiateEmailChange(userID, req)
}

func (c *UserController) CompleteEmailChange(token string) error {
	return c.userService.CompleteEmailChange(token)
}

func (c *UserController) UpdateProfile(userID uint, req models.UpdateProfileRequest) (*models.User, error) {
	return c.userService.UpdateProfile(userID, req)
}
