-- Access tokens table
CREATE TABLE IF NOT EXISTS access_tokens (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  token TEXT NOT NULL UNIQUE,
  is_active BOOLEAN NOT NULL DEFAULT 1,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indices for performance
CREATE INDEX IF NOT EXISTS idx_access_tokens_token ON access_tokens(token);
CREATE INDEX IF NOT EXISTS idx_access_tokens_expires_at ON access_tokens(expires_at);