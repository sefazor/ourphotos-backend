package service

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sefazor/ourphotos-backend/internal/models"
	"github.com/sefazor/ourphotos-backend/internal/repository"
	"github.com/sefazor/ourphotos-backend/pkg/payment"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/price"
	"github.com/stripe/stripe-go/v74/product"
)

type PaymentService struct {
	stripeService *payment.StripeService
	userRepo      *repository.UserRepository
	packageRepo   *repository.CreditPackageRepository
	purchaseRepo  *repository.UserCreditPurchaseRepository
}

func NewPaymentService(stripeService *payment.StripeService, userRepo *repository.UserRepository, packageRepo *repository.CreditPackageRepository, purchaseRepo *repository.UserCreditPurchaseRepository) *PaymentService {
	return &PaymentService{
		stripeService: stripeService,
		userRepo:      userRepo,
		packageRepo:   packageRepo,
		purchaseRepo:  purchaseRepo,
	}
}

func (s *PaymentService) CreateCheckoutSession(userID uint, packageID uint) (*models.CheckoutSession, error) {
	// Paketi bul
	creditPackage, err := s.packageRepo.GetByID(packageID)
	if err != nil {
		return nil, err
	}

	// Kullanıcıyı bul
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Stripe'da geçici product oluştur
	productParams := &stripe.ProductParams{
		Name: stripe.String(creditPackage.Name),
		Description: stripe.String(fmt.Sprintf("%d events, %d photos",
			creditPackage.EventLimit,
			creditPackage.PhotoLimit)),
	}
	prod, err := product.New(productParams)
	if err != nil {
		return nil, err
	}

	// Product için price oluştur
	priceParams := &stripe.PriceParams{
		Product:    stripe.String(prod.ID),
		UnitAmount: stripe.Int64(int64(creditPackage.Price * 100)), // USD to cents
		Currency:   stripe.String(string(stripe.CurrencyUSD)),
	}
	p, err := price.New(priceParams)
	if err != nil {
		return nil, err
	}

	// Checkout session oluştur
	session, err := s.stripeService.CreateCheckoutSession(
		user.Email,
		p.ID,
		map[string]string{
			"user_id":    fmt.Sprintf("%d", userID),
			"package_id": fmt.Sprintf("%d", packageID),
		},
	)
	if err != nil {
		return nil, err
	}

	// Purchase kaydı oluştur
	purchase := &models.UserCreditPurchase{
		UserID:          userID,
		PackageID:       packageID,
		EventLimit:      creditPackage.EventLimit,
		PhotoLimit:      creditPackage.PhotoLimit,
		Price:           creditPackage.Price,
		StripeSessionID: session.ID,
		Status:          "pending",
	}

	if err := s.purchaseRepo.Create(purchase); err != nil {
		return nil, err
	}

	return &models.CheckoutSession{
		ID:  session.ID,
		URL: session.URL,
	}, nil
}

// Webhook handler for Stripe events
func (s *PaymentService) HandleStripeWebhook(event *stripe.Event) error {
	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			return err
		}

		// Metadata'dan user_id ve package_id'yi al
		userID, err := strconv.ParseUint(session.Metadata["user_id"], 10, 32)
		if err != nil {
			return err
		}

		// Purchase'ı bul ve güncelle
		purchase, err := s.purchaseRepo.GetBySessionID(session.ID)
		if err != nil {
			return err
		}

		purchase.Status = "completed"
		if err := s.purchaseRepo.Update(purchase); err != nil {
			return err
		}

		// Kullanıcıyı bul
		user, err := s.userRepo.GetByID(uint(userID))
		if err != nil {
			return err
		}

		// Kullanıcının limitlerini güncelle
		user.EventLimit += purchase.EventLimit
		user.PhotoLimit += purchase.PhotoLimit

		return s.userRepo.Update(user)
	}

	return nil
}

func (s *PaymentService) GetCreditPackages() ([]models.CreditPackage, error) {
	// Aktif paketleri getir
	packages, err := s.packageRepo.GetAll()
	if err != nil {
		return nil, err
	}
	return packages, nil
}

func (s *PaymentService) GetUserPurchaseHistory(userID uint) ([]models.UserCreditPurchase, error) {
	return s.purchaseRepo.GetUserPurchaseHistory(userID)
}
