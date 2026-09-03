package auth

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"

	"webauthn-server/internal/timeutil"
)

// Repository defines auth-related persistence operations.
type Repository interface {
	GetAuthenticatorsForUser(ctx context.Context, userID string) ([]Authenticator, error)
	SaveAuthenticator(ctx context.Context, userID string, credentialID string, publicKey []byte, counter int, transports []string, aaguid, deviceType *string, backedUp *bool, attachmentType, nickname *string) error
	FindAuthenticatorByCredentialID(ctx context.Context, credentialID string) (*Authenticator, error)
	UpdateAuthenticatorCounter(ctx context.Context, credentialID string, counter int) error
	DeleteAuthenticatorsByUserID(ctx context.Context, userID string) error
	GetAuthenticatorCount(ctx context.Context, userID string) (int, error)
	DeleteAuthenticatorByID(ctx context.Context, authID, userID string) (bool, error)
	UpdateAuthenticatorNickname(ctx context.Context, authID, userID, newNickname string) error
}

// GetAuthenticatorCount returns the number of passkeys owned by a user.
func (r *SQLiteRepository) GetAuthenticatorCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.pool.QueryRowContext(ctx, "SELECT COUNT(*) FROM authenticators WHERE user_id = ?", userID).Scan(&count)
	return count, err
}

// DeleteAuthenticatorsByUserID deletes all passkeys belonging to a user.
func (r *SQLiteRepository) DeleteAuthenticatorsByUserID(ctx context.Context, userID string) error {
	_, err := r.pool.ExecContext(ctx, "DELETE FROM authenticators WHERE user_id = ?", userID)
	return err
}

// SQLiteRepository is a concrete auth repository backed by database/sql.
type SQLiteRepository struct {
	pool *sql.DB
}

// NewSQLiteRepository creates a new auth repository.
func NewSQLiteRepository(pool *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{pool: pool}
}

