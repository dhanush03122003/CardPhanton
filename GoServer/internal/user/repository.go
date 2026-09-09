package user

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"webauthn-server/internal/timeutil"
)

// UserRepository defines DB operations needed by the user domain.
type UserRepository interface {
	FindUserByUsername(ctx context.Context, username string) (*User, error)
	FindUserByID(ctx context.Context, id string) (*User, error)
	GetAdminUsers(ctx context.Context) ([]User, error)
	GetUsers(ctx context.Context, search string, limit, offset int) ([]User, int, error)
	CreateUser(ctx context.Context, username string) (*User, error)
	CreateUserWithID(ctx context.Context, id, username, role, status string) (*User, error)
	UpdateUserStatus(ctx context.Context, userID, status string) error
	GetPendingUsers(ctx context.Context) ([]User, error)
	DeleteUser(ctx context.Context, userID string) error
}

// SQLiteUserRepository is a concrete implementation of UserRepository using database/sql.
type SQLiteUserRepository struct {
	pool *sql.DB
}

// NewSQLiteUserRepository creates a new repository with the provided pool.
func NewSQLiteUserRepository(pool *sql.DB) *SQLiteUserRepository {
	return &SQLiteUserRepository{pool: pool}
}

// FindUserByUsername finds a user by username.
func (r *SQLiteUserRepository) FindUserByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := r.pool.QueryRowContext(ctx,
		"SELECT id, username, role, status, created_at FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.Role, &user.Status, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByID finds a user by ID.
func (r *SQLiteUserRepository) FindUserByID(ctx context.Context, id string) (*User, error) {
	var user User
	err := r.pool.QueryRowContext(ctx,
		"SELECT id, username, role, status, created_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.Role, &user.Status, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetAdminUsers returns all administrator accounts without user search/pagination.
func (r *SQLiteUserRepository) GetAdminUsers(ctx context.Context) ([]User, error) {
	rows, err := r.pool.QueryContext(ctx, "SELECT id, username, role, status, created_at FROM users WHERE role = ? ORDER BY created_at ASC", "ADMIN")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var item User
		if err := rows.Scan(&item.ID, &item.Username, &item.Role, &item.Status, &item.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, item)
	}
	return users, rows.Err()
}

// GetUsers returns a searchable paginated page of normal users.
func (r *SQLiteUserRepository) GetUsers(ctx context.Context, search string, limit, offset int) ([]User, int, error) {
	pattern := "%" + search + "%"
	var total int
	if err := r.pool.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE role != ? AND (username LIKE ? OR id LIKE ?)", "ADMIN", pattern, pattern).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.QueryContext(ctx, "SELECT id, username, role, status, created_at FROM users WHERE role != ? AND (username LIKE ? OR id LIKE ?) ORDER BY created_at ASC LIMIT ? OFFSET ?", "ADMIN", pattern, pattern, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var item User
		if err := rows.Scan(&item.ID, &item.Username, &item.Role, &item.Status, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, item)
	}
	return users, total, rows.Err()
}

// CreateUser creates a new user (Postgres generates the UUID).
func (r *SQLiteUserRepository) CreateUser(ctx context.Context, username string) (*User, error) {
	return r.CreateUserWithID(ctx, uuid.NewString(), username, "USER", StatusActive)
}

// CreateUserWithID creates a new user using a specific pre-generated UUID (stateless flow).
func (r *SQLiteUserRepository) CreateUserWithID(ctx context.Context, id, username, role, status string) (*User, error) {
	createdAt := timeutil.Now()
	_, err := r.pool.ExecContext(ctx,
		"INSERT INTO users (id, username, role, status, created_at) VALUES (?, ?, ?, ?, ?)",
		id, username, role, status, createdAt,
	)
	if err != nil {
		return nil, err
	}
	return &User{ID: id, Username: username, Role: role, Status: status, CreatedAt: createdAt}, nil
}

// UpdateUserStatus updates a user's approval status.
func (r *SQLiteUserRepository) UpdateUserStatus(ctx context.Context, userID, status string) error {
	_, err := r.pool.ExecContext(ctx, "UPDATE users SET status = ? WHERE id = ?", status, userID)
	return err
}

// GetPendingUsers returns users waiting for admin approval.
func (r *SQLiteUserRepository) GetPendingUsers(ctx context.Context) ([]User, error) {
	rows, err := r.pool.QueryContext(ctx, "SELECT id, username, role, status, created_at FROM users WHERE status = ? ORDER BY created_at ASC", StatusPendingApproval)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var pending User
		if err := rows.Scan(&pending.ID, &pending.Username, &pending.Role, &pending.Status, &pending.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, pending)
	}
	return users, rows.Err()
}

// DeleteUser deletes a user by ID.
func (r *SQLiteUserRepository) DeleteUser(ctx context.Context, userID string) error {
	_, err := r.pool.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
	return err
}
