package card

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"webauthn-server/internal/config"
)

type CardHandler struct {
	service *CardService
	cfg     *config.Config
}

func NewCardHandler(service *CardService, cfg *config.Config) *CardHandler {
	return &CardHandler{
		service: service,
		cfg:     cfg,
	}
}

// GetCards retrieves all cards linked to the authenticated user
func (h *CardHandler) GetCards(c *gin.Context) {
	userID, err := h.service.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	cards, err := h.service.GetAllCards(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cards": cards})
}

// CreateCardRequest represents the request body for creating a card
type CreateCardRequest struct {
	PAN                string `json:"pan" binding:"required"`
	CardholderName     string `json:"cardholder_name" binding:"required"`
	BankName           string `json:"bank_name" binding:"required"`
	PaymentMethodType  string `json:"payment_method_type" binding:"required"`
	CardBrand          string `json:"card_brand" binding:"required"`
	ExpMonth           int    `json:"exp_month" binding:"required,min=1,max=12"`
	ExpYear            int    `json:"exp_year" binding:"required,min=2024"`
	Cvv                int    `json:"cvv" binding:"required,min=100,max=9999"`
}

// CreateCard creates a new card for the authenticated user
func (h *CardHandler) CreateCard(c *gin.Context) {
	userID, err := h.service.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := CreateCardInput{
		PAN:                req.PAN,
		CardholderName:     req.CardholderName,
		BankName:           req.BankName,
		PaymentMethodType:  PaymentMethodType(req.PaymentMethodType),
		CardBrand:          req.CardBrand,
		ExpMonth:           req.ExpMonth,
		ExpYear:            req.ExpYear,
		Cvv:                req.Cvv,
	}

	actionInput := CardActionInput{
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}

	card, err := h.service.CreateCard(c.Request.Context(), userID, input, actionInput)
	if err != nil {
		switch err {
		case ErrInvalidPAN, ErrInvalidCVV, ErrInvalidExpiry, ErrInvalidCardBrand,
			ErrInvalidPaymentMethod, ErrInvalidExpMonth, ErrInvalidExpYear:
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
	PAN                string `json:"pan" binding:"required"`
	CardholderName     string `json:"cardholder_name" binding:"required"`
	BankName           string `json:"bank_name" binding:"required"`
	PaymentMethodType  string `json:"payment_method_type" binding:"required"`
	CardBrand          string `json:"card_brand" binding:"required"`
	ExpMonth           int    `json:"exp_month" binding:"required,min=1,max=12"`
	ExpYear            int    `json:"exp_year" binding:"required,min=2024"`
	Cvv                int    `json:"cvv" binding:"required,min=100,max=9999"`
}

// UpdateCard updates an existing card
func (h *CardHandler) UpdateCard(c *gin.Context) {
	userID, err := h.service.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	cardID := c.Param("id")
	if cardID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "card id is required"})
		return
	}

	var req UpdateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := UpdateCardInput{
		PAN:                req.PAN,
		CardholderName:     req.CardholderName,
		BankName:           req.BankName,
		PaymentMethodType:  PaymentMethodType(req.PaymentMethodType),
		CardBrand:          req.CardBrand,
		ExpMonth:           req.ExpMonth,
		ExpYear:            req.ExpYear,
		Cvv:                req.Cvv,
	}

	actionInput := CardActionInput{
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}

	card, err := h.service.UpdateCard(c.Request.Context(), userID, cardID, input, actionInput)
	if err != nil {
		switch err {
		case ErrCardNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		case ErrNotAuthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to update this card"})
		case ErrInvalidPAN, ErrInvalidCVV, ErrInvalidExpiry, ErrInvalidCardBrand,
			ErrInvalidPaymentMethod, ErrInvalidExpMonth, ErrInvalidExpYear:
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	cardID := c.Param("id")
	if cardID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "card id is required"})
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
			c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		case ErrNotAuthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to delete this card"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "card deleted successfully"})
}