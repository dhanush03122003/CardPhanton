package auth

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"webauthn-server/internal/api/middleware"
	"webauthn-server/internal/apierrors"
)

type GoogleAuth struct {
	config *oauth2.Config
}

func newGoogleAuth() *GoogleAuth {
	return &GoogleAuth{
		config: &oauth2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  googleRedirectURL(),
			Scopes:       []string{gmail.GmailReadonlyScope},
			Endpoint:     google.Endpoint,
		},
	}
}

func (h *AuthHandler) GetGoogleConsentURL(c *gin.Context) {
	userID := middleware.GetUserID(c)
	cardID := c.Param("cardID")
	if userID == "" || cardID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user and card are required"})
		return
	}

	cardRecord, err := h.cardRepo.GetCardByID(c.Request.Context(), cardID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load card"})
		return
	}
	if cardRecord == nil || cardRecord.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}
	if strings.TrimSpace(cardRecord.Email) == "" {
		apierrors.Error(c, http.StatusUnprocessableEntity, "CARD_EMAIL_REQUIRED", "Card Email Required", "Add an email address to this card before requesting Google consent.")
		return
	}

	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create OAuth state"})
		return
	}

	state := base64.RawURLEncoding.EncodeToString(stateBytes)
	if err := middleware.SaveGoogleOAuthState(c, state, userID, cardID, string(h.jwtSecret), h.cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save OAuth state"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"consent_required": true,
		"auth_url": h.googleAuth.config.AuthCodeURL(
			state,
			oauth2.AccessTypeOffline,
			oauth2.ApprovalForce,
			oauth2.SetAuthURLParam("include_granted_scopes", "true"),
		),
	})
}

// GoogleOAuthCallback exchanges the authorization code returned by Google.
func (h *AuthHandler) GoogleOAuthCallback(c *gin.Context) {
	if oauthError := c.Query("error"); oauthError != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": oauthError})
		return
	}

	state, err := middleware.GetGoogleOAuthState(c, string(h.jwtSecret))
	if err != nil || state.State != c.Query("state") || state.UserID == "" || state.CardID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid OAuth state"})
		return
	}
	middleware.ClearGoogleOAuthState(c, h.cfg)

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing OAuth code"})
		return
	}

	token, err := h.googleAuth.config.Exchange(c.Request.Context(), code)
	if err != nil {
		log.Printf("Google OAuth token exchange failed: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Google OAuth token exchange failed"})
		return
	}
	if token.RefreshToken == "" {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Google did not return a refresh token"})
		return
	}

	gmailService, err := gmail.NewService(c.Request.Context(), option.WithTokenSource(oauth2.StaticTokenSource(token)))
	if err != nil {
		log.Printf("Google Gmail service creation failed: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "could not initialize Gmail service"})
		return
	}
	profile, err := gmailService.Users.GetProfile("me").Do()
	if err != nil {
		log.Printf("Google Gmail profile lookup failed: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "could not identify Gmail account"})
		return
	}

	if err := h.repo.SaveGoogleOAuthToken(c.Request.Context(), state.UserID, state.CardID, profile.EmailAddress, token.RefreshToken); err != nil {
		log.Printf("Google OAuth token persistence failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save Google account"})
		return
	}

	c.Redirect(http.StatusFound, frontendConsentRedirectURL())
}

func googleRedirectURL() string {
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	return strings.ReplaceAll(redirectURL, "{CLIENT_ORIGIN}", os.Getenv("CLIENT_ORIGIN"))
}

func frontendConsentRedirectURL() string {
	redirectURL := os.Getenv("FRONTEND_REDIRECT_URL")
	if redirectURL == "" {
		redirectURL = os.Getenv("CLIENT_ORIGIN") + "/dashboard?tab=consents"
	}
	return strings.ReplaceAll(redirectURL, "{CLIENT_ORIGIN}", os.Getenv("CLIENT_ORIGIN"))
}

func (h *AuthHandler) GetGoogleConsentCards(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	cards, err := h.repo.GetGoogleConsentCards(c.Request.Context(), userID)
	if err != nil {
		log.Printf("Google consent cards lookup failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load consent status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cards": cards})
}

func (h *AuthHandler) RevokeGoogleConsent(c *gin.Context) {
	userID := middleware.GetUserID(c)
	cardID := c.Param("cardID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	if cardID == "" {
		apierrors.Error(c, http.StatusBadRequest, "CARD_ID_REQUIRED", "Card ID Required", "A card ID is required to revoke consent.")
		return
	}

	if err := h.repo.DeleteGoogleOAuthToken(c.Request.Context(), userID, cardID); err != nil {
		log.Printf("Google consent revocation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not revoke Google consent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "card_id": cardID})
}
