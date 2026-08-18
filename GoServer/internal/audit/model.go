package audit

import "time"

// AuditLog represents an audit log entry
type AuditLog struct {
    ID           string    `json:"id"`
    UserID       string    `json:"user_id"`
    CredentialID string    `json:"credential_id"`
    IPAddress    string    `json:"ip_address"`
    Location     string    `json:"location"`
    UserAgent    string    `json:"user_agent"`
    ActionType   string    `json:"action_type"`
    LoginTime    time.Time `json:"login_time"`
}

type ActionType string

// 2. Define the enum constants
const (
	ActionTypeLogin        ActionType = "LOGIN"
	ActionTypeRegister     ActionType = "CREATED"
	ActionTypeDelete       ActionType = "DELETED"

	ActionTypeAddCard      ActionType = "ADD CARD"
	ActionTypeUPDATECard   ActionType = "MODIFY CARD"
	ActionTypeDELETECard   ActionType = "DELETE CARD"

	ActionTypePasskeyAdded    ActionType = "PASSKEY_ADDED"
	ActionTypePasskeyDeleted  ActionType = "PASSKEY_DELETED"
)