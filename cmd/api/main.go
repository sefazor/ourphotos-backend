package main

import (
	"fmt"
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
	// .env dosyasını yüklemeye çalış ama bulunamazsa devam et
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
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

	// Rate Limiting Middleware'leri

	// 1. Çok hassas işlemler için sıkı limit
	authLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Hour,
		KeyGenerator: func(c *fiber.Ctx) string {
			// IP bazlı sınırlama
			return c.IP() + "_auth"
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(models.ErrorResponse(
				"Too many authentication attempts. Please try again later.",
			))
		},
	})

	// 2. Ödeme işlemleri için özel limit
	paymentLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Hour,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Kullanıcı ID bazlı sınırlama (oturum açmış kullanıcılar için)
			if userID := c.Locals("userID"); userID != nil {
				return fmt.Sprintf("payment_user_%v", userID)
			}
			return c.IP() + "_payment"
		},
	})

	// 3. Veri değiştirme işlemleri için limit
	writeLimiter := limiter.New(limiter.Config{
		Max:        20,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			if userID := c.Locals("userID"); userID != nil {
				return fmt.Sprintf("write_user_%v", userID)
			}
			return c.IP() + "_write"
		},
	})

	// 4. Okuma işlemleri için esnek limit
	readLimiter := limiter.New(limiter.Config{
		Max:        60,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			if userID := c.Locals("userID"); userID != nil {
				return fmt.Sprintf("read_user_%v", userID)
			}
			return c.IP() + "_read"
		},
	})

	// 5. Fotoğraf yükleme için özel yüksek limit
	uploadLimiter := limiter.New(limiter.Config{
		Max:        50,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			if userID := c.Locals("userID"); userID != nil {
				return fmt.Sprintf("upload_user_%v", userID)
			}
			return c.IP() + "_upload"
		},
	})

	// 6. Public erişim için çok yüksek limit
	publicLimiter := limiter.New(limiter.Config{
		Max:        200,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() + "_public"
		},
	})

	// Global rate limiter (varsayılan olarak tüm endpoint'ler için)
	globalLimiter := limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	})

	// Global limiter'ı uygula
	app.Use(globalLimiter)

	// Health check endpoint (API grubunun dışında, ana seviyede)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"message":   "Service is running",
			"version":   "1.0.0",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	api := app.Group("/api")

	// Public routes
	auth := api.Group("/auth")
	auth.Post("/register", authLimiter, authHandler.Register)
	auth.Post("/login", authLimiter, authHandler.Login)
	auth.Post("/forgot-password", authLimiter, authHandler.ForgotPassword)
	auth.Post("/reset-password", authLimiter, authHandler.ResetPassword)
	auth.Post("/verify-email", authLimiter, authHandler.VerifyEmail)
	auth.Post("/resend-verification", authLimiter, authHandler.ResendVerificationEmail)
	auth.Post("/complete-email-change", authLimiter, userHandler.CompleteEmailChange)

	// Public event routes
	api.Get("/events/:url", publicLimiter, eventHandler.GetEventByURL)
	api.Post("/events/url/:url/check-password", authLimiter, eventHandler.CheckEventPassword)
	api.Get("/gallery/:url", publicLimiter, photoHandler.GetPublicEventPhotos)

	// Public photo routes (authentication middleware'den ÖNCE olmalı)
	api.Post("/events/guest-upload/:url", uploadLimiter, photoHandler.UploadPhoto)

	// Stripe webhook (public)
	api.Post("/payments/webhook", paymentHandler.HandleStripeWebhook)

	// Public routes (auth middleware'den ÖNCE olmalı)
	api.Get("/payments/packages", publicLimiter, paymentHandler.GetCreditPackages)

	// Protected routes
	api.Use(middleware.AuthMiddleware())
	{
		user := api.Group("/user")
		user.Get("/profile", readLimiter, userHandler.GetMyProfile)
		user.Put("/profile", writeLimiter, userHandler.UpdateProfile)
		user.Post("/change-password", authLimiter, userHandler.ChangePassword)
		user.Post("/change-email", authLimiter, userHandler.InitiateEmailChange)

		events := api.Group("/events")
		events.Post("/", writeLimiter, eventHandler.CreateEvent)
		events.Get("/", readLimiter, eventHandler.GetUserEvents)
		events.Get("/detail/:url", readLimiter, eventHandler.GetEvent)
		events.Put("/:url", writeLimiter, eventHandler.UpdateEvent)
		events.Delete("/:url", writeLimiter, eventHandler.DeleteEvent)
		events.Post("/:url/photos", uploadLimiter, eventHandler.UploadEventPhotos)
		events.Get("/:url/qrcode", readLimiter, eventHandler.GetEventQRCode)

		// Photo routes
		photos := api.Group("/photos")
		photos.Get("/event/:url", readLimiter, photoHandler.GetEventPhotos)
		photos.Delete("/:id", writeLimiter, photoHandler.DeletePhoto)

		// Payment routes (protected)
		payments := api.Group("/payments")
		payments.Get("/history", readLimiter, paymentHandler.GetPurchaseHistory)
		payments.Post("/checkout/:packageId", paymentLimiter, paymentHandler.CreateCheckoutSession)

		// Credit package routes
		packages := api.Group("/packages")
		packages.Get("/", readLimiter, creditPackageHandler.GetAllPackages)
		packages.Get("/:id", readLimiter, creditPackageHandler.GetPackageByID)
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
