package audit

import (
    "context"
    "database/sql"
    "time"

    "github.com/google/uuid"
)

// Repository defines audit persistence operations.
type Repository interface {
    LogAuthEvent(ctx context.Context, userID, credentialID, ipAddress, userAgent, location, actionType string) error
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
func (r *SQLiteRepository) LogAuthEvent(ctx context.Context, userID, credentialID, ipAddress, userAgent, location, actionType string) error {
    now := time.Now().UTC()
    _, err := r.pool.ExecContext(ctx, `
        INSERT INTO audit_logs (id, user_id, credential_id, ip_address, location, user_agent, action_type, login_time) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `, uuid.NewString(), userID, credentialID, ipAddress, location, userAgent, actionType, now)
    if err != nil {
        return err
    }

    if actionType != "DELETED" {
        _, err = r.pool.ExecContext(ctx, `
            UPDATE authenticators 
            SET last_used_at = ? 
            WHERE credential_id = ?
        `, credentialID)
    }

    return err
}
