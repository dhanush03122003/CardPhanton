package card

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/gin-gonic/gin"

	"webauthn-server/internal/api/middleware"
	"webauthn-server/internal/audit"
	"webauthn-server/internal/timeutil"
	"webauthn-server/internal/user"
)

var (
	ErrUnauthorized             = errors.New("unauthorized")
	ErrCardNotFound             = errors.New("card not found")
	ErrNotAuthorized            = errors.New("not authorized to modify this card")
	ErrInvalidRequest           = errors.New("invalid request")
	ErrFailedToFetch            = errors.New("failed to fetch card")
	ErrFailedToCreate           = errors.New("failed to create card")
	ErrFailedToUpdate           = errors.New("failed to update card")
	ErrFailedToDelete           = errors.New("failed to delete card")
	ErrInvalidPAN               = errors.New("invalid PAN: must be a valid Visa, Mastercard, Amex, or Discover card number")
	ErrDuplicatePAN             = errors.New("a card with this PAN already exists")
	ErrInvalidCVV               = errors.New("invalid CVV: must be exactly 3 digits")
	ErrInvalidExpiry            = errors.New("invalid expiry date: card has expired")
	ErrInvalidCardBrand         = errors.New("invalid card brand: must be Visa, Mastercard, American Express, Discover, RuPay, or Other")
	ErrInvalidPaymentMethod     = errors.New("invalid payment method type: must be Credit or Debit")
	ErrInvalidExpMonth          = errors.New("invalid expiry month: must be 1-12")
	ErrInvalidExpYear           = errors.New("invalid expiry year: must be 2024 or later")
	ErrInvalidLinkedPhoneNumber = errors.New("invalid linked phone number: must be exactly 10 digits and start with 6-9")
)

// ValidCardBrands contains valid card brand names
var ValidCardBrands = map[string]bool{
	"Visa":             true,
	"Mastercard":       true,
	"American Express": true,
	"Discover":         true,
	"RuPay":            true,
	"Other":            true,
}

// PAN regex: Valid Visa, Mastercard, Amex, or Discover card numbers
var panRegex = regexp.MustCompile(`^(?:4[0-9]{12}(?:[0-9]{3})?|(?:5[1-5][0-9]{2}|222[1-9]|22[3-9][0-9]|2[3-6][0-9]{2}|27[01][0-9]|2720)[0-9]{12}|6(?:011|5[0-9]{2})[0-9]{12}|(?:60|65|81|82|508)[0-9]{13})$`)

// CVV regex: exactly 3 digits
var cvvRegex = regexp.MustCompile(`^[0-9]{3}$`)

var phoneRegex = regexp.MustCompile(`^[6-9][0-9]{9}$`)

// CardService handles card business logic
type CardService struct {
	repo     Repository
	userRepo user.UserRepository
	audit    audit.Repository
}

// NewCardService creates a new card service
func NewCardService(repo Repository, userRepo user.UserRepository, auditRepo audit.Repository) *CardService {
	return &CardService{
		repo:     repo,
		userRepo: userRepo,
		audit:    auditRepo,
	}
}

// GetUserIDFromContext extracts user ID from JWT context
func (s *CardService) GetUserIDFromContext(c *gin.Context) (string, error) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return "", ErrUnauthorized
	}
	return userID, nil
}

// GetAllCards returns all cards for a user
func (s *CardService) GetAllCards(ctx context.Context, userID string) ([]*Card, error) {
	cards, err := s.repo.GetCardsLinkedToUserId(ctx, userID)
	if err != nil {
		return nil, ErrFailedToFetch
	}

	if cards == nil {
		cards = []*Card{}
	}

	return cards, nil
}

// GetGlobalCards returns every card for authenticated users.
func (s *CardService) GetGlobalCards(ctx context.Context) ([]*Card, error) {
	cards, err := s.repo.GetAllCards(ctx)
	if err != nil {
		return nil, ErrFailedToFetch
	}
	if cards == nil {
		cards = []*Card{}
	}
	return cards, nil
}

// CreateCardInput represents the input for creating a card
type CreateCardInput struct {
	PAN               string
	CardholderName    string
	BankName          string
	PaymentMethodType PaymentMethodType
	CardBrand         string
	ProductName       string
	LinkedPhoneNumber string
	ExpMonth          int
	ExpYear           int
	Cvv               int
}

// validateCardInput validates the card input fields
func (s *CardService) validateCardInput(input CreateCardInput) error {
	// Validate PAN: must be 15 or 16 digits
	if !panRegex.MatchString(input.PAN) {
		return ErrInvalidPAN
	}
	if !phoneRegex.MatchString(input.LinkedPhoneNumber) {
		return ErrInvalidLinkedPhoneNumber
	}

	// Validate CVV: must be exactly 3 digits
	if !cvvRegex.MatchString(fmt.Sprintf("%d", input.Cvv)) {
		return ErrInvalidCVV
	}

	// Validate Card Brand
	if !ValidCardBrands[input.CardBrand] {
		return ErrInvalidCardBrand
	}

	// Validate Payment Method Type
	if !input.PaymentMethodType.IsValid() {
		return ErrInvalidPaymentMethod
	}

	// Validate Expiry Month: must be 1-12
	if input.ExpMonth < 1 || input.ExpMonth > 12 {
		return ErrInvalidExpMonth
	}

	// Validate Expiry Year: must be 2024 or later
	now := timeutil.Now()
	currentYear := now.Year()
	if input.ExpYear < currentYear || input.ExpYear > currentYear+20 {
		return ErrInvalidExpYear
	}

	// Check if card is expired
	if input.ExpYear == currentYear && input.ExpMonth < int(now.Month()) {
		return ErrInvalidExpiry
	}

	return nil
}

