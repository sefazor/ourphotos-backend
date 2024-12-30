package utils

import (
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func NewValidator() *Validator {
	v := validator.New()

	// Custom validations
	v.RegisterValidation("supported_image", validateImageType)

	return &Validator{
		validate: v,
	}
}

func (v *Validator) Struct(s interface{}) error {
	return v.validate.Struct(s)
}

// Desteklenen resim formatlarını kontrol et
func validateImageType(fl validator.FieldLevel) bool {
	mimeType := fl.Field().String()
	supportedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	return supportedTypes[mimeType]
}
