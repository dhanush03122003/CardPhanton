package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"webauthn-server/internal/apierrors"
)

// AdminOnly checks the authenticated user's role from context.
// It requires JWTAuth to run first.
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "ADMIN" {
			apierrors.Error(c, http.StatusForbidden, "ADMIN_REQUIRED", "Administrator Access Required", "Administrator privileges are required for this resource.")
			c.Abort()
			return
		}

		c.Next()
	}
}
