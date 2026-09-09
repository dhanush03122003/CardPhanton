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
	ID             string    `json:"id"`
	AdminUserID    string    `json:"admin_user_id"`
	AdminUserName  string    `json:"admin_user_name"`
	TargetUserID   string    `json:"target_user_id"`
	TargetUserName string    `json:"target_user_name"`
	ActionType     string    `json:"action_type"`
	Details        string    `json:"details"`
	CreatedAt      time.Time `json:"created_at"`
}

type ActionType string

const (
	ActionTypeLogin    ActionType = "LOGIN"
	ActionTypeRegister ActionType = "CREATED"
	ActionTypeDelete   ActionType = "DELETED"

	ActionTypeAddCard    ActionType = "ADD CARD"
	ActionTypeUPDATECard ActionType = "MODIFY CARD"
	ActionTypeDELETECard ActionType = "DELETE CARD"

	ActionTypePasskeyAdded      ActionType = "PASSKEY_ADDED"
	ActionTypePasskeyDeleted    ActionType = "PASSKEY_DELETED"
	ActionTypeAccountCreated    ActionType = "ACCOUNT_CREATED"
	ActionTypeSentForApproval   ActionType = "SENT_FOR_APPROVAL"
	AdminActionUserApproved     ActionType = "USER_APPROVED"
	AdminActionApprovalReceived ActionType = "APPROVAL_RECEIVED"
	AdminActionUserRejected     ActionType = "USER_REJECTED"
	AdminActionUserSuspended    ActionType = "USER_SUSPENDED"
	AdminActionUserReactivated  ActionType = "USER_REACTIVATED"
	AdminActionUserDeleted      ActionType = "USER_DELETED"
	AdminActionAuthDeleted      ActionType = "AUTH_DELETED"
	AdminActionCardDeleted      ActionType = "CARD_DELETED"
)
