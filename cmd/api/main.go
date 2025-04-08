package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"

	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/sefazor/ourphotos-backend/internal/config"
	"github.com/sefazor/ourphotos-backend/internal/handler"
	"github.com/sefazor/ourphotos-backend/internal/middleware"
	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/repository"
	"github.com/sefazor/ourphotos-backend/internal/service"
	"github.com/sefazor/ourphotos-backend/pkg/database"
	"github.com/sefazor/ourphotos-backend/pkg/email"
	"github.com/sefazor/ourphotos-backend/pkg/payment"
	"github.com/sefazor/ourphotos-backend/pkg/qrcode"
	"github.com/sefazor/ourphotos-backend/pkg/storage"
	"github.com/sefazor/ourphotos-backend/pkg/utils"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Config'i yükle
	cfg := config.LoadConfig()

	// Initialize database
	db := database.NewDatabase()

	// Run migrations
	if err := db.AutoMigrate(
		&models.User{},
		&models.Event{},
		&models.Photos{},
		&models.CreditPackage{},
		&models.UserCreditPurchase{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Repositories
	userRepo := repository.NewUserRepository(db)
	eventRepo := repository.NewEventRepository(db)
	photoRepo := repository.NewPhotoRepository(db)
	packageRepo := repository.NewCreditPackageRepository(db)
	purchaseRepo := repository.NewUserCreditPurchaseRepository(db)

	// Storage services
	imgStorage := storage.NewCloudflareImages(
		cfg.CloudflareImages.AccountID,
		cfg.CloudflareImages.Token,
		cfg.CloudflareImages.Hash,
	)

	// Email service
	emailService := email.NewEmailService()

	// Services
	authService := service.NewAuthService(userRepo, emailService)
	userService := service.NewUserService(userRepo, emailService)
	photoService := service.NewPhotoService(
		photoRepo,
		eventRepo,
		imgStorage,
		userRepo,
	)

	// QR Code Service
	qrService := qrcode.NewQRService("https://ourphotos.co/e/")

	eventService := service.NewEventService(eventRepo, userRepo, photoService, qrService)

	// Stripe service
	stripeService := payment.NewStripeService(os.Getenv("STRIPE_SECRET_KEY"))

	// Payment service
	paymentService := service.NewPaymentService(
		stripeService,
		userRepo,
		packageRepo,
		purchaseRepo,
	)

	// Validator'ı önce tanımla
	validator := utils.NewValidator()

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	eventHandler := handler.NewEventHandler(eventService, userService, validator)
	userHandler := handler.NewUserHandler(userService)
	photoHandler := handler.NewPhotoHandler(photoService, eventService)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	packageService := service.NewPackageService(packageRepo)
	creditPackageHandler := handler.NewCreditPackageHandler(packageService)

	// Router
	app := fiber.New(fiber.Config{
		BodyLimit:    300 * 1024 * 1024, // 300MB limit
		ReadTimeout:  5 * time.Minute,   // 5 dakika timeout
		WriteTimeout: 5 * time.Minute,   // 5 dakika timeout
	})

	// Global Middleware'ler önce tanımlanmalı
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "https://ourphotos.co, https://www.ourphotos.co, http://localhost:5173",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE",
		AllowCredentials: true,
	}))
	app.Use(logger.New())
	app.Use(limiter.New(limiter.Config{
		Max:        20,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}))

	api := app.Group("/api")

	// Public routes
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/forgot-password", authHandler.ForgotPassword)
	auth.Post("/reset-password", authHandler.ResetPassword)
	auth.Post("/verify-email", authHandler.VerifyEmail)
	auth.Post("/resend-verification", authHandler.ResendVerificationEmail)
	auth.Post("/complete-email-change", userHandler.CompleteEmailChange)

	// Public event routes
	api.Get("/events/:url", eventHandler.GetEventByURL)
	api.Post("/events/url/:url/check-password", eventHandler.CheckEventPassword)
	api.Get("/gallery/:url", photoHandler.GetPublicEventPhotos)

	// Public photo routes (authentication middleware'den ÖNCE olmalı)
	api.Post("/events/guest-upload/:url", photoHandler.UploadPhoto)

	// Stripe webhook (public)
	api.Post("/payments/webhook", paymentHandler.HandleStripeWebhook)

	// Public routes (auth middleware'den ÖNCE olmalı)
	api.Get("/payments/packages", paymentHandler.GetCreditPackages)

	// Protected routes
	api.Use(middleware.AuthMiddleware())
	{
		user := api.Group("/user")
		user.Get("/profile", userHandler.GetMyProfile)
		user.Put("/profile", userHandler.UpdateProfile)
		user.Post("/change-password", userHandler.ChangePassword)
		user.Post("/change-email", userHandler.InitiateEmailChange)

		events := api.Group("/events")
		events.Post("/", eventHandler.CreateEvent)
		events.Get("/", eventHandler.GetUserEvents)
		events.Get("/detail/:url", eventHandler.GetEvent)
		events.Put("/:url", eventHandler.UpdateEvent)
		events.Delete("/:url", eventHandler.DeleteEvent)
		events.Post("/:url/photos", eventHandler.UploadEventPhotos)
		events.Get("/:url/qrcode", eventHandler.GetEventQRCode)

		// Photo routes
		photos := api.Group("/photos")
		photos.Get("/event/:url", photoHandler.GetEventPhotos)
		photos.Delete("/:id", photoHandler.DeletePhoto)

		// Payment routes (protected)
		payments := api.Group("/payments")
		payments.Get("/history", paymentHandler.GetPurchaseHistory)
		payments.Post("/checkout/:packageId", paymentHandler.CreateCheckoutSession)

		// Credit package routes
		packages := api.Group("/packages")
		packages.Get("/", creditPackageHandler.GetAllPackages)
		packages.Get("/:id", creditPackageHandler.GetPackageByID)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Süresi dolmuş etkinlikleri temizleme zamanlayıcısı
	go func() {
		// İlk temizleme işlemi
		if err := eventService.CleanupExpiredEvents(); err != nil {
			log.Printf("Error cleaning up expired events: %v\n", err)
		}

		// Her gün aynı saatte çalışacak zamanlayıcı
		ticker := time.NewTicker(24 * time.Hour)
		for range ticker.C {
			if err := eventService.CleanupExpiredEvents(); err != nil {
				log.Printf("Error cleaning up expired events: %v\n", err)
			}
		}
	}()

	log.Fatal(app.Listen(":" + port))
}
