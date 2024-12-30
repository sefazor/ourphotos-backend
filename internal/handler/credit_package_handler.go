package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/service"
)

type CreditPackageHandler struct {
	packageService *service.PackageService
}

func NewCreditPackageHandler(packageService *service.PackageService) *CreditPackageHandler {
	return &CreditPackageHandler{
		packageService: packageService,
	}
}

func (h *CreditPackageHandler) GetAllPackages(c *fiber.Ctx) error {
	packages, err := h.packageService.GetAllPackages()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse(err.Error()))
	}

	return c.JSON(models.SuccessResponse(packages, "Packages retrieved successfully"))
}

func (h *CreditPackageHandler) GetPackageByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse("Invalid package ID"))
	}

	pkg, err := h.packageService.GetPackageByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse("Package not found"))
	}

	return c.JSON(models.SuccessResponse(pkg, "Package retrieved successfully"))
}