// CardActionInput contains additional context for audit logging
type CardActionInput struct {
	IPAddress string
	UserAgent string
}

// CreateCard creates a new card for a user
func (s *CardService) CreateCard(ctx context.Context, userID string, input CreateCardInput, actionInput CardActionInput) (*Card, error) {
	// Validate input
	if err := s.validateCardInput(input); err != nil {
		return nil, err
	}

	// Check if PAN already exists (PAN must be unique)
	existingCard, err := s.repo.GetCardByPAN(ctx, input.PAN)
	if err != nil {
		return nil, ErrFailedToCreate
	}
	if existingCard != nil {
		return nil, ErrDuplicatePAN
	}

	card := &Card{
		UserID:            userID,
		PAN:               input.PAN,
		CardholderName:    input.CardholderName,
		BankName:          input.BankName,
		PaymentMethodType: input.PaymentMethodType,
		CardBrand:         input.CardBrand,
		ProductName:       input.ProductName,
		LinkedPhoneNumber: input.LinkedPhoneNumber,
		ExpMonth:          input.ExpMonth,
		ExpYear:           input.ExpYear,
		Cvv:               input.Cvv,
	}

	if err := s.repo.SaveCardUnderUserID(ctx, card); err != nil {
		return nil, ErrFailedToCreate
	}

	// Log audit event
	if s.audit != nil {
		_ = s.audit.LogAuthEvent(ctx, userID, card.ID, actionInput.IPAddress, actionInput.UserAgent, "", audit.ActionTypeAddCard)
	}

	return card, nil
}

// UpdateCardInput represents the input for updating a card
type UpdateCardInput struct {
	PAN               string
	CardholderName    string
	BankName          string
	PaymentMethodType PaymentMethodType
	CardBrand         string
	ProductName       string
	LinkedPhoneNumber string
	ExpMonth          int
	ExpYear           int
	Cvv               int
}

// UpdateCard updates an existing card (with authorization check)
func (s *CardService) UpdateCard(ctx context.Context, userID, cardID string, input UpdateCardInput, actionInput CardActionInput) (*Card, error) {
	// Verify card belongs to user
	existingCard, err := s.repo.GetCardByID(ctx, cardID)
	if err != nil {
		return nil, ErrFailedToFetch
	}
	if existingCard == nil {
		return nil, ErrCardNotFound
	}
	if existingCard.UserID != userID {
		return nil, ErrNotAuthorized
	}

	// Validate input
	createInput := CreateCardInput{
		PAN:               input.PAN,
		CardholderName:    input.CardholderName,
		BankName:          input.BankName,
		PaymentMethodType: input.PaymentMethodType,
		CardBrand:         input.CardBrand,
		ProductName:       input.ProductName,
		LinkedPhoneNumber: input.LinkedPhoneNumber,
		ExpMonth:          input.ExpMonth,
		ExpYear:           input.ExpYear,
		Cvv:               input.Cvv,
	}
	if err := s.validateCardInput(createInput); err != nil {
		return nil, err
	}

	cardWithPAN, err := s.repo.GetCardByPAN(ctx, input.PAN)
	if err != nil {
		return nil, ErrFailedToUpdate
	}
	if cardWithPAN != nil && cardWithPAN.ID != cardID {
		return nil, ErrDuplicatePAN
	}

	card := &Card{
		ID:                cardID,
		UserID:            userID,
		PAN:               input.PAN,
		CardholderName:    input.CardholderName,
		BankName:          input.BankName,
		PaymentMethodType: input.PaymentMethodType,
		CardBrand:         input.CardBrand,
		ProductName:       input.ProductName,
		LinkedPhoneNumber: input.LinkedPhoneNumber,
		ExpMonth:          input.ExpMonth,
		ExpYear:           input.ExpYear,
		Cvv:               input.Cvv,
	}

	if err := s.repo.UpdateCardDetails(ctx, card); err != nil {
		return nil, ErrFailedToUpdate
	}

	// Log audit event
	if s.audit != nil {
		_ = s.audit.LogAuthEvent(ctx, userID, cardID, actionInput.IPAddress, actionInput.UserAgent, "", audit.ActionTypeUPDATECard)
	}

	return s.repo.GetCardByID(ctx, cardID)
}

// DeleteCard deletes a card (with authorization check)
func (s *CardService) DeleteCard(ctx context.Context, userID, cardID string, actionInput CardActionInput) error {
	// Verify card belongs to user
	existingCard, err := s.repo.GetCardByID(ctx, cardID)
	if err != nil {
		return ErrFailedToFetch
	}
	if existingCard == nil {
		return ErrCardNotFound
	}
	if existingCard.UserID != userID {
		return ErrNotAuthorized
	}

	if err := s.repo.DeleteCard(ctx, cardID, userID); err != nil {
		return ErrFailedToDelete
	}

	// Log audit event
	if s.audit != nil {
		_ = s.audit.LogAuthEvent(ctx, userID, cardID, actionInput.IPAddress, actionInput.UserAgent, "", audit.ActionTypeDELETECard)
	}

	return nil
}
