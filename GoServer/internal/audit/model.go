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

// AdminAuditLog records an administrative action independently of user activity logs.
type AdminAuditLog struct {
	ID           string    `json:"id"`
	AdminUserID  string    `json:"admin_user_id"`
	TargetUserID string    `json:"target_user_id"`
	ActionType   string    `json:"action_type"`
	Details      string    `json:"details"`
	CreatedAt    time.Time `json:"created_at"`
}

type ActionType string

// 2. Define the enum constants
const (
	ActionTypeLogin    ActionType = "LOGIN"
	ActionTypeRegister ActionType = "CREATED"
	ActionTypeDelete   ActionType = "DELETED"

	ActionTypeAddCard    ActionType = "ADD CARD"
	ActionTypeUPDATECard ActionType = "MODIFY CARD"
	ActionTypeDELETECard ActionType = "DELETE CARD"

	ActionTypePasskeyAdded   ActionType = "PASSKEY_ADDED"
	ActionTypePasskeyDeleted ActionType = "PASSKEY_DELETED"
	ActionTypeUserApproved   ActionType = "USER_APPROVED"
	ActionTypeUserRejected   ActionType = "USER_REJECTED"

	AdminActionUserApproved  ActionType = "USER_APPROVED"
	AdminActionUserSuspended ActionType = "USER_SUSPENDED"
	AdminActionUserDeleted   ActionType = "USER_DELETED"
	AdminActionAuthDeleted   ActionType = "AUTH_DELETED"
)
