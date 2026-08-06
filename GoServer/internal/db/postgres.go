package db

import (
    "context"
    "database/sql"
    "fmt"
    "strings"

    "github.com/joho/godotenv"
    _ "modernc.org/sqlite"
)

// DB holds the database connection pool
type DB struct {
    pool *sql.DB
}

// New creates a new database connection
func New(connectionString string) (*DB, error) {
    _ = godotenv.Load()

    if connectionString == "" || strings.HasPrefix(connectionString, "postgres://") || strings.HasPrefix(connectionString, "postgresql://") {
        connectionString = "webauthn.db"
    }

    pool, err := sql.Open("sqlite", connectionString)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }
    pool.SetMaxOpenConns(1)

    if err := pool.PingContext(context.Background()); err != nil {
        _ = pool.Close()
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    return &DB{pool: pool}, nil
}

// Close closes the database connection
func (db *DB) Close() {
    if db.pool != nil {
        db.pool.Close()
    }
}

// Pool returns the underlying pgx pool
func (db *DB) Pool() *sql.DB {
    return db.pool
}

// InitializeDatabase runs the migrations
func (db *DB) InitializeDatabase(ctx context.Context) error {
    if _, err := db.pool.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
        return fmt.Errorf("failed to enable foreign keys: %w", err)
    }

    statements := []string{
        `CREATE TABLE IF NOT EXISTS users (
            id TEXT PRIMARY KEY,
            username TEXT UNIQUE NOT NULL,
            created_at TIMESTAMP NOT NULL
        )`,
        `CREATE TABLE IF NOT EXISTS authenticators (
            id TEXT PRIMARY KEY,
            credential_id TEXT UNIQUE NOT NULL,
            public_key BLOB NOT NULL,
            counter INTEGER NOT NULL DEFAULT 0,
            user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
            transports TEXT,
            aaguid TEXT,
            device_type TEXT,
            backed_up INTEGER,
            attachment_type TEXT,
            nickname TEXT,
            last_used_at TIMESTAMP,
            created_at TIMESTAMP NOT NULL
        )`,
        `CREATE TABLE IF NOT EXISTS audit_logs (
            id TEXT PRIMARY KEY,
            user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
            credential_id TEXT NOT NULL,
            ip_address TEXT,
            location TEXT,
            user_agent TEXT,
            action_type TEXT,
            login_time TIMESTAMP NOT NULL
        )`,
        `CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
        `CREATE INDEX IF NOT EXISTS idx_authenticators_user_id ON authenticators(user_id)`,
        `CREATE INDEX IF NOT EXISTS idx_authenticators_credential_id ON authenticators(credential_id)`,
        `CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id)`,
        `CREATE INDEX IF NOT EXISTS idx_audit_logs_credential_id ON audit_logs(credential_id)`,
    }

    for _, statement := range statements {
        if _, err := db.pool.ExecContext(ctx, statement); err != nil {
            return fmt.Errorf("failed to initialize database: %w", err)
        }
    }

    fmt.Println("Database tables initialized successfully")
    return nil
}
