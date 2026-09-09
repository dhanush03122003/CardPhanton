package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/oschwald/geoip2-golang"

	"webauthn-server/internal/api/middleware"
	"webauthn-server/internal/apierrors"
	"webauthn-server/internal/audit"
	"webauthn-server/internal/config"
	"webauthn-server/internal/timeutil"
	"webauthn-server/internal/user"
)

// TokenClaims is the JWT claims used for application tokens.
// User identity and role are resolved from the database using Username.
type TokenClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthHandler handles WebAuthn authentication requests
type AuthHandler struct {
	repo      Repository
	userRepo  user.UserRepository
	audit     audit.Repository
	webAuthn  *WebAuthnService
	cfg       *config.Config
	jwtSecret []byte
	geoIP     *geoip2.Reader
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(repo Repository, userRepo user.UserRepository, auditRepo audit.Repository, webAuthn *WebAuthnService, cfg *config.Config) (*AuthHandler, error) {
	jwtSecret := []byte(cfg.JWTSecret)

	// Try to load GeoIP database if available
	var geoIP *geoip2.Reader
	geoIP, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Printf("GeoIP database not found, location tracking will be disabled: %v", err)
	}

	return &AuthHandler{
		repo:      repo,
		userRepo:  userRepo,
		audit:     auditRepo,
		webAuthn:  webAuthn,
		cfg:       cfg,
		jwtSecret: jwtSecret,
		geoIP:     geoIP,
	}, nil
}

// GenerateRegistrationOptions handles GET /generate-registration-options
func (h *AuthHandler) GenerateRegistrationOptions(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrMsgUsernameRequired})
		return
	}

	ctx := c.Request.Context()

	userObj, err := h.userRepo.FindUserByUsername(ctx, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrMsgDatabase})
		return
	}

	if userObj != nil {
		c.JSON(http.StatusConflict, gin.H{"error": ErrMsgUsernameTaken})
		return
	}

	userIDStr := uuid.New().String()
	existingCreds := make([]Credential, 0)

	regData, err := h.webAuthn.BeginRegistration(userIDStr, username, existingCreds)
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, "DATABASE_ERROR", "Database Error", "The request could not be completed.")
		return
	}

	err = middleware.SaveWebAuthnSession(c, regData.Session, string(h.jwtSecret), h.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrMsgSaveSession})
		return
	}

	c.JSON(http.StatusOK, regData.Options)
}

