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
	LLMProvider           string
	GeminiAPIKey          string
	GeminiModel           string
	OllamaBaseURL         string
	OllamaModel           string
	EmailAPI              string
	IMAPHost              string
	IMAPAppPassword       string
	TemporalAddress       string
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
		LLMProvider:           getEnv("LLM_PROVIDER", "ollama"),
		GeminiAPIKey:          getEnv("GEMINI_API_KEY", ""),
		GeminiModel:           getEnv("GEMINI_MODEL", "gemini-2.0-flash"),
		OllamaBaseURL:         getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
		OllamaModel:           getEnv("OLLAMA_MODEL", "llama3.2"),
		EmailAPI:              getEnv("EMAIL_API", "imp"),
		IMAPHost:              getEnv("IMAP_HOST", "imap.gmail.com:993"),
		IMAPAppPassword:       getEnv("IMAP_APP_PASSWORD", ""),
		TemporalAddress:       getEnv("TEMPORAL_ADDRESS", "localhost:7233"),
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
