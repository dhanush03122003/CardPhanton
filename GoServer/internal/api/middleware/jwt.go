package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/golang-jwt/jwt/v5"

	"webauthn-server/internal/apierrors"
	"webauthn-server/internal/config"
	"webauthn-server/internal/timeutil"
	"webauthn-server/internal/user"
)

// Claims represents the JWT claims structure.
// Role is intentionally not included; authorization uses the database.
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWTAuth middleware validates JWT tokens from cookie or header
func JWTAuth(secret string, userRepo user.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ""

		// First, try to get token from cookie
		cookie, err := c.Cookie("auth_token")
		if err == nil && cookie != "" {
			tokenString = cookie
		}

		// Fall back to Authorization header
		if tokenString == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString = parts[1]
				}
			}
		}

		if tokenString == "" {
			apierrors.Error(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication Required", "Authorization token required")
			c.Abort()
			return
		}

		// Parse and validate token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			apierrors.Error(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid Token", "The authentication token is invalid or expired.")
			c.Abort()
			return
		}

		userRow, err := userRepo.FindUserByUsername(c.Request.Context(), claims.Username)
		if err != nil || userRow == nil {
			apierrors.Error(c, http.StatusUnauthorized, "USER_NOT_FOUND", "User Not Found", "The authenticated user could not be found.")
			c.Abort()
			return
		}

		// Set authenticated user data once for downstream handlers and middleware.
		c.Set("userId", userRow.ID)
		c.Set("userName", claims.Username)
		c.Set("role", userRow.Role)

		c.Next()
	}
}

// GetUserID extracts the authenticated user's database ID from context.
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("userId")
	if !exists {
		return ""
	}
	userIDString, ok := userID.(string)
	if !ok {
		return ""
	}
	return userIDString
}

// SaveAuthToken stores the JWT in a cookie with configurable expiry
func SaveAuthToken(c *gin.Context, tokenString string, cfg *config.Config) {
	secure := strings.EqualFold(cfg.Environment, "PROD")
	c.SetCookie("auth_token", tokenString, cfg.AuthTokenExpiry, "/", "", secure, true)
}

// ClearAuthToken removes the auth token cookie (used for logout)
func ClearAuthToken(c *gin.Context, cfg *config.Config) {
	secure := strings.EqualFold(cfg.Environment, "PROD")
	c.SetCookie("auth_token", "", -1, "/", "", secure, true)
}

// GetAuthToken reads the JWT from cookie
func GetAuthToken(c *gin.Context) string {
	cookie, err := c.Cookie("auth_token")
	if err != nil {
		return ""
	}
	return cookie
}

// GetUsername extracts userName from context
func GetUsername(c *gin.Context) string {
	username, exists := c.Get("userName")
	if !exists {
		return ""
	}
	return username.(string)
}

// SessionClaims represents short-lived session JWT claims used for WebAuthn cookie flow.
type SessionClaims struct {
	SessionData string `json:"session_data"`
	jwt.RegisteredClaims
}

// GetWebAuthnSession reads WebAuthn session data from the JWT cookie
func GetWebAuthnSession(c *gin.Context, jwtSecret string) (*webauthn.SessionData, error) {
	cookie, err := c.Cookie("webauthn_session")
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(cookie, &SessionClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid or expired session token")
	}

	claims, ok := token.Claims.(*SessionClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	var sessionData webauthn.SessionData
	err = json.Unmarshal([]byte(claims.SessionData), &sessionData)
	if err != nil {
		return nil, err
	}

	return &sessionData, nil
}

// SaveWebAuthnSession stores WebAuthn session data in a JWT cookie
func SaveWebAuthnSession(c *gin.Context, session webauthn.SessionData, jwtSecret string, cfg *config.Config) error {
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return err
	}

	claims := SessionClaims{
		SessionData: string(sessionJSON),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(timeutil.Now().Add(time.Duration(cfg.WebAuthnSessionExpiry) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(timeutil.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return err
	}

	secure := strings.EqualFold(cfg.Environment, "PROD")
	c.SetCookie("webauthn_session", tokenString, cfg.WebAuthnSessionExpiry, "/", "", secure, true)
	return nil
}

// ClearWebAuthnSession removes the WebAuthn session cookie (used after auth completion)
func ClearWebAuthnSession(c *gin.Context, cfg *config.Config) {
	secure := strings.EqualFold(cfg.Environment, "PROD")
	c.SetCookie("webauthn_session", "", -1, "/", "", secure, true)
}
