package main

import (
	"context"
	"log"
	"time"

	"github.com/joho/godotenv"

	"webauthn-server/internal/admin"
	"webauthn-server/internal/api"
	"webauthn-server/internal/audit"
	"webauthn-server/internal/auth"
	"webauthn-server/internal/card"
	"webauthn-server/internal/config"
	"webauthn-server/internal/db"
	"webauthn-server/internal/user"
)

func main() {
	// Load .env file if exists
	_ = godotenv.Load()

	// Load configuration
	cfg := config.Load()

	// Initialize database
	database, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize database schema
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := database.InitializeDatabase(ctx); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize WebAuthn service
	webAuthnService, err := auth.NewWebAuthnService(&auth.Config{
		RPID:      cfg.RPID,
		RPName:    cfg.RPName,
		RPOrigins: []string{cfg.RPOrigin},
	})
	if err != nil {
		log.Fatalf("Failed to initialize WebAuthn service: %v", err)
	}

	// Initialize repository and handlers
	authRepo := auth.NewSQLiteRepository(database.Pool())
	userRepo := user.NewSQLiteUserRepository(database.Pool())
	auditRepo := audit.NewSQLiteRepository(database.Pool())
	cardRepo := card.NewSQLiteCardRepository(database.Pool())

	authHandler, err := auth.NewAuthHandler(authRepo, userRepo, auditRepo, webAuthnService, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize auth handler: %v", err)
	}

	cardService := card.NewCardService(cardRepo, userRepo, auditRepo)
	cardHandler := card.NewCardHandler(cardService, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize card handler: %v", err)
	}
	adminRepo := admin.NewSQLiteRepository(userRepo, authRepo, cardRepo, auditRepo)
	adminHandler := admin.NewHandler(admin.NewService(adminRepo))

	// Setup router via API package
	router := api.SetupRouter(authHandler, adminHandler, cardHandler, userRepo, cfg.ClientOrigin, cfg.JWTSecret)

	// Start server
	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	log.Printf("WebAuthn relying party: %s", cfg.RPOrigin)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
