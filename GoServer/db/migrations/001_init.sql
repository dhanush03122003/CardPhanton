-- Database migrations for WebAuthn server (SQLite)

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'USER',
    status VARCHAR(20) NOT NULL DEFAULT 'ENABLED',
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
    target_user_id TEXT,
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
    payment_method_type TEXT NOT NULL,
    card_brand TEXT NOT NULL,
    product_name TEXT NOT NULL DEFAULT '',
    linked_phone_number TEXT NOT NULL DEFAULT '9000000000',
    exp_month INTEGER NOT NULL,
    exp_year INTEGER NOT NULL,
    cvv INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_authenticators_user_id ON authenticators(user_id);
CREATE INDEX IF NOT EXISTS idx_authenticators_credential_id ON authenticators(credential_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_credential_id ON audit_logs(credential_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_logs_target_user_id ON admin_audit_logs(target_user_id);
CREATE INDEX IF NOT EXISTS idx_cards_user_id ON cards(user_id);