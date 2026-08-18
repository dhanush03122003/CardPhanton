package user

import (
    "context"
    "database/sql"
    "time"

    "github.com/google/uuid"
)

// UserRepository defines DB operations needed by the user domain.
type UserRepository interface {
    FindUserByUsername(ctx context.Context, username string) (*User, error)
    FindUserByID(ctx context.Context, id string) (*User, error)
    CreateUser(ctx context.Context, username string) (*User, error)
    CreateUserWithID(ctx context.Context, id, username string) (*User, error)
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
        "SELECT id, username, created_at FROM users WHERE username = ?",
        username,
    ).Scan(&user.ID, &user.Username, &user.CreatedAt)

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
        "SELECT id, username, created_at FROM users WHERE id = ?",
        id,
    ).Scan(&user.ID, &user.Username, &user.CreatedAt)

    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// CreateUser creates a new user (Postgres generates the UUID).
func (r *SQLiteUserRepository) CreateUser(ctx context.Context, username string) (*User, error) {
    return r.CreateUserWithID(ctx, uuid.NewString(), username)
}

// CreateUserWithID creates a new user using a specific pre-generated UUID (stateless flow).
func (r *SQLiteUserRepository) CreateUserWithID(ctx context.Context, id, username string) (*User, error) {
    createdAt := time.Now().UTC()
    _, err := r.pool.ExecContext(ctx,
        "INSERT INTO users (id, username, created_at) VALUES (?, ?, ?)",
        id, username, createdAt,
    )
    if err != nil {
        return nil, err
    }
    return &User{ID: id, Username: username, CreatedAt: createdAt}, nil
}
