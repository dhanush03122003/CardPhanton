package audit

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"webauthn-server/internal/timeutil"
)

// Repository defines audit persistence operations.
type Repository interface {
	LogAuthEvent(ctx context.Context, userID, credentialID, ipAddress, userAgent, location string, actionType ActionType) error
	DeleteAuthLogsByUserID(ctx context.Context, userID string) error
	LogAdminAction(ctx context.Context, adminID, adminName, targetID, targetName, action, details string) error
	GetAdminLogsByTargetUser(ctx context.Context, targetID, targetName string) ([]AdminAuditLog, error)
	GetAuthLogsByUserID(ctx context.Context, userID string) ([]AuditLog, error)
}

// DeleteAuthLogsByUserID removes user activity while preserving admin audit logs.
func (r *SQLiteRepository) DeleteAuthLogsByUserID(ctx context.Context, userID string) error {
	_, err := r.pool.ExecContext(ctx, "DELETE FROM audit_logs WHERE user_id = ?", userID)
	return err
}

// LogAdminAction stores an administrative action independently from user activity.
func (r *SQLiteRepository) LogAdminAction(ctx context.Context, adminID, adminName, targetID, targetName, action, details string) error {
	_, err := r.pool.ExecContext(ctx, `
		INSERT INTO admin_audit_logs (id, admin_user_id, admin_user_name, target_user_id, target_user_name, action_type, details, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, uuid.NewString(), adminID, adminName, targetID, targetName, action, details, timeutil.Now())
	return err
}

// GetAdminLogsByTargetUser returns every admin action for a target user.
// Matching by name preserves history across re-registration; matching by ID
// keeps older records readable when the username column was not populated.
func (r *SQLiteRepository) GetAdminLogsByTargetUser(ctx context.Context, targetID, targetName string) ([]AdminAuditLog, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT id, admin_user_id, admin_user_name, target_user_id, target_user_name, action_type, details, created_at
		FROM admin_audit_logs
		WHERE target_user_name = ? OR target_user_id = ?
		ORDER BY created_at DESC
	`, targetName, targetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AdminAuditLog
	for rows.Next() {
		var log AdminAuditLog
		if err := rows.Scan(&log.ID, &log.AdminUserID, &log.AdminUserName, &log.TargetUserID, &log.TargetUserName, &log.ActionType, &log.Details, &log.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

// GetAuthLogsByUserID returns standard activity logs for a user.
func (r *SQLiteRepository) GetAuthLogsByUserID(ctx context.Context, userID string) ([]AuditLog, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT id, user_id, credential_id, ip_address, location, user_agent, action_type, login_time
		FROM audit_logs WHERE user_id = ? ORDER BY login_time DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		if err := rows.Scan(&log.ID, &log.UserID, &log.CredentialID, &log.IPAddress, &log.Location, &log.UserAgent, &log.ActionType, &log.LoginTime); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

// SQLiteRepository stores audit logs in SQLite.
type SQLiteRepository struct {
	pool *sql.DB
}

// NewSQLiteRepository creates a new audit repository.
func NewSQLiteRepository(pool *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{pool: pool}
}

// LogAuthEvent logs an authentication event and updates last_used_at when applicable.
func (r *SQLiteRepository) LogAuthEvent(ctx context.Context, userID, credentialID, ipAddress, userAgent, location string, actionType ActionType) error {
	now := timeutil.Now()
	_, err := r.pool.ExecContext(ctx, `
        INSERT INTO audit_logs (id, user_id, credential_id, ip_address, location, user_agent, action_type, login_time) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `, uuid.NewString(), userID, credentialID, ipAddress, location, userAgent, actionType, now)
	if err != nil {
		return err
	}

	if actionType == ActionTypeLogin {
		_, err = r.pool.ExecContext(ctx, `
            UPDATE authenticators 
            SET last_used_at = ? 
            WHERE credential_id = ?
		`, now, credentialID)
	}

	return err
}
