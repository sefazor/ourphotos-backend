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
	"github.com/sefazor/ourphotos-backend/internal/controller"
	"github.com/sefazor/ourphotos-backend/internal/handler"
	"github.com/sefazor/ourphotos-backend/internal/middleware"
	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/repository"
	"github.com/sefazor/ourphotos-backend/internal/service"
	"github.com/sefazor/ourphotos-backend/pkg/database"
	"github.com/sefazor/ourphotos-backend/pkg/email"
	"github.com/sefazor/ourphotos-backend/pkg/payment"
	"github.com/sefazor/ourphotos-backend/pkg/storage"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize database
	db := database.NewDatabase()

	// Run migrations
	if err := db.AutoMigrate(
		&models.User{},
		&models.Event{},
		&models.CreditPackage{},      // Yeni
		&models.UserCreditPurchase{}, // Yeni
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
	cloudflareR2, err := storage.NewCloudflareStorage()
	if err != nil {
		log.Fatal("Failed to initialize Cloudflare R2:", err)
	}
	cloudflareImages := storage.NewCloudflareImages()

	// Services
	emailService := email.NewEmailService()
	authService := service.NewAuthService(userRepo, emailService)
	userService := service.NewUserService(userRepo, emailService)
	eventService := service.NewEventService(eventRepo, userRepo)
	photoService := service.NewPhotoService(photoRepo, eventRepo, cloudflareR2, cloudflareImages)

	// Stripe service
	stripeService := payment.NewStripeService(os.Getenv("STRIPE_SECRET_KEY"))

	// Payment service
	paymentService := service.NewPaymentService(stripeService, userRepo, packageRepo, purchaseRepo)

	// Payment controller
	paymentController := controller.NewPaymentController(paymentService)

	// Payment handler
	paymentHandler := handler.NewPaymentHandler(paymentController)

	// Controllers
	authController := controller.NewAuthController(authService)
	userController := controller.NewUserController(userService)
	eventController := controller.NewEventController(eventService)
	photoController := controller.NewPhotoController(photoService)

	// Handlers
	authHandler := handler.NewAuthHandler(authController)
	userHandler := handler.NewUserHandler(userController)
	eventHandler := handler.NewEventHandler(eventController, userController)
	photoHandler := handler.NewPhotoHandler(
		photoController,
		userController,
		eventController,
	)

	// Router
	app := fiber.New()

	// Global Middleware'ler önce tanımlanmalı
	app.Use(cors.New())
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
	auth.Post("/verify-email", userHandler.CompleteEmailChange)

	// Public event routes
	api.Get("/events/:url", eventHandler.GetEventByURL)
	api.Post("/events/:id/check-password", eventHandler.CheckEventPassword)
	api.Get("/gallery/:url", photoHandler.GetPublicEventPhotos)
	api.Post("/events/photos", photoHandler.UploadPhoto)

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
		events.Get("/:id", eventHandler.GetEvent)
		events.Put("/:id", eventHandler.UpdateEvent)
		events.Delete("/:id", eventHandler.DeleteEvent)

		// Photo routes
		photos := api.Group("/photos")
		photos.Get("/event/:eventId", photoHandler.GetEventPhotos)
		photos.Delete("/:id", photoHandler.DeletePhoto)

		// Payment routes (protected)
		payments := api.Group("/payments")
		payments.Get("/history", paymentHandler.GetPurchaseHistory)
		payments.Post("/checkout/:packageId", paymentHandler.CreateCheckoutSession)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(app.Listen(":" + port))
}
