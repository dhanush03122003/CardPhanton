package api

import (
    "github.com/gin-gonic/gin"

    "webauthn-server/internal/auth"
    "webauthn-server/internal/api/middleware"
)

// SetupRouter initializes Gin, applies middleware, and mounts routes
func SetupRouter(authHandler *auth.AuthHandler, clientOrigin string, jwtSecret string) *gin.Engine {
    router := gin.Default()

    // CORS
    router.Use(middleware.CorsMiddleware(clientOrigin))

    apiGroup := router.Group("/api")
    {
        auth := apiGroup.Group("/auth")
        {
            // Public
            auth.GET("/generate-registration-options", authHandler.GenerateRegistrationOptions)
            auth.POST("/verify-registration", authHandler.VerifyRegistration)
            auth.GET("/generate-authentication-options", authHandler.GenerateAuthenticationOptions)
            auth.POST("/verify-authentication", authHandler.VerifyAuthentication)
            auth.POST("/verify-token", authHandler.VerifyToken)
            auth.GET("/generate-conditional-options", authHandler.GenerateConditionalOptions)

            // Protected
            protected := auth.Group("")
            protected.Use(middleware.JWTAuth(jwtSecret))
            {
                protected.GET("/user", authHandler.GetUser)
                protected.GET("/authenticators", authHandler.GetAuthenticators)
                protected.GET("/me", authHandler.GetMe)
                protected.GET("/generate-additional-device-options", authHandler.GenerateAdditionalDeviceOptions)
                protected.DELETE("/authenticator/:id", authHandler.DeleteAuthenticator)
                protected.PUT("/authenticator/:id/nickname", authHandler.UpdateAuthenticatorNickname)
            }
        }

    }

    return router
}
