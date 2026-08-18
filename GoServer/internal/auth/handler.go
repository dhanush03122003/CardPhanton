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
	"webauthn-server/internal/audit"
	"webauthn-server/internal/config"
	"webauthn-server/internal/user"
)

// TokenClaims is the JWT claims used for application tokens
// Username is exposed to client, UserID is kept server-side
type TokenClaims struct {
	Username string `json:"username"`
	UserID   string `json:"-"` // Server-side only, not exposed in token
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
        c.JSON(http.StatusBadRequest, gin.H{"error": "Username required"})
        return
    }

    ctx := c.Request.Context()

    userObj, err := h.userRepo.FindUserByUsername(ctx, username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    if userObj != nil {
        c.JSON(http.StatusConflict, gin.H{"error": "Username is already taken. Please log in or choose a different username."})
        return
    }

    userIDStr := uuid.New().String()
    existingCreds := make([]Credential, 0)

    regData, err := h.webAuthn.BeginRegistration(userIDStr, username, existingCreds)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    err = middleware.SaveWebAuthnSession(c, regData.Session, string(h.jwtSecret))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session cookie"})
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
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    ctx := c.Request.Context()

    sessionData, err := middleware.GetWebAuthnSession(c, string(h.jwtSecret))
    if err != nil || sessionData == nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired or missing. Please restart registration."})
        return
    }

    parsedUUID, err := uuid.FromBytes(sessionData.UserID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID in session"})
        return
    }
    userIDStr := parsedUUID.String()

    userForAuth := NewUser(userIDStr, body.Username)

    parsedResponse, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(body.Verification))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse WebAuthn response: " + err.Error()})
        return
    }

    credential, err := h.webAuthn.CreateCredential(userForAuth, *sessionData, parsedResponse)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Biometric verification failed: " + err.Error()})
        return
    }

    userRow, err := h.userRepo.FindUserByID(ctx, userIDStr)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    if userRow == nil {
        userRow, err = h.userRepo.CreateUserWithID(ctx, userIDStr, body.Username)
        if err != nil {
            c.JSON(http.StatusConflict, gin.H{"error": "This username was just taken. Please try again with a different username."})
            return
        }
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
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save authenticator"})
        return
    }

    ip := c.ClientIP()
    if xForwardedFor := c.GetHeader("X-Forwarded-For"); xForwardedFor != "" {
        parts := strings.Split(xForwardedFor, ",")
        ip = strings.TrimSpace(parts[0])
    }
    location := h.getLocation(ip)
    userAgent := c.GetHeader("User-Agent")

    _ = h.audit.LogAuthEvent(ctx, userIDStr, base64.RawURLEncoding.EncodeToString(credential.ID), ip, userAgent, location, audit.ActionTypePasskeyAdded)

	// token := jwt.NewWithClaims(jwt.SigningMethodHS256, &TokenClaims{
	// 	Username: body.Username,
	// 	UserID:   userIDStr,
	// 	RegisteredClaims: jwt.RegisteredClaims{
	// 		IssuedAt:  jwt.NewNumericDate(time.Now()),
	// 		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	// 	},
	// })

	// tokenString, err := token.SignedString(h.jwtSecret)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
	// 	return
	// }

	// Clear WebAuthn session and set auth cookie
	c.SetCookie("webauthn_session", "", -1, "/", "", false, true)
	// c.SetCookie("auth_token", tokenString, 86400, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"username":  body.Username,
	})
}

