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
)

// Claims represents the JWT claims structure
// Username is exposed to client, UserID is kept server-side for auth lookups
type Claims struct {
	Username string `json:"username"`
	UserID   string `json:"-"` // Server-side only, not exposed in token
	jwt.RegisteredClaims
}

// JWTAuth middleware validates JWT tokens from cookie or header
func JWTAuth(secret string) gin.HandlerFunc {
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		// Parse and validate token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}

// SaveAuthToken stores the JWT in a cookie
func SaveAuthToken(c *gin.Context, tokenString string) {
	c.SetCookie("auth_token", tokenString, 86400, "/", "", false, true)
}

// GetAuthToken reads the JWT from cookie
func GetAuthToken(c *gin.Context) string {
	cookie, err := c.Cookie("auth_token")
	if err != nil {
		return ""
	}
	return cookie
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("userId")
	if !exists {
		return ""
	}
	return userID.(string)
}

// GetUsername extracts username from context
func GetUsername(c *gin.Context) string {
	username, exists := c.Get("username")
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
func SaveWebAuthnSession(c *gin.Context, session webauthn.SessionData, jwtSecret string) error {
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return err
	}

	claims := SessionClaims{
		SessionData: string(sessionJSON),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return err
	}

	c.SetCookie("webauthn_session", tokenString, 300, "/", "", false, true)
	return nil
}