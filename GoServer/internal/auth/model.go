package auth

import (
    "time"

    "github.com/golang-jwt/jwt/v5"
)

// Authenticator represents a WebAuthn authenticator in the domain
// owned by the auth package.
type Authenticator struct {
    ID             string     `json:"id"`
    CredentialID   string     `json:"credential_id"`
    PublicKey      []byte     `json:"public_key"`
    Counter        int        `json:"counter"`
    UserID         string     `json:"user_id"`
    Transports     []string   `json:"transports"`
    AAGUID         *string    `json:"aaguid"`
    DeviceType     *string    `json:"device_type"`
    BackedUp       *bool      `json:"backed_up"`
    AttachmentType *string    `json:"attachment_type"`
    Nickname       *string    `json:"nickname"`
    LastUsedAt     *time.Time `json:"last_used_at"`
    CreatedAt      time.Time  `json:"created_at"`
    LastLocation   *string    `json:"last_location,omitempty"`
}

// SessionClaims represents short-lived session JWT claims used for WebAuthn cookie flow.
type SessionClaims struct {
    SessionData string `json:"session_data"`
    jwt.RegisteredClaims
}