// GetAuthenticatorsForUser returns all authenticators for a user.
func (r *SQLiteRepository) GetAuthenticatorsForUser(ctx context.Context, userID string) ([]Authenticator, error) {
	rows, err := r.pool.QueryContext(ctx, `
        SELECT a.id, a.credential_id, a.public_key, a.counter, a.user_id, 
               a.transports, a.aaguid, a.device_type, a.backed_up, a.attachment_type,
               a.nickname, a.last_used_at, a.created_at,
               (SELECT al.location FROM audit_logs al 
                WHERE al.credential_id = a.credential_id 
                ORDER BY al.login_time DESC LIMIT 1) as last_location
        FROM authenticators a 
    WHERE a.user_id = ?
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authenticators []Authenticator
	for rows.Next() {
		var a Authenticator
		var transportsRaw sql.NullString
		var aaguid sql.NullString
		var deviceType sql.NullString
		var backedUp sql.NullBool
		var attachmentType sql.NullString
		var nickname sql.NullString
		var lastUsedAt sql.NullTime
		var lastLocation sql.NullString
		err := rows.Scan(
			&a.ID, &a.CredentialID, &a.PublicKey, &a.Counter, &a.UserID,
			&transportsRaw, &aaguid, &deviceType, &backedUp, &attachmentType,
			&nickname, &lastUsedAt, &a.CreatedAt, &lastLocation,
		)
		if err != nil {
			return nil, err
		}
		if transportsRaw.Valid && transportsRaw.String != "" {
			_ = json.Unmarshal([]byte(transportsRaw.String), &a.Transports)
		}
		if aaguid.Valid {
			a.AAGUID = &aaguid.String
		}
		if deviceType.Valid {
			a.DeviceType = &deviceType.String
		}
		if backedUp.Valid {
			value := backedUp.Bool
			a.BackedUp = &value
		}
		if attachmentType.Valid {
			a.AttachmentType = &attachmentType.String
		}
		if nickname.Valid {
			a.Nickname = &nickname.String
		}
		if lastUsedAt.Valid {
			a.LastUsedAt = &lastUsedAt.Time
		}
		if lastLocation.Valid {
			a.LastLocation = &lastLocation.String
		}
		if len(a.PublicKey) > 0 {
			a.PublicKey = append([]byte(nil), a.PublicKey...)
		}
		authenticators = append(authenticators, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return authenticators, nil
}

// SaveAuthenticator saves a new authenticator.
func (r *SQLiteRepository) SaveAuthenticator(ctx context.Context, userID string, credentialID string, publicKey []byte, counter int, transports []string, aaguid, deviceType *string, backedUp *bool, attachmentType, nickname *string) error {
	transportsJSON, err := json.Marshal(transports)
	if err != nil {
		return err
	}

	now := timeutil.Now()
	_, err = r.pool.ExecContext(ctx, `
        INSERT INTO authenticators (
            id, credential_id, public_key, counter, user_id, transports, 
            aaguid, device_type, backed_up, attachment_type, nickname, last_used_at, created_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `, uuid.NewString(), credentialID, publicKey, counter, userID, string(transportsJSON), aaguid, deviceType, boolPtrToInt(backedUp), attachmentType, nickname, now, now)
	return err
}

// FindAuthenticatorByCredentialID finds an authenticator by credential ID.
func (r *SQLiteRepository) FindAuthenticatorByCredentialID(ctx context.Context, credentialID string) (*Authenticator, error) {
	var a Authenticator
	var transportsRaw sql.NullString
	var aaguid sql.NullString
	var deviceType sql.NullString
	var backedUp sql.NullBool
	var attachmentType sql.NullString
	var nickname sql.NullString
	var lastUsedAt sql.NullTime
	err := r.pool.QueryRowContext(ctx, `
        SELECT id, credential_id, public_key, counter, user_id, 
               transports, aaguid, device_type, backed_up, attachment_type,
               nickname, last_used_at, created_at
        FROM authenticators 
        WHERE credential_id = ?
    `, credentialID).Scan(
		&a.ID, &a.CredentialID, &a.PublicKey, &a.Counter, &a.UserID,
		&transportsRaw, &aaguid, &deviceType, &backedUp, &attachmentType,
		&nickname, &lastUsedAt, &a.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if transportsRaw.Valid && transportsRaw.String != "" {
		_ = json.Unmarshal([]byte(transportsRaw.String), &a.Transports)
	}
	if aaguid.Valid {
		a.AAGUID = &aaguid.String
	}
	if deviceType.Valid {
		a.DeviceType = &deviceType.String
	}
	if backedUp.Valid {
		value := backedUp.Bool
		a.BackedUp = &value
	}
	if attachmentType.Valid {
		a.AttachmentType = &attachmentType.String
	}
	if nickname.Valid {
		a.Nickname = &nickname.String
	}
	if lastUsedAt.Valid {
		a.LastUsedAt = &lastUsedAt.Time
	}
	return &a, nil
}

// UpdateAuthenticatorCounter updates the counter for an authenticator.
func (r *SQLiteRepository) UpdateAuthenticatorCounter(ctx context.Context, credentialID string, counter int) error {
	_, err := r.pool.ExecContext(ctx,
		"UPDATE authenticators SET counter = ? WHERE credential_id = ?",
		counter, credentialID,
	)
	return err
}

// DeleteAuthenticatorByID deletes an authenticator by ID.
func (r *SQLiteRepository) DeleteAuthenticatorByID(ctx context.Context, authID, userID string) (bool, error) {
	result, err := r.pool.ExecContext(ctx,
		"DELETE FROM authenticators WHERE id = ? AND user_id = ?",
		authID, userID,
	)
	if err != nil {
		return false, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

// UpdateAuthenticatorNickname updates the nickname of an authenticator.
func (r *SQLiteRepository) UpdateAuthenticatorNickname(ctx context.Context, authID, userID, newNickname string) error {
	_, err := r.pool.ExecContext(ctx,
		"UPDATE authenticators SET nickname = ? WHERE id = ? AND user_id = ?",
		newNickname, authID, userID,
	)
	return err
}

func boolPtrToInt(value *bool) any {
	if value == nil {
		return nil
	}
	if *value {
		return 1
	}
	return 0
}
