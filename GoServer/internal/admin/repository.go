package admin

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"webauthn-server/internal/audit"
	authrepo "webauthn-server/internal/auth"
	cardrepo "webauthn-server/internal/card"
	"webauthn-server/internal/timeutil"
	"webauthn-server/internal/user"
)

// Repository defines persistence operations used by the admin workflow.
type Repository interface {
	GetAdminUsers(ctx context.Context) ([]user.User, error)
	GetUsers(ctx context.Context, search string, limit, offset int) ([]user.User, int, error)
	FindUserByID(ctx context.Context, userID string) (*user.User, error)
	UpdateUserStatus(ctx context.Context, userID, status string) error
	DeleteAuthenticatorsByUserID(ctx context.Context, userID string) error
	DeleteCardsByUserID(ctx context.Context, userID string) error
	DeleteAuthLogsByUserID(ctx context.Context, userID string) error
	DeleteUserDataAndLog(ctx context.Context, adminID, adminName, targetID, targetName, action, details string) error
	GetAuthenticatorsForUser(ctx context.Context, userID string) ([]authrepo.Authenticator, error)
	GetAuthenticatorCount(ctx context.Context, userID string) (int, error)
	DeleteAuthenticatorByID(ctx context.Context, authID, userID string) (bool, error)
	DeleteCard(ctx context.Context, cardID, userID string) error
	GetCardsLinkedToUserID(ctx context.Context, userID string) ([]*cardrepo.Card, error)
	DeleteUser(ctx context.Context, userID string) error
	LogAdminAction(ctx context.Context, adminID, adminName, targetID, targetName, action, details string) error
	GetAdminLogsByTargetUser(ctx context.Context, targetID, targetName string) ([]audit.AdminAuditLog, error)
	GetAuthLogsByUserID(ctx context.Context, userID string) ([]audit.AuditLog, error)
}

// SQLiteRepository coordinates the existing SQLite repositories for admin operations.
type SQLiteRepository struct {
	userRepo  user.UserRepository
	authRepo  authrepo.Repository
	cardRepo  cardrepo.Repository
	auditRepo audit.Repository
	pool      *sql.DB
}

// NewSQLiteRepository creates an admin repository from the existing domain repositories.
func NewSQLiteRepository(userRepo user.UserRepository, authRepo authrepo.Repository, cardRepo cardrepo.Repository, auditRepo audit.Repository, pool *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{userRepo: userRepo, authRepo: authRepo, cardRepo: cardRepo, auditRepo: auditRepo, pool: pool}
}

func (r *SQLiteRepository) GetAdminUsers(ctx context.Context) ([]user.User, error) {
	return r.userRepo.GetAdminUsers(ctx)
}

func (r *SQLiteRepository) GetUsers(ctx context.Context, search string, limit, offset int) ([]user.User, int, error) {
	return r.userRepo.GetUsers(ctx, search, limit, offset)
}

func (r *SQLiteRepository) FindUserByID(ctx context.Context, userID string) (*user.User, error) {
	return r.userRepo.FindUserByID(ctx, userID)
}

func (r *SQLiteRepository) UpdateUserStatus(ctx context.Context, userID, status string) error {
	return r.userRepo.UpdateUserStatus(ctx, userID, status)
}

func (r *SQLiteRepository) DeleteAuthenticatorsByUserID(ctx context.Context, userID string) error {
	return r.authRepo.DeleteAuthenticatorsByUserID(ctx, userID)
}

func (r *SQLiteRepository) DeleteCardsByUserID(ctx context.Context, userID string) error {
	return r.cardRepo.DeleteCardsByUserID(ctx, userID)
}

func (r *SQLiteRepository) DeleteAuthLogsByUserID(ctx context.Context, userID string) error {
	return r.auditRepo.DeleteAuthLogsByUserID(ctx, userID)
}

func (r *SQLiteRepository) DeleteUserDataAndLog(ctx context.Context, adminID, adminName, targetID, targetName, action, details string) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	rollback := func() {
		_ = tx.Rollback()
	}

	for _, statement := range []string{
		"DELETE FROM cards WHERE user_id = ?",
		"DELETE FROM audit_logs WHERE user_id = ?",
		"DELETE FROM authenticators WHERE user_id = ?",
		"DELETE FROM users WHERE id = ?",
	} {
		if _, err := tx.ExecContext(ctx, statement, targetID); err != nil {
			rollback()
			return err
		}
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO admin_audit_logs (id, admin_user_id, admin_user_name, target_user_id, target_user_name, action_type, details, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, uuid.NewString(), adminID, adminName, targetID, targetName, action, details, timeutil.Now())
	if err != nil {
		rollback()
		return err
	}

	return tx.Commit()
}

func (r *SQLiteRepository) GetAuthenticatorsForUser(ctx context.Context, userID string) ([]authrepo.Authenticator, error) {
	return r.authRepo.GetAuthenticatorsForUser(ctx, userID)
}

func (r *SQLiteRepository) GetAuthenticatorCount(ctx context.Context, userID string) (int, error) {
	return r.authRepo.GetAuthenticatorCount(ctx, userID)
}

func (r *SQLiteRepository) DeleteAuthenticatorByID(ctx context.Context, authID, userID string) (bool, error) {
	return r.authRepo.DeleteAuthenticatorByID(ctx, authID, userID)
}

func (r *SQLiteRepository) DeleteCard(ctx context.Context, cardID, userID string) error {
	return r.cardRepo.DeleteCard(ctx, cardID, userID)
}

func (r *SQLiteRepository) GetCardsLinkedToUserID(ctx context.Context, userID string) ([]*cardrepo.Card, error) {
	return r.cardRepo.GetCardsLinkedToUserId(ctx, userID)
}

func (r *SQLiteRepository) DeleteUser(ctx context.Context, userID string) error {
	return r.userRepo.DeleteUser(ctx, userID)
}

func (r *SQLiteRepository) LogAdminAction(ctx context.Context, adminID, adminName, targetID, targetName, action, details string) error {
	return r.auditRepo.LogAdminAction(ctx, adminID, adminName, targetID, targetName, action, details)
}

func (r *SQLiteRepository) GetAdminLogsByTargetUser(ctx context.Context, targetID, targetName string) ([]audit.AdminAuditLog, error) {
	return r.auditRepo.GetAdminLogsByTargetUser(ctx, targetID, targetName)
}

func (r *SQLiteRepository) GetAuthLogsByUserID(ctx context.Context, targetID string) ([]audit.AuditLog, error) {
	return r.auditRepo.GetAuthLogsByUserID(ctx, targetID)
}