// GenerateAuthenticationOptions handles GET /generate-authentication-options
func (h *AuthHandler) GenerateAuthenticationOptions(c *gin.Context) {
    username := c.Query("username")
    if username == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Username required"})
        return
    }

    ctx := c.Request.Context()

    userRow, err := h.userRepo.FindUserByUsername(ctx, username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    if userRow == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
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

    err = middleware.SaveWebAuthnSession(c, loginData.Session, string(h.jwtSecret))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session cookie"})
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
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    ctx := c.Request.Context()

    sessionData, err := middleware.GetWebAuthnSession(c, string(h.jwtSecret))
    if err != nil || sessionData == nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired or missing. Please try again."})
        return
    }

    parsedResponse, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(body.Verification))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid response format: " + err.Error()})
        return
    }

    credentialIDStr := parsedResponse.ID

    rawIDBytes, err := base64.RawURLEncoding.DecodeString(credentialIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credential ID encoding"})
        return
    }

    authenticator, err := h.repo.FindAuthenticatorByCredentialID(ctx, credentialIDStr)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error while looking up passkey"})
        return
    }
    if authenticator == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Passkey not recognized"})
        return
    }

    userRow, err := h.userRepo.FindUserByID(ctx, authenticator.UserID)
    if err != nil || userRow == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
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
        c.JSON(http.StatusBadRequest, gin.H{"error": "Authentication failed: " + err.Error()})
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

    c.SetCookie("webauthn_session", "", -1, "/", "", false, true)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &TokenClaims{
		Username: userRow.Username,
		UserID:   userRow.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	})

	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Set token in cookie
	c.SetCookie("auth_token", tokenString, h.cfg.AuthTokenExpiry, "/", "", true, true)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"username": userRow.Username,
	})
}

// GetUser handles GET /user - JWT protected
func (h *AuthHandler) GetUser(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	ctx := c.Request.Context()
	userRow, err := h.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if userRow == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": userRow.ID, "username": userRow.Username})
}

// getUserIDFromContext retrieves userId by looking up username from context
func (h *AuthHandler) getUserIDFromContext(c *gin.Context) (string, error) {
	username, exists := c.Get("username")
	if !exists || username == "" {
		return "", fmt.Errorf("username not found in context")
	}

	usernameStr, ok := username.(string)
	if !ok {
		return "", fmt.Errorf("invalid username type")
	}

	user, err := h.userRepo.FindUserByUsername(c.Request.Context(), usernameStr)
	if err != nil || user == nil {
		return "", fmt.Errorf("user not found")
	}

	return user.ID, nil
}

// GetAuthenticators handles GET /authenticators - JWT protected
func (h *AuthHandler) GetAuthenticators(c *gin.Context) {
    userID, err := h.getUserIDFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    ctx := c.Request.Context()
    authenticators, err := h.repo.GetAuthenticatorsForUser(ctx, userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, authenticators)
}

// DeleteAuthenticator handles DELETE /authenticators/:id - JWT protected
func (h *AuthHandler) DeleteAuthenticator(c *gin.Context) {
    userID, err := h.getUserIDFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    authID := c.Param("id")
    if authID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Authenticator ID required"})
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
        c.JSON(http.StatusNotFound, gin.H{"error": "Authenticator not found"})
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
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    authID := c.Param("id")
    if authID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Authenticator ID required"})
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
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    ctx := c.Request.Context()
    userRow, err := h.userRepo.FindUserByID(ctx, userID)
    if err != nil || userRow == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User no longer exists"})
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
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    ctx := c.Request.Context()
    userRow, err := h.userRepo.FindUserByID(ctx, userID)
    if err != nil || userRow == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
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

    err = middleware.SaveWebAuthnSession(c, regData.Session, string(h.jwtSecret))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session cookie"})
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
        c.JSON(http.StatusBadRequest, gin.H{"error": "Token required"})
        return
    }

    token, err := jwt.ParseWithClaims(body.Token, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
        return h.jwtSecret, nil
    })

    if err != nil || !token.Valid {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
        return
    }

    claims, ok := token.Claims.(*TokenClaims)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
        return
    }

    userRow, err := h.userRepo.FindUserByID(c.Request.Context(), claims.UserID)
    if err != nil || userRow == nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
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

    err = middleware.SaveWebAuthnSession(c, loginData.Session, string(h.jwtSecret))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session cookie"})
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
	// Clear the auth token cookie
	c.SetCookie("auth_token", "", -1, "/", "", true, true)

	// Clear the WebAuthn session cookie
	c.SetCookie("webauthn_session", "", -1, "/", "", true, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "logged out successfully",
	})
}