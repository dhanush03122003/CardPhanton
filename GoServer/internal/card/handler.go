package card

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"webauthn-server/internal/apierrors"
	"webauthn-server/internal/config"
)

type CardHandler struct {
	service  *CardService
	cfg      *config.Config
	workflow WorkflowStarter
}

type WorkflowStarter interface {
	StartLimitExtraction(ctx context.Context, cardID, pan, bankName, cardLast4, email string) (workflowID string, runID string, err error)
}

func NewCardHandler(service *CardService, cfg *config.Config, workflow WorkflowStarter) *CardHandler {
	return &CardHandler{
		service:  service,
		cfg:      cfg,
		workflow: workflow,
	}
}

// GetCards retrieves all cards linked to the authenticated user
func (h *CardHandler) GetCards(c *gin.Context) {
	userID, err := h.service.GetUserIDFromContext(c)
	if err != nil {
		apierrors.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", string(ErrMsgUnauthorized))
		return
	}

	cards, err := h.service.GetAllCards(c.Request.Context(), userID)
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, "CARDS_FETCH_FAILED", "Cards Fetch Failed", "The cards could not be loaded.")
		return
	}

	c.JSON(http.StatusOK, gin.H{"cards": cards})
}

// GetGlobalCards handles GET /api/cards for authenticated users.
func (h *CardHandler) GetGlobalCards(c *gin.Context) {
	cards, err := h.service.GetGlobalCards(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cards": cards})
}

func (h *CardHandler) RefreshCard(c *gin.Context) {
	if h.workflow == nil {
		apierrors.Error(c, http.StatusServiceUnavailable, "WORKFLOW_UNAVAILABLE", "Workflow Unavailable", "The credit-limit workflow is not configured.")
		return
	}
	userID, err := h.service.GetUserIDFromContext(c)
	if err != nil {
		apierrors.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", string(ErrMsgUnauthorized))
		return
	}
	card, err := h.service.GetCardForUser(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		switch err {
		case ErrCardNotFound:
			apierrors.Error(c, http.StatusNotFound, "CARD_NOT_FOUND", "Card Not Found", string(ErrMsgCardNotFound))
		case ErrNotAuthorized:
			apierrors.Error(c, http.StatusForbidden, "CARD_ACCESS_FORBIDDEN", "Card Access Forbidden", string(ErrMsgNotAuthorizedUpdate))
		default:
			apierrors.Error(c, http.StatusInternalServerError, "CARD_LOOKUP_FAILED", "Card Lookup Failed", "The card could not be loaded.")
		}
		return
	}
	if len(card.PAN) < 4 {
		apierrors.Error(c, http.StatusUnprocessableEntity, "INVALID_CARD_NUMBER", "Invalid Card Number", "The card does not contain a valid last-four value.")
		return
	}
	if card.Email == "" {
		apierrors.Error(c, http.StatusUnprocessableEntity, "CARD_EMAIL_REQUIRED", "Card Email Required", "Add an email address to this card before refreshing its limits.")
		return
	}
	workflowID, runID, err := h.workflow.StartLimitExtraction(c.Request.Context(), card.ID, card.PAN, card.BankName, card.PAN[len(card.PAN)-4:], card.Email)
	if err != nil {
		apierrors.Error(c, http.StatusBadGateway, "WORKFLOW_START_FAILED", "Workflow Start Failed", "The credit-limit workflow could not be started.")
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"card_id":     card.ID,
		"workflow_id": workflowID,
		"run_id":      runID,
	})
}

// CreateCardRequest represents the request body for creating a card
type CreateCardRequest struct {
	PAN               string   `json:"pan" binding:"required"`
	CardholderName    string   `json:"cardholder_name" binding:"required"`
	BankName          string   `json:"bank_name" binding:"required"`
	Email             string   `json:"email"`
	PaymentMethodType string   `json:"payment_method_type" binding:"required"`
	CardBrand         string   `json:"card_brand" binding:"required"`
	ProductName       string   `json:"product_name"`
	LinkedPhoneNumber string   `json:"linked_phone_number" binding:"required"`
	ExpMonth          int      `json:"exp_month" binding:"required,min=1,max=12"`
	ExpYear           int      `json:"exp_year" binding:"required,min=2024"`
	Cvv               int      `json:"cvv" binding:"required,min=100,max=9999"`
	TotalLimit        *float64 `json:"total_limit"`
	AvailableLimit    *float64 `json:"available_limit"`
}

