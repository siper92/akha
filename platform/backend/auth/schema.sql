-- Workers registered with a static access token
CREATE TABLE IF NOT EXISTS workers (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  token_hash TEXT NOT NULL UNIQUE,
  tier TEXT NOT NULL DEFAULT 'worker',
  is_active BOOLEAN NOT NULL DEFAULT 1,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- JWT tokens issued to workers
CREATE TABLE IF NOT EXISTS issued_tokens (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  worker_id INTEGER REFERENCES workers(id),
  token TEXT NOT NULL UNIQUE,
  is_active BOOLEAN NOT NULL DEFAULT 1,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Every login attempt, successful or not
CREATE TABLE IF NOT EXISTS login_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  worker_id INTEGER REFERENCES workers(id),
  addr TEXT NOT NULL,
  ok BOOLEAN NOT NULL,
  reason TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL
);

-- Indices for performance
CREATE INDEX IF NOT EXISTS idx_workers_token_hash ON workers(token_hash);
CREATE INDEX IF NOT EXISTS idx_issued_tokens_token ON issued_tokens(token);
CREATE INDEX IF NOT EXISTS idx_issued_tokens_expires_at ON issued_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_login_log_worker_id ON login_log(worker_id);
