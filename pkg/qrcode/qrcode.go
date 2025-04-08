package qrcode

import (
	"fmt"

	"github.com/skip2/go-qrcode"
)

// QRService, QR kod oluşturma ve yönetme işlemlerini sağlayan servis
type QRService struct {
	baseURL string // Temel URL (örn: "https://ourphotos.co/e/")
}

// NewQRService, yeni bir QRService oluşturur
func NewQRService(baseURL string) *QRService {
	return &QRService{
		baseURL: baseURL,
	}
}

// GenerateQRCode, verilen etkinlik URL kodu için PNG formatında QR kod bayt dizisi oluşturur
func (s *QRService) GenerateQRCode(eventURLCode string, size int) ([]byte, error) {
	// Tam URL'i oluştur
	fullURL := fmt.Sprintf("%s%s", s.baseURL, eventURLCode)

	// QR kod oluştur ve PNG olarak döndür
	png, err := qrcode.Encode(fullURL, qrcode.Medium, size)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code PNG: %w", err)
	}

	return png, nil
}