// CreateCard creates a new card for the authenticated user
func (h *CardHandler) CreateCard(c *gin.Context) {
	userID, err := h.service.GetUserIDFromContext(c)
	if err != nil {
		apierrors.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", string(ErrMsgUnauthorized))
		return
	}

	var req CreateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.Error(c, http.StatusBadRequest, "INVALID_CARD_REQUEST", "Invalid Card Request", "The card request is invalid.")
		return
	}

	input := CreateCardInput{
		PAN:               req.PAN,
		CardholderName:    req.CardholderName,
		BankName:          req.BankName,
		Email:             req.Email,
		PaymentMethodType: PaymentMethodType(req.PaymentMethodType),
		CardBrand:         req.CardBrand,
		ProductName:       req.ProductName,
		LinkedPhoneNumber: req.LinkedPhoneNumber,
		ExpMonth:          req.ExpMonth,
		ExpYear:           req.ExpYear,
		Cvv:               req.Cvv,
		TotalLimit:        req.TotalLimit,
		AvailableLimit:    req.AvailableLimit,
	}

	actionInput := CardActionInput{
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}

	card, err := h.service.CreateCard(c.Request.Context(), userID, input, actionInput)
	if err != nil {
		switch err {
		case ErrInvalidPAN, ErrInvalidCVV, ErrInvalidExpiry, ErrInvalidCardBrand,
			ErrInvalidPaymentMethod, ErrInvalidExpMonth, ErrInvalidExpYear,
			ErrInvalidLimit,
			ErrInvalidLinkedPhoneNumber:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case ErrDuplicatePAN:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"card": card})
}

// UpdateCardRequest represents the request body for updating a card
type UpdateCardRequest struct {
	PAN               string   `json:"pan" binding:"required"`
	CardholderName    string   `json:"cardholder_name" binding:"required"`
	BankName          string   `json:"bank_name" binding:"required"`
	Email             string   `json:"email"`
	PaymentMethodType string   `json:"payment_method_type" binding:"required"`
	CardBrand         string   `json:"card_brand" binding:"required"`
	ProductName       string   `json:"product_name"`
	LinkedPhoneNumber string   `json:"linked_phone_number" binding:"required"`
	ExpMonth          int      `json:"exp_month" binding:"required,min=1,max=12"`
	ExpYear           int      `json:"exp_year" binding:"required,min=2024"`
	Cvv               int      `json:"cvv" binding:"required,min=100,max=9999"`
	TotalLimit        *float64 `json:"total_limit"`
	AvailableLimit    *float64 `json:"available_limit"`
}

// UpdateCard updates an existing card
func (h *CardHandler) UpdateCard(c *gin.Context) {
	userID, err := h.service.GetUserIDFromContext(c)
	if err != nil {
		apierrors.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", string(ErrMsgUnauthorized))
		return
	}

	cardID := c.Param("id")
	if cardID == "" {
		apierrors.Error(c, http.StatusBadRequest, "CARD_ID_REQUIRED", "Card ID Required", string(ErrMsgCardIDRequired))
		return
	}

	var req UpdateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.Error(c, http.StatusBadRequest, "INVALID_CARD_REQUEST", "Invalid Card Request", "The card request is invalid.")
		return
	}

	input := UpdateCardInput{
		PAN:               req.PAN,
		CardholderName:    req.CardholderName,
		BankName:          req.BankName,
		Email:             req.Email,
		PaymentMethodType: PaymentMethodType(req.PaymentMethodType),
		CardBrand:         req.CardBrand,
		ProductName:       req.ProductName,
		LinkedPhoneNumber: req.LinkedPhoneNumber,
		ExpMonth:          req.ExpMonth,
		ExpYear:           req.ExpYear,
		Cvv:               req.Cvv,
		TotalLimit:        req.TotalLimit,
		AvailableLimit:    req.AvailableLimit,
	}

	actionInput := CardActionInput{
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}

	card, err := h.service.UpdateCard(c.Request.Context(), userID, cardID, input, actionInput)
	if err != nil {
		switch err {
		case ErrCardNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": ErrMsgCardNotFound})
		case ErrNotAuthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": ErrMsgNotAuthorizedUpdate})
		case ErrDuplicatePAN:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case ErrInvalidPAN, ErrInvalidCVV, ErrInvalidExpiry, ErrInvalidCardBrand,
			ErrInvalidPaymentMethod, ErrInvalidExpMonth, ErrInvalidExpYear,
			ErrInvalidLimit,
			ErrInvalidLinkedPhoneNumber:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"card": card})
}

// DeleteCard deletes a card
func (h *CardHandler) DeleteCard(c *gin.Context) {
	userID, err := h.service.GetUserIDFromContext(c)
	if err != nil {
		apierrors.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", string(ErrMsgUnauthorized))
		return
	}

	cardID := c.Param("id")
	if cardID == "" {
		apierrors.Error(c, http.StatusBadRequest, "CARD_ID_REQUIRED", "Card ID Required", string(ErrMsgCardIDRequired))
		return
	}

	actionInput := CardActionInput{
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}

	err = h.service.DeleteCard(c.Request.Context(), userID, cardID, actionInput)
	if err != nil {
		switch err {
		case ErrCardNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": ErrMsgCardNotFound})
		case ErrNotAuthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": ErrMsgNotAuthorizedDelete})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "card deleted successfully"})
}
