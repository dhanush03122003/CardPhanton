package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"webauthn-server/internal/audit"
	authrepo "webauthn-server/internal/auth"
	cardrepo "webauthn-server/internal/card"
	"webauthn-server/internal/user"
)

var ErrUserNotFound = errors.New("user not found")

const (
	_ = audit.AdminActionUserApproved
)

type UserDetails struct {
	User           user.User                `json:"user"`
	Authenticators []authrepo.Authenticator `json:"authenticators"`
	UserAuditLogs  []audit.AuditLog         `json:"user_audit_logs"`
	AdminAuditLogs []audit.AdminAuditLog    `json:"admin_audit_logs"`
	Cards          []*cardrepo.Card         `json:"cards"`
}

// Service contains the admin approval workflow.
type Service struct {
	repo Repository
}

// NewService creates an admin service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetAdminUsers(ctx context.Context) ([]user.User, error) {
	return s.repo.GetAdminUsers(ctx)
}

func (s *Service) GetUsers(ctx context.Context, search string, limit, offset int) ([]user.User, int, error) {
	return s.repo.GetUsers(ctx, search, limit, offset)
}

func (s *Service) GetUserDetails(ctx context.Context, userID string) (*UserDetails, error) {
	item, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrUserNotFound
	}
	authenticators, err := s.repo.GetAuthenticatorsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	userLogs, err := s.repo.GetAuthLogsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	adminLogs, err := s.repo.GetAdminLogsByTargetUser(ctx, item.ID, item.Username)
	if err != nil {
		return nil, err
	}
	cards, err := s.repo.GetCardsLinkedToUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if cards == nil {
		cards = []*cardrepo.Card{}
	}
	return &UserDetails{User: *item, Authenticators: authenticators, UserAuditLogs: userLogs, AdminAuditLogs: adminLogs, Cards: cards}, nil
}

func (s *Service) ApproveUser(ctx context.Context, adminID, userID string) error {
	admin, target, err := s.getActionUsers(ctx, adminID, userID)
	if err != nil {
		return err
	}
	if target.Status != user.StatusPendingApproval {
		return ErrInvalidStatusTransition
	}
	if err := s.repo.UpdateUserStatus(ctx, userID, user.StatusActive); err != nil {
		return err
	}
	return s.repo.LogAdminAction(ctx, admin.ID, admin.Username, target.ID, target.Username, string(audit.AdminActionUserApproved), "User approved")
}

func (s *Service) SuspendUser(ctx context.Context, adminID, userID string) error {
	admin, target, err := s.getActionUsers(ctx, adminID, userID)
	if err != nil {
		return err
	}
	if target.Status != user.StatusActive {
		return ErrInvalidStatusTransition
	}
	if err := s.repo.UpdateUserStatus(ctx, userID, user.StatusSuspended); err != nil {
		return err
	}
	return s.repo.LogAdminAction(ctx, admin.ID, admin.Username, target.ID, target.Username, string(audit.AdminActionUserSuspended), "User suspended")
}

func (s *Service) ReactivateUser(ctx context.Context, adminID, userID string) error {
	admin, target, err := s.getActionUsers(ctx, adminID, userID)
	if err != nil {
		return err
	}
	if target.Status != user.StatusSuspended {
		return ErrInvalidStatusTransition
	}
	if err := s.repo.UpdateUserStatus(ctx, userID, user.StatusActive); err != nil {
		return err
	}
	return s.repo.LogAdminAction(ctx, admin.ID, admin.Username, target.ID, target.Username, string(audit.AdminActionUserReactivated), "User re-enabled")
}

func (s *Service) RejectUser(ctx context.Context, adminID, userID string) error {
	admin, target, err := s.getActionUsers(ctx, adminID, userID)
	if err != nil {
		return err
	}
	if target.Status != user.StatusPendingApproval {
		return ErrInvalidStatusTransition
	}
	return s.repo.DeleteUserDataAndLog(ctx, admin.ID, admin.Username, target.ID, target.Username, string(audit.AdminActionUserRejected), "User rejected and account deleted")
}

func (s *Service) DeleteUser(ctx context.Context, adminID, userID string) error {
	admin, target, err := s.getActionUsers(ctx, adminID, userID)
	if err != nil {
		return err
	}
	if target.Status == user.StatusPendingApproval {
		return ErrInvalidStatusTransition
	}
	return s.repo.DeleteUserDataAndLog(ctx, admin.ID, admin.Username, target.ID, target.Username, string(audit.AdminActionUserDeleted), "User deleted")
}

func (s *Service) DeleteAuthenticator(ctx context.Context, adminID, userID, authID string) error {
	admin, target, err := s.getActionUsers(ctx, adminID, userID)
	if err != nil {
		return err
	}
	count, err := s.repo.GetAuthenticatorCount(ctx, userID)
	if err != nil {
		return err
	}
	if count <= 1 {
		return ErrLastAuthenticator
	}
	deleted, err := s.repo.DeleteAuthenticatorByID(ctx, authID, userID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrAuthenticatorNotFound
	}
	return s.repo.LogAdminAction(ctx, admin.ID, admin.Username, target.ID, target.Username, string(audit.AdminActionAuthDeleted), fmt.Sprintf("Authenticator deleted: %s", authID))
}

func (s *Service) DeleteCard(ctx context.Context, adminID, userID, cardID string) error {
	admin, target, err := s.getActionUsers(ctx, adminID, userID)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteCard(ctx, cardID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCardNotFound
		}
		return err
	}
	return s.repo.LogAdminAction(ctx, admin.ID, admin.Username, target.ID, target.Username, string(audit.AdminActionCardDeleted), fmt.Sprintf("Card deleted: %s", cardID))
}

func (s *Service) getActionUsers(ctx context.Context, adminID, targetID string) (*user.User, *user.User, error) {
	admin, err := s.repo.FindUserByID(ctx, adminID)
	if err != nil {
		return nil, nil, err
	}
	if admin == nil {
		return nil, nil, ErrUserNotFound
	}
	target, err := s.repo.FindUserByID(ctx, targetID)
	if err != nil {
		return nil, nil, err
	}
	if target == nil {
		return nil, nil, ErrUserNotFound
	}
	return admin, target, nil
}

func (s *Service) ensureUser(ctx context.Context, userID string) error {
	item, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if item == nil {
		return ErrUserNotFound
	}
	return nil
}

var ErrLastAuthenticator = errors.New("cannot delete the last authenticator")
var ErrAuthenticatorNotFound = errors.New("authenticator not found")
var ErrInvalidStatusTransition = errors.New("invalid user status transition")
var ErrCardNotFound = errors.New("card not found")
