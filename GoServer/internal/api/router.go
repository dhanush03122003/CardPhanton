package api

import (
	"github.com/gin-gonic/gin"

	"webauthn-server/internal/admin"
	"webauthn-server/internal/api/middleware"
	"webauthn-server/internal/auth"
	"webauthn-server/internal/card"
	"webauthn-server/internal/user"
)

// SetupRouter initializes Gin, applies middleware, and mounts routes
func SetupRouter(authHandler *auth.AuthHandler, adminHandler *admin.Handler, cardHandler *card.CardHandler, userRepo user.UserRepository, clientOrigin string, jwtSecret string) *gin.Engine {
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
			protected.Use(middleware.JWTAuth(jwtSecret, userRepo))
			{
				protected.GET("/user", authHandler.GetUser)
				protected.GET("/authenticators", authHandler.GetAuthenticators)
				protected.GET("/me", authHandler.GetMe)
				protected.GET("/generate-additional-device-options", authHandler.GenerateAdditionalDeviceOptions)
				protected.DELETE("/authenticator/:id", authHandler.DeleteAuthenticator)
				protected.PUT("/authenticator/:id/nickname", authHandler.UpdateAuthenticatorNickname)
				protected.POST("/logout", authHandler.Logout)
			}

		}

		cards := apiGroup.Group("/cards")
		cards.Use(middleware.JWTAuth(jwtSecret, userRepo))
		{
			cards.GET("", cardHandler.GetGlobalCards)
			cards.GET("/mine", cardHandler.GetCards)
			cards.POST("", cardHandler.CreateCard)
			cards.PUT("/:id", cardHandler.UpdateCard)
			cards.DELETE("/:id", cardHandler.DeleteCard)
		}

		admin := apiGroup.Group("/admin")
		admin.Use(middleware.JWTAuth(jwtSecret, userRepo))
		{
			admin.GET("/status", adminHandler.GetStatus)
			admin.GET("/admins", adminHandler.GetAdmins)

			approval := admin.Group("")
			approval.Use(middleware.AdminOnly())
			{
				approval.GET("/users", adminHandler.GetUsers)
				approval.GET("/users/:id/details", adminHandler.GetUserDetails)
				approval.PUT("/users/:id/approve", adminHandler.ApproveUser)
				approval.DELETE("/users/:id/reject", adminHandler.RejectUser)
				approval.PUT("/users/:id/suspend", adminHandler.SuspendUser)
				approval.PUT("/users/:id/reactivate", adminHandler.ReactivateUser)
				approval.DELETE("/users/:id", adminHandler.DeleteUser)
				approval.DELETE("/users/:id/authenticators/:authId", adminHandler.DeleteAuthenticator)
				approval.DELETE("/users/:id/cards/:cardId", adminHandler.DeleteCard)
			}
		}

	}

	return router
}
