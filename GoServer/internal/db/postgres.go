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

		// FIX: Only skip if the chunk is completely empty.
		// Let SQLite handle the inline comments natively.
		if statement == "" {
			continue
		}

		if _, err := db.pool.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	// Older databases may have been created before users.role was introduced.
	rows, err := db.pool.QueryContext(ctx, "PRAGMA table_info(users)")
	if err != nil {
		return fmt.Errorf("failed to inspect users table: %w", err)
	}
	roleColumnExists := false
	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull, primaryKey int
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			rows.Close()
			return fmt.Errorf("failed to inspect users columns: %w", err)
		}
		if name == "role" {
			roleColumnExists = true
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("failed to read users columns: %w", err)
	}
	rows.Close()

	if !roleColumnExists {
		if _, err := db.pool.ExecContext(ctx, "ALTER TABLE users ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'USER'"); err != nil {
			return fmt.Errorf("failed to add users role column: %w", err)
		}
	}

	rows, err = db.pool.QueryContext(ctx, "PRAGMA table_info(users)")
	if err != nil {
		return fmt.Errorf("failed to inspect users status column: %w", err)
	}
	statusColumnExists := false
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			rows.Close()
			return fmt.Errorf("failed to inspect users status column: %w", err)
		}
		if name == "status" {
			statusColumnExists = true
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("failed to read users status column: %w", err)
	}
	rows.Close()

	if !statusColumnExists {
		if _, err := db.pool.ExecContext(ctx, "ALTER TABLE users ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'ENABLED'"); err != nil {
			return fmt.Errorf("failed to add users status column: %w", err)
		}
	}

	rows, err = db.pool.QueryContext(ctx, "PRAGMA table_info(cards)")
	if err != nil {
		return fmt.Errorf("failed to inspect cards product column: %w", err)
	}
	productNameColumnExists := false
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			rows.Close()
			return fmt.Errorf("failed to inspect cards product column: %w", err)
		}
		if name == "product_name" {
			productNameColumnExists = true
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("failed to read cards product column: %w", err)
	}
	rows.Close()

	if !productNameColumnExists {
		if _, err := db.pool.ExecContext(ctx, "ALTER TABLE cards ADD COLUMN product_name TEXT NOT NULL DEFAULT ''"); err != nil {
			return fmt.Errorf("failed to add cards product column: %w", err)
		}
	}

	rows, err = db.pool.QueryContext(ctx, "PRAGMA table_info(cards)")
	if err != nil {
		return fmt.Errorf("failed to inspect cards phone column: %w", err)
	}
	phoneColumnExists := false
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			rows.Close()
			return fmt.Errorf("failed to inspect cards phone column: %w", err)
		}
		if name == "linked_phone_number" {
			phoneColumnExists = true
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("failed to read cards phone column: %w", err)
	}
	rows.Close()
	if !phoneColumnExists {
		if _, err := db.pool.ExecContext(ctx, "ALTER TABLE cards ADD COLUMN linked_phone_number TEXT NOT NULL DEFAULT '9000000000'"); err != nil {
			return fmt.Errorf("failed to add cards phone column: %w", err)
		}
	}

	fmt.Println("Database tables initialized successfully")
	return nil
}
