-- Database migrations for WebAuthn server (SQLite)

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'USER',
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING_APPROVAL',
    created_at TIMESTAMP NOT NULL
);

-- Authenticators table
CREATE TABLE IF NOT EXISTS authenticators (
    id TEXT PRIMARY KEY,
    credential_id TEXT UNIQUE NOT NULL,
    public_key BLOB NOT NULL,
    counter INTEGER NOT NULL DEFAULT 0,
    user_id TEXT,
    transports TEXT,
    aaguid TEXT,
    device_type TEXT,
    backed_up INTEGER,
    attachment_type TEXT,
    nickname TEXT,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL
);

-- Audit logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id TEXT PRIMARY KEY,
    user_id TEXT,
    credential_id TEXT NOT NULL,
    ip_address TEXT,
    location TEXT,
    user_agent TEXT,
    action_type TEXT,
    login_time TIMESTAMP NOT NULL
);

-- Administrative actions are retained independently of user records.
CREATE TABLE IF NOT EXISTS admin_audit_logs (
    id TEXT PRIMARY KEY,
    admin_user_id TEXT NOT NULL,
    admin_user_name TEXT NOT NULL DEFAULT '',
    target_user_id TEXT,
    target_user_name TEXT NOT NULL DEFAULT '',
    action_type TEXT NOT NULL,
    details TEXT,
    created_at TIMESTAMP NOT NULL
);

-- Cards table
CREATE TABLE IF NOT EXISTS cards (
    id TEXT PRIMARY KEY,
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    PAN TEXT UNIQUE NOT NULL,
    cardholder_name TEXT NOT NULL,
    bank_name TEXT NOT NULL,
    email TEXT,
    payment_method_type TEXT NOT NULL,
    card_brand TEXT NOT NULL,
    product_name TEXT NOT NULL DEFAULT '',
    linked_phone_number TEXT NOT NULL DEFAULT '9000000000',
    exp_month INTEGER NOT NULL,
    exp_year INTEGER NOT NULL,
    cvv INTEGER NOT NULL,
    total_limit REAL,
    available_limit REAL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Temporal refresh runs for card limit synchronization.
CREATE TABLE IF NOT EXISTS card_sync_runs (
    id TEXT PRIMARY KEY,
    card_id TEXT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    PAN TEXT NOT NULL,
    workflow_id TEXT NOT NULL,
    run_id TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL CHECK (status IN ('RUNNING', 'COMPLETED', 'FAILED')),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Google OAuth credentials linked to a user's card.
CREATE TABLE IF NOT EXISTS google_oauth_tokens (
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_id TEXT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    gmail TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    last_updated TIMESTAMP NOT NULL,
    PRIMARY KEY (user_id, card_id)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_authenticators_user_id ON authenticators(user_id);
CREATE INDEX IF NOT EXISTS idx_authenticators_credential_id ON authenticators(credential_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_credential_id ON audit_logs(credential_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_logs_target_user_id ON admin_audit_logs(target_user_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_logs_target_user_name ON admin_audit_logs(target_user_name);
CREATE INDEX IF NOT EXISTS idx_cards_user_id ON cards(user_id);
CREATE INDEX IF NOT EXISTS idx_card_sync_runs_card_id ON card_sync_runs(card_id);
CREATE INDEX IF NOT EXISTS idx_google_oauth_tokens_card_id ON google_oauth_tokens(card_id);
CREATE INDEX IF NOT EXISTS idx_card_sync_runs_created_at ON card_sync_runs(created_at);