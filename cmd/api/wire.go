package wire

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
	"github.com/sefazor/ourphotos-backend/internal/controller"
	"github.com/sefazor/ourphotos-backend/internal/handler"
	"github.com/sefazor/ourphotos-backend/internal/repository"
	"github.com/sefazor/ourphotos-backend/internal/service"
	"github.com/sefazor/ourphotos-backend/pkg/config"
	"github.com/sefazor/ourphotos-backend/pkg/storage"
	"github.com/sefazor/ourphotos-backend/pkg/utils"
)

func InitializeAPI(config *config.Config) (*fiber.App, error) {
	wire.Build(
		// Repositories
		repository.NewEventRepository,
		repository.NewUserRepository,
		repository.NewPhotoRepository,

		// Storage
		storage.NewCloudflareStorage,
		storage.NewCloudflareImages,

		// Services
		wire.Struct(new(service.EventService), "*"),
		wire.Struct(new(service.UserService), "*"),
		wire.Struct(new(service.PhotoService), "*"),

		// Controllers
		wire.Struct(new(controller.EventController), "*"),
		wire.Struct(new(controller.UserController), "*"),
		wire.Struct(new(controller.PhotoController), "*"),

		// Validator
		utils.NewValidator,

		// Handlers
		wire.Struct(new(handler.EventHandler), "*"),
		wire.Struct(new(handler.UserHandler), "*"),
		wire.Struct(new(handler.PhotoHandler), "*"),

		// App
		NewFiberApp,
	)
	return nil, nil
}

func NewFiberApp(
	eventHandler *handler.EventHandler,
	userHandler *handler.UserHandler,
	photoHandler *handler.PhotoHandler,
	config *config.Config,
) *fiber.App {
	// ... fiber app setup ...
	return app
}
