package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration values
type Config struct {
	DatabaseURL           string
	JWTSecret             string
	RPID                  string
	RPName                string
	RPOrigin              string
	Port                  string
	ClientOrigin          string
	Environment           string   // "PROD" or "DEV", defaults to "DEV"
	AuthTokenExpiry       int      // in seconds
	WebAuthnSessionExpiry int      // in seconds
	AdminUsernames        []string // list of usernames with admin role
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		DatabaseURL:           getEnv("DATABASE_URL", "postgres://webauthn:webauthn_password@localhost:5432/webauthn"),
		JWTSecret:             getEnv("JWT_SECRET", "default-secret"),
		RPID:                  getEnv("RP_ID", "localhost"),
		RPName:                getEnv("RP_NAME", "WebAuthn App"),
		RPOrigin:              getEnv("RP_ORIGIN", "http://localhost:5173"),
		Port:                  getEnv("PORT", "8080"),
		ClientOrigin:          getEnv("CLIENT_ORIGIN", "http://localhost:5173"),
		Environment:           getEnv("ENVIRONMENT", "DEV"),
		AuthTokenExpiry:       getEnvInt("AUTH_TOKEN_EXPIRY", 86400),
		WebAuthnSessionExpiry: getEnvInt("WEBAUTHN_SESSION_EXPIRY", 300),
		AdminUsernames:        parseAdminUsernames(getEnv("ADMIN_USERNAMES", "")),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// parseAdminUsernames parses comma-separated admin usernames and returns them as lowercase list
func parseAdminUsernames(adminUsernamesStr string) []string {
	if adminUsernamesStr == "" {
		return []string{}
	}

	parts := strings.Split(adminUsernamesStr, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, strings.ToLower(trimmed))
		}
	}
	return result
}
