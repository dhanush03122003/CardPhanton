-- Database migrations for WebAuthn server
-- Updated for the Stateless Go server architecture

-- Users table (current_challenge removed, challenge state is now in JWT)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT timezone('Asia/Kolkata', now())
);

-- Authenticators table
CREATE TABLE IF NOT EXISTS authenticators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    credential_id TEXT UNIQUE NOT NULL,
    public_key BYTEA NOT NULL,
    counter INTEGER NOT NULL DEFAULT 0,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    transports TEXT[],
    aaguid TEXT,
    device_type TEXT,
    backed_up BOOLEAN,
    attachment_type TEXT,
    nickname VARCHAR(255),
    last_used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT timezone('Asia/Kolkata', now())
);

-- Audit logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    credential_id TEXT NOT NULL,
    ip_address TEXT,
    location TEXT,
    user_agent TEXT,
    action_type VARCHAR(50),
    login_time TIMESTAMP DEFAULT timezone('Asia/Kolkata', now())
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_authenticators_user_id ON authenticators(user_id);
CREATE INDEX IF NOT EXISTS idx_authenticators_credential_id ON authenticators(credential_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_credential_id ON audit_logs(credential_id);