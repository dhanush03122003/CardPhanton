package admin

import (
	"context"

	"webauthn-server/internal/audit"
	authrepo "webauthn-server/internal/auth"
	cardrepo "webauthn-server/internal/card"
	"webauthn-server/internal/user"
)

// Repository defines persistence operations used by the admin workflow.
type Repository interface {
	GetAdminUsers(ctx context.Context) ([]user.User, error)
	GetUsers(ctx context.Context, search string, limit, offset int) ([]user.User, int, error)
	FindUserByID(ctx context.Context, userID string) (*user.User, error)
	UpdateUserStatus(ctx context.Context, userID, status string) error
	DeleteAuthenticatorsByUserID(ctx context.Context, userID string) error
	GetAuthenticatorsForUser(ctx context.Context, userID string) ([]authrepo.Authenticator, error)
	GetAuthenticatorCount(ctx context.Context, userID string) (int, error)
	DeleteAuthenticatorByID(ctx context.Context, authID, userID string) (bool, error)
	GetCardsLinkedToUserID(ctx context.Context, userID string) ([]*cardrepo.Card, error)
	DeleteUser(ctx context.Context, userID string) error
	LogAdminAction(ctx context.Context, adminID, targetID, action, details string) error
	GetAdminLogsByTargetUser(ctx context.Context, targetID string) ([]audit.AdminAuditLog, error)
	GetAuthLogsByUserID(ctx context.Context, userID string) ([]audit.AuditLog, error)
}

// SQLiteRepository coordinates the existing SQLite repositories for admin operations.
type SQLiteRepository struct {
	userRepo  user.UserRepository
	authRepo  authrepo.Repository
	cardRepo  cardrepo.Repository
	auditRepo audit.Repository
}

// NewSQLiteRepository creates an admin repository from the existing domain repositories.
func NewSQLiteRepository(userRepo user.UserRepository, authRepo authrepo.Repository, cardRepo cardrepo.Repository, auditRepo audit.Repository) *SQLiteRepository {
	return &SQLiteRepository{userRepo: userRepo, authRepo: authRepo, cardRepo: cardRepo, auditRepo: auditRepo}
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

func (r *SQLiteRepository) GetAuthenticatorsForUser(ctx context.Context, userID string) ([]authrepo.Authenticator, error) {
	return r.authRepo.GetAuthenticatorsForUser(ctx, userID)
}

func (r *SQLiteRepository) GetAuthenticatorCount(ctx context.Context, userID string) (int, error) {
	return r.authRepo.GetAuthenticatorCount(ctx, userID)
}

func (r *SQLiteRepository) DeleteAuthenticatorByID(ctx context.Context, authID, userID string) (bool, error) {
	return r.authRepo.DeleteAuthenticatorByID(ctx, authID, userID)
}

func (r *SQLiteRepository) GetCardsLinkedToUserID(ctx context.Context, userID string) ([]*cardrepo.Card, error) {
	return r.cardRepo.GetCardsLinkedToUserId(ctx, userID)
}

func (r *SQLiteRepository) DeleteUser(ctx context.Context, userID string) error {
	return r.userRepo.DeleteUser(ctx, userID)
}

func (r *SQLiteRepository) LogAdminAction(ctx context.Context, adminID, targetID, action, details string) error {
	return r.auditRepo.LogAdminAction(ctx, adminID, targetID, action, details)
}

func (r *SQLiteRepository) GetAdminLogsByTargetUser(ctx context.Context, targetID string) ([]audit.AdminAuditLog, error) {
	return r.auditRepo.GetAdminLogsByTargetUser(ctx, targetID)
}

func (r *SQLiteRepository) GetAuthLogsByUserID(ctx context.Context, targetID string) ([]audit.AuditLog, error) {
	return r.auditRepo.GetAuthLogsByUserID(ctx, targetID)
}
