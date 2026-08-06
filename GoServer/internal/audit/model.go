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
