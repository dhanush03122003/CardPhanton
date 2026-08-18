package db

import (
    "context"
    "database/sql"
    "fmt"
    "io/ioutil"
    "path/filepath"
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

// InitializeDatabase runs the migrations from file
func (db *DB) InitializeDatabase(ctx context.Context) error {
    if _, err := db.pool.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
        return fmt.Errorf("failed to enable foreign keys: %w", err)
    }

    // Read migration file
    migrationFile := filepath.Join("db", "migrations", "001_init.sql")
    sqlContent, err := ioutil.ReadFile(migrationFile)
    if err != nil {
        return fmt.Errorf("failed to read migration file: %w", err)
    }

    // Split by semicolon and execute each statement
    statements := strings.Split(string(sqlContent), ";")
    for _, statement := range statements {
        statement = strings.TrimSpace(statement)
        if statement == "" || strings.HasPrefix(statement, "--") {
            continue
        }
        if _, err := db.pool.ExecContext(ctx, statement); err != nil {
            return fmt.Errorf("failed to execute migration: %w", err)
        }
    }

    fmt.Println("Database tables initialized successfully")
    return nil
}
