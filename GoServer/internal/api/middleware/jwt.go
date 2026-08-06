package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims structure
type Claims struct {
    UserID   string `json:"userId"`
    Username string `json:"username"`
    jwt.RegisteredClaims
}

// JWTAuth middleware validates JWT tokens
func JWTAuth(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get token from Authorization header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        // Extract token from "Bearer <token>"
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
            c.Abort()
            return
        }

        tokenString := parts[1]

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