// VerifyRegistration handles POST /verify-registration
func (h *AuthHandler) VerifyRegistration(c *gin.Context) {
	var body struct {
		Username     string          `json:"username" binding:"required"`
		Verification json.RawMessage `json:"verification" binding:"required"`
		Nickname     string          `json:"nickname"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		apierrors.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid Request", "The request body is invalid.")
		return
	}

	ctx := c.Request.Context()

	sessionData, err := middleware.GetWebAuthnSession(c, string(h.jwtSecret))
	if err != nil || sessionData == nil {
		apierrors.Error(c, http.StatusUnauthorized, "REGISTRATION_SESSION_EXPIRED", "Registration Session Expired", string(ErrMsgRegistrationSession))
		return
	}

	parsedUUID, err := uuid.FromBytes(sessionData.UserID)
	if err != nil {
		apierrors.Error(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid User ID", string(ErrMsgInvalidUserID))
		return
	}
	userIDStr := parsedUUID.String()

	userForAuth := NewUser(userIDStr, body.Username)

	parsedResponse, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(body.Verification))
	if err != nil {
		apierrors.Error(c, http.StatusBadRequest, "INVALID_WEBAUTHN_RESPONSE", "Invalid WebAuthn Response", string(ErrMsgParseWebAuthnResponse)+err.Error())
		return
	}

	credential, err := h.webAuthn.CreateCredential(userForAuth, *sessionData, parsedResponse)
	if err != nil {
		apierrors.Error(c, http.StatusBadRequest, "BIOMETRIC_VERIFICATION_FAILED", "Biometric Verification Failed", string(ErrMsgBiometricVerification)+err.Error())
		return
	}

	userRow, err := h.userRepo.FindUserByID(ctx, userIDStr)
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, "DATABASE_ERROR", "Database Error", string(ErrMsgDatabase))
		return
	}

	approvalRequired := false
	accountCreated := false
	if userRow == nil {
		// Determine user role: check if username is in admin list (case-insensitive)
		role := "USER"
		for _, adminName := range h.cfg.AdminUsernames {
			if strings.EqualFold(body.Username, adminName) {
				role = "ADMIN"
				break
			}
		}

		status := user.StatusPendingApproval
		if role == "ADMIN" {
			status = user.StatusActive
		} else {
			approvalRequired = true
		}

		userRow, err = h.userRepo.CreateUserWithID(ctx, userIDStr, body.Username, role, status)
		if err != nil {
			apierrors.Error(c, http.StatusConflict, "USERNAME_TAKEN", "Username Already Taken", string(ErrMsgUsernameTaken))
			return
		}
		accountCreated = true
	}

	attachmentType := "cross-platform"
	if credential.Transport != nil {
		for _, t := range credential.Transport {
			if string(t) == "internal" {
				attachmentType = "platform"
				break
			}
		}
	}

	deviceType := "singleDevice"
	backedUp := credential.Flags.BackupState

	var aaguidStr *string
	if len(credential.Authenticator.AAGUID) == 16 {
		if parsedAAGUID, err := uuid.FromBytes(credential.Authenticator.AAGUID); err == nil {
			str := parsedAAGUID.String()
			aaguidStr = &str
		}
	}

	transports := []string{"internal"}
	if len(credential.Transport) > 0 {
		transports = make([]string, len(credential.Transport))
		for i, t := range credential.Transport {
			transports[i] = string(t)
		}
	}
	log.Printf("registration debug: credentialID=%s aaguidBytes=%x aaguidStr=%v attachmentType=%s transports=%v", base64.RawURLEncoding.EncodeToString(credential.ID), credential.Authenticator.AAGUID, aaguidStr, attachmentType, transports)

	nickname := body.Nickname
	if nickname == "" {
		nickname = "Unnamed Passkey"
	}

	err = h.repo.SaveAuthenticator(
		ctx,
		userIDStr,
		base64.RawURLEncoding.EncodeToString(credential.ID),
		credential.PublicKey,
		int(credential.Authenticator.SignCount),
		transports,
		aaguidStr,
		&deviceType,
		&backedUp,
		&attachmentType,
		&nickname,
	)
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, "AUTHENTICATOR_SAVE_FAILED", "Authenticator Save Failed", string(ErrMsgSaveAuthenticator))
		return
	}

	ip := c.ClientIP()
	if xForwardedFor := c.GetHeader("X-Forwarded-For"); xForwardedFor != "" {
		parts := strings.Split(xForwardedFor, ",")
		ip = strings.TrimSpace(parts[0])
	}
	location := h.getLocation(ip)
	userAgent := c.GetHeader("User-Agent")

	credentialID := base64.RawURLEncoding.EncodeToString(credential.ID)
	if accountCreated {
		_ = h.audit.LogAuthEvent(ctx, userIDStr, credentialID, ip, userAgent, location, audit.ActionTypeAccountCreated)
		if approvalRequired {
			_ = h.audit.LogAuthEvent(ctx, userIDStr, credentialID, ip, userAgent, location, audit.ActionTypeSentForApproval)
		}
	}
	_ = h.audit.LogAuthEvent(ctx, userIDStr, credentialID, ip, userAgent, location, audit.ActionTypePasskeyAdded)

	if approvalRequired {
		_ = h.audit.LogAdminAction(
			ctx,
			"SYSTEM",
			"SYSTEM",
			userRow.ID,
			userRow.Username,
			string(audit.AdminActionApprovalReceived),
			"New user registration received and is awaiting admin approval",
		)
	}

	// Registration successful - user must login separately
	// Clear WebAuthn session to ensure fresh state for login
	middleware.ClearWebAuthnSession(c, h.cfg)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"username": body.Username,
	})
}

// GenerateAuthenticationOptions handles GET /generate-authentication-options
func (h *AuthHandler) GenerateAuthenticationOptions(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrMsgUsernameRequired})
		return
	}

	ctx := c.Request.Context()

	userRow, err := h.userRepo.FindUserByUsername(ctx, username)
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, "WEBAUTHN_ERROR", "WebAuthn Error", "The WebAuthn operation could not be completed.")
		return
	}
	if userRow == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrMsgUserNotFound})
		return
	}

	authenticators, err := h.repo.GetAuthenticatorsForUser(ctx, userRow.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userForAuth := NewUser(userRow.ID, userRow.Username)

	for _, a := range authenticators {
		rawIDBytes, err := base64.RawURLEncoding.DecodeString(a.CredentialID)
		if err != nil {
			continue
		}

		userForAuth.Credentials = append(userForAuth.Credentials, webauthn.Credential{ID: rawIDBytes})
	}

	loginData, err := h.webAuthn.BeginLoginDirect(userForAuth)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = middleware.SaveWebAuthnSession(c, loginData.Session, string(h.jwtSecret), h.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrMsgSaveSession})
		return
	}

	c.JSON(http.StatusOK, loginData.Options)
}

// VerifyAuthentication handles POST /verify-authentication
func (h *AuthHandler) VerifyAuthentication(c *gin.Context) {
	var body struct {
		Username     string          `json:"username"`
		Verification json.RawMessage `json:"verification" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		apierrors.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid Request", "The request body is invalid.")
		return
	}

	ctx := c.Request.Context()

	sessionData, err := middleware.GetWebAuthnSession(c, string(h.jwtSecret))
	if err != nil || sessionData == nil {
		apierrors.Error(c, http.StatusUnauthorized, "AUTHENTICATION_SESSION_EXPIRED", "Authentication Session Expired", string(ErrMsgAuthenticationSession))
		return
	}

	parsedResponse, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(body.Verification))
	if err != nil {
		apierrors.Error(c, http.StatusBadRequest, "INVALID_WEBAUTHN_RESPONSE", "Invalid WebAuthn Response", string(ErrMsgInvalidResponseFormat))
		return
	}

	credentialIDStr := parsedResponse.ID

	rawIDBytes, err := base64.RawURLEncoding.DecodeString(credentialIDStr)
	if err != nil {
		apierrors.Error(c, http.StatusBadRequest, "INVALID_CREDENTIAL_ID", "Invalid Credential ID", string(ErrMsgInvalidCredentialID))
		return
	}

	authenticator, err := h.repo.FindAuthenticatorByCredentialID(ctx, credentialIDStr)
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, "PASSKEY_LOOKUP_FAILED", "Passkey Lookup Failed", string(ErrMsgPasskeyLookup))
		return
	}
	if authenticator == nil {
		apierrors.Error(c, http.StatusBadRequest, "PASSKEY_NOT_RECOGNIZED", "Passkey Not Recognized", string(ErrMsgPasskeyNotRecognized))
		return
	}

	userRow, err := h.userRepo.FindUserByID(ctx, authenticator.UserID)
	if err != nil || userRow == nil {
		apierrors.Error(c, http.StatusBadRequest, "USER_NOT_FOUND", "User Not Found", string(ErrMsgUserNotFound))
		return
	}
	if userRow.Status == user.StatusPendingApproval {
		apierrors.Error(c, http.StatusForbidden, "PENDING_ADMIN_APPROVAL", "Account Pending Approval", string(ErrMsgPendingApproval))
		return
	}
	if userRow.Status == user.StatusSuspended {
		apierrors.Error(c, http.StatusForbidden, "ACCOUNT_SUSPENDED", "Account Suspended", string(ErrMsgSuspended))
		return
	}
	if userRow.Status != user.StatusActive {
		apierrors.Error(c, http.StatusForbidden, "ACCOUNT_UNAVAILABLE", "Account Unavailable", "This account cannot authenticate in its current state.")
		return
	}

	userForAuth := NewUser(userRow.ID, userRow.Username)
	incomingFlags := parsedResponse.Response.AuthenticatorData.Flags

	userForAuth.Credentials = []webauthn.Credential{
		{
			ID:        rawIDBytes,
			PublicKey: authenticator.PublicKey,
			Authenticator: webauthn.Authenticator{
				SignCount: uint32(authenticator.Counter),
			},
			Flags: webauthn.CredentialFlags{
				BackupEligible: incomingFlags.HasBackupEligible(),
				BackupState:    incomingFlags.HasBackupState(),
			},
		},
	}

	sessionData.UserID = userForAuth.WebAuthnID()
	sessionData.AllowedCredentialIDs = [][]byte{rawIDBytes}

	credential, err := h.webAuthn.ValidateLogin(userForAuth, *sessionData, parsedResponse)
	if err != nil {
		apierrors.Error(c, http.StatusBadRequest, "AUTHENTICATION_FAILED", "Authentication Failed", string(ErrMsgAuthenticationFailed))
		return
	}

	err = h.repo.UpdateAuthenticatorCounter(ctx, credentialIDStr, int(credential.Authenticator.SignCount))
	if err != nil {
		log.Printf("Failed to update counter: %v", err)
	}

	ip := c.ClientIP()
	if xForwardedFor := c.GetHeader("X-Forwarded-For"); xForwardedFor != "" {
		parts := strings.Split(xForwardedFor, ",")
		ip = strings.TrimSpace(parts[0])
	}
	location := h.getLocation(ip)
	userAgent := c.GetHeader("User-Agent")

	err = h.audit.LogAuthEvent(ctx, userRow.ID, credentialIDStr, ip, userAgent, location, audit.ActionTypeLogin)
	if err != nil {
		log.Printf("Failed to log auth event: %v", err)
	}

	// Clear WebAuthn session after successful authentication
	middleware.ClearWebAuthnSession(c, h.cfg)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &TokenClaims{
		Username: userRow.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(timeutil.Now()),
			ExpiresAt: jwt.NewNumericDate(timeutil.Now().Add(time.Duration(h.cfg.AuthTokenExpiry) * time.Second)),
		},
	})

	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrMsgGenerateToken})
		return
	}

	// Set token in cookie using helper function (expiry from AUTH_TOKEN_EXPIRY env var)
	middleware.SaveAuthToken(c, tokenString, h.cfg)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"username": userRow.Username,
	})
}

// GetUser handles GET /user - JWT protected
func (h *AuthHandler) GetUser(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrMsgUnauthorized})
		return
	}

	ctx := c.Request.Context()
	userRow, err := h.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, "DATABASE_ERROR", "Database Error", "The request could not be completed.")
		return
	}
	if userRow == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrMsgUserNotFound})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": userRow.ID, "username": userRow.Username})
}

// getUserIDFromContext retrieves the database user ID populated by JWTAuth.
func (h *AuthHandler) getUserIDFromContext(c *gin.Context) (string, error) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return "", fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}

// GetAuthenticators handles GET /authenticators - JWT protected
func (h *AuthHandler) GetAuthenticators(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrMsgUnauthorized})
		return
	}

	ctx := c.Request.Context()
	authenticators, err := h.repo.GetAuthenticatorsForUser(ctx, userID)
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, "WEBAUTHN_ERROR", "WebAuthn Error", "The WebAuthn operation could not be completed.")
		return
	}

	c.JSON(http.StatusOK, authenticators)
}

// DeleteAuthenticator handles DELETE /authenticators/:id - JWT protected
func (h *AuthHandler) DeleteAuthenticator(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrMsgUnauthorized})
		return
	}

	authID := c.Param("id")
	if authID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrMsgAuthenticatorID})
		return
	}

	ctx := c.Request.Context()

	authenticators, err := h.repo.GetAuthenticatorsForUser(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var credentialID string
	for _, a := range authenticators {
		if a.ID == authID {
			credentialID = a.CredentialID
			break
		}
	}

	deleted, err := h.repo.DeleteAuthenticatorByID(ctx, authID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrMsgAuthenticatorNotFound})
		return
	}

	if credentialID != "" {
		ip := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		_ = h.audit.LogAuthEvent(ctx, userID, credentialID, ip, userAgent, "", audit.ActionTypePasskeyDeleted)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// UpdateAuthenticatorNickname handles PUT /authenticators/:id - JWT protected
func (h *AuthHandler) UpdateAuthenticatorNickname(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrMsgUnauthorized})
		return
	}

	authID := c.Param("id")
	if authID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrMsgAuthenticatorID})
		return
	}

	var body struct {
		Nickname string `json:"nickname" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	err = h.repo.UpdateAuthenticatorNickname(ctx, authID, userID, body.Nickname)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetMe handles GET /me - JWT protected
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrMsgUnauthorized})
		return
	}

	ctx := c.Request.Context()
	userRow, err := h.userRepo.FindUserByID(ctx, userID)
	if err != nil || userRow == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrMsgUserDeleted})
		return
	}
	authenticators, err := h.repo.GetAuthenticatorsForUser(ctx, userID)
	if err != nil {
		authenticators = []Authenticator{}
	}

	type AuthResponse struct {
		ID             string     `json:"id"`
		CredentialID   string     `json:"credentialId"`
		Counter        int        `json:"counter"`
		CreatedAt      time.Time  `json:"createdAt"`
		AttachmentType *string    `json:"attachmentType"`
		DeviceType     *string    `json:"deviceType"`
		BackedUp       *bool      `json:"backedUp"`
		AAGUID         *string    `json:"aaguid"`
		Nickname       *string    `json:"nickname"`
		LastUsedAt     *time.Time `json:"lastUsedAt"`
		Location       string     `json:"location"`
	}

	var authResponses []AuthResponse
	for _, authn := range authenticators {
		loc := "Unknown Location"
		if authn.LastLocation != nil {
			loc = *authn.LastLocation
		}
		authResponses = append(authResponses, AuthResponse{
			ID:             authn.ID,
			CredentialID:   authn.CredentialID,
			Counter:        authn.Counter,
			CreatedAt:      authn.CreatedAt,
			AttachmentType: authn.AttachmentType,
			DeviceType:     authn.DeviceType,
			BackedUp:       authn.BackedUp,
			AAGUID:         authn.AAGUID,
			Nickname:       authn.Nickname,
			LastUsedAt:     authn.LastUsedAt,
			Location:       loc,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":       userRow.ID,
			"username": userRow.Username,
		},
		"authenticators": authResponses,
	})
}

// GenerateAdditionalDeviceOptions handles GET /generate-additional-device-options - JWT protected
func (h *AuthHandler) GenerateAdditionalDeviceOptions(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrMsgUnauthorized})
		return
	}

	ctx := c.Request.Context()
	userRow, err := h.userRepo.FindUserByID(ctx, userID)
	if err != nil || userRow == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrMsgUserNotFound})
		return
	}

	authenticators, err := h.repo.GetAuthenticatorsForUser(ctx, userID)
	var existingCreds []Credential
	if err == nil {
		for _, a := range authenticators {
			existingCreds = append(existingCreds, Credential{ID: a.CredentialID, Transports: a.Transports})
		}
	}

	regData, err := h.webAuthn.BeginRegistration(userRow.ID, userRow.Username, existingCreds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = middleware.SaveWebAuthnSession(c, regData.Session, string(h.jwtSecret), h.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrMsgSaveSession})
		return
	}

	c.JSON(http.StatusOK, regData.Options)
}

// VerifyToken handles POST /verify-token
func (h *AuthHandler) VerifyToken(c *gin.Context) {
	var body struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrMsgTokenRequired})
		return
	}

	token, err := jwt.ParseWithClaims(body.Token, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return h.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrMsgInvalidToken})
		return
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrMsgInvalidTokenClaims})
		return
	}

	userRow, err := h.userRepo.FindUserByUsername(c.Request.Context(), claims.Username)
	if err != nil || userRow == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrMsgUserNotFound})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": gin.H{"id": userRow.ID, "username": userRow.Username}})
}

// GenerateConditionalOptions handles GET /generate-conditional-options
func (h *AuthHandler) GenerateConditionalOptions(c *gin.Context) {
	loginData, err := h.webAuthn.BeginDiscoverableLogin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = middleware.SaveWebAuthnSession(c, loginData.Session, string(h.jwtSecret), h.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrMsgSaveSession})
		return
	}

	c.JSON(http.StatusOK, loginData.Options)
}

// getLocation looks up the IP address location
func (h *AuthHandler) getLocation(ip string) string {
	if h.geoIP == nil {
		return "Unknown Location"
	}

	if ip == "::1" || ip == "127.0.0.1" {
		return "Local Development (Localhost)"
	}

	record, err := h.geoIP.City(net.ParseIP(ip))
	if err != nil {
		return "Unknown Location"
	}

	country := ""
	if record.Country.Names != nil && record.Country.Names["en"] != "" {
		country = record.Country.Names["en"]
	}

	location := ""
	if record.City.Names != nil && record.City.Names["en"] != "" {
		location = record.City.Names["en"]
	}

	if location != "" && country != "" {
		return location + ", " + country
	}
	if country != "" {
		return country
	}

	return "Unknown Location"
}

// Logout handles POST /logout - clears authentication cookies
func (h *AuthHandler) Logout(c *gin.Context) {
	// Clear the auth token cookie using helper function
	middleware.ClearAuthToken(c, h.cfg)

	// Clear the WebAuthn session cookie using helper function
	middleware.ClearWebAuthnSession(c, h.cfg)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "logged out successfully",
	})
}
