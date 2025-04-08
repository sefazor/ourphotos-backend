package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
	"github.com/sefazor/ourphotos-backend/internal/config"
	"github.com/sefazor/ourphotos-backend/internal/handler"
	"github.com/sefazor/ourphotos-backend/internal/middleware"
	"github.com/sefazor/ourphotos-backend/internal/repository"
	"github.com/sefazor/ourphotos-backend/internal/service"
	"github.com/sefazor/ourphotos-backend/pkg/email"
	"github.com/sefazor/ourphotos-backend/pkg/payment"
	"github.com/sefazor/ourphotos-backend/pkg/qrcode"
	"github.com/sefazor/ourphotos-backend/pkg/storage"
	"github.com/sefazor/ourphotos-backend/pkg/utils"
)

// InitializeAPI wire yapısını tanımlar, uygulamanın bağımlılık enjeksiyonu için kullanılır
func InitializeAPI(cfg *config.Config) (*fiber.App, error) {
	wire.Build(
		// Repositories
		repository.NewEventRepository,
		repository.NewUserRepository,
		repository.NewPhotoRepository,
		repository.NewCreditPackageRepository,
		repository.NewUserCreditPurchaseRepository,

		// Storage & External Services
		storage.NewCloudflareStorage,
		storage.NewCloudflareImages,
		email.NewEmailService,
		payment.NewStripeService,
		qrcode.NewQRService,

		// Services
		service.NewAuthService,
		service.NewUserService,
		service.NewEventService,
		service.NewPhotoService,
		service.NewPackageService,
		service.NewPaymentService,

		// Validator
		utils.NewValidator,

		// Handlers
		handler.NewAuthHandler,
		handler.NewEventHandler,
		handler.NewUserHandler,
		handler.NewPhotoHandler,
		handler.NewPaymentHandler,
		handler.NewCreditPackageHandler,

		// Middleware
		middleware.AuthMiddleware,

		// App initialization
		initializeFiberApp,
	)
	return nil, nil
}

// initializeFiberApp fiber uygulama örneğini oluşturur ve döndürür
func initializeFiberApp(
	authHandler *handler.AuthHandler,
	eventHandler *handler.EventHandler,
	userHandler *handler.UserHandler,
	photoHandler *handler.PhotoHandler,
	paymentHandler *handler.PaymentHandler,
	packageHandler *handler.CreditPackageHandler,
	authMiddleware func() fiber.Handler,
) *fiber.App {
	app := fiber.New()
	// Routing ve middleware yapılandırması burada yapılabilir
	return app
}
