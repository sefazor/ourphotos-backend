package email

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/resendlabs/resend-go"
)

type EmailService struct {
	client       *resend.Client
	from         string
	fromName     string
	templatesDir string
	logger       *log.Logger
}

func NewEmailService() *EmailService {
	// Log dosyası oluştur
	logFile, err := os.OpenFile("logs/email.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Printf("Error opening log file: %v", err)
		// Hata durumunda stdout'a log al
		return &EmailService{
			client:       resend.NewClient(os.Getenv("RESEND_API_KEY")),
			from:         os.Getenv("EMAIL_FROM_ADDRESS"),
			fromName:     os.Getenv("EMAIL_FROM_NAME"),
			templatesDir: "pkg/email/templates",
			logger:       log.New(os.Stdout, "EMAIL: ", log.LstdFlags),
		}
	}

	// Multi writer ile hem dosyaya hem stdout'a yaz
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	return &EmailService{
		client:       resend.NewClient(os.Getenv("RESEND_API_KEY")),
		from:         os.Getenv("EMAIL_FROM_ADDRESS"),
		fromName:     os.Getenv("EMAIL_FROM_NAME"),
		templatesDir: "pkg/email/templates",
		logger:       log.New(multiWriter, "EMAIL: ", log.LstdFlags),
	}
}

func (s *EmailService) SendWelcomeEmail(email, fullName string) error {
	s.logger.Printf("Sending welcome email to: %s (%s)", email, fullName)

	templateData := map[string]interface{}{
		"FullName": fullName,
		"Email":    email,
		"Year":     time.Now().Year(),
	}

	html, err := s.parseTemplate("welcome.html", templateData)
	if err != nil {
		s.logger.Printf("Error parsing welcome template for %s: %v", email, err)
		return err
	}

	params := &resend.SendEmailRequest{
		From:    s.fromName + " <" + s.from + ">",
		To:      []string{email},
		Subject: "Welcome to OurPhotos!",
		Html:    html,
	}

	resp, err := s.client.Emails.Send(params)
	if err != nil {
		s.logger.Printf("Failed to send welcome email to %s: %v", email, err)
		return err
	}

	s.logger.Printf("Successfully sent welcome email to %s (ID: %s)", email, resp.Id)
	return nil
}

func (s *EmailService) SendPasswordResetEmail(email string, resetToken string) error {
	s.logger.Printf("Sending password reset email to: %s", email)

	resetLink := os.Getenv("FRONTEND_URL") + "/reset-password?token=" + resetToken

	templateData := map[string]interface{}{
		"ResetLink": resetLink,
		"Email":     email,
		"Year":      time.Now().Year(),
	}

	html, err := s.parseTemplate("reset-password.html", templateData)
	if err != nil {
		s.logger.Printf("Error parsing reset password template for %s: %v", email, err)
		return err
	}

	params := &resend.SendEmailRequest{
		From:    s.fromName + " <" + s.from + ">",
		To:      []string{email},
		Subject: "Reset Your Password - OurPhotos",
		Html:    html,
	}

	resp, err := s.client.Emails.Send(params)
	if err != nil {
		s.logger.Printf("Failed to send reset password email to %s: %v", email, err)
		return err
	}

	s.logger.Printf("Successfully sent reset password email to %s (ID: %s)", email, resp.Id)
	return nil
}

func (s *EmailService) SendEmailChangeVerification(email, token string) error {
	templateData := map[string]interface{}{
		"VerificationLink": os.Getenv("FRONTEND_URL") + "/verify-email?token=" + token,
		"Email":            email,
		"Year":             time.Now().Year(),
	}

	html, err := s.parseTemplate("verify-email.html", templateData)
	if err != nil {
		return err
	}

	params := &resend.SendEmailRequest{
		From:    s.fromName + " <" + s.from + ">",
		To:      []string{email},
		Subject: "Verify Your New Email - OurPhotos",
		Html:    html,
	}

	_, err = s.client.Emails.Send(params)
	return err
}

func (s *EmailService) SendVerificationEmail(email, fullName, token string) error {
	s.logger.Printf("Sending verification email to: %s", email)

	verificationLink := os.Getenv("FRONTEND_URL") + "/verify-email?token=" + token

	templateData := map[string]interface{}{
		"FullName":         fullName,
		"VerificationLink": verificationLink,
		"Email":            email,
		"Year":             time.Now().Year(),
	}

	html, err := s.parseTemplate("verify-email.html", templateData)
	if err != nil {
		s.logger.Printf("Error parsing verification template for %s: %v", email, err)
		return err
	}

	params := &resend.SendEmailRequest{
		From:    s.fromName + " <" + s.from + ">",
		To:      []string{email},
		Subject: "Verify Your Email - OurPhotos",
		Html:    html,
	}

	resp, err := s.client.Emails.Send(params)
	if err != nil {
		s.logger.Printf("Failed to send verification email to %s: %v", email, err)
		return err
	}

	s.logger.Printf("Successfully sent verification email to %s (ID: %s)", email, resp.Id)
	return nil
}

func (s *EmailService) parseTemplate(templateName string, data interface{}) (string, error) {
	templatePath := filepath.Join(s.templatesDir, templateName)

	s.logger.Printf("Parsing template: %s", templatePath)

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		s.logger.Printf("Error parsing template %s: %v", templateName, err)
		return "", err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		s.logger.Printf("Error executing template %s: %v", templateName, err)
		return "", err
	}

	return body.String(), nil
}
