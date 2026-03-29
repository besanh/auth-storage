CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  email TEXT UNIQUE,
  password_hash TEXT,
  status TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ
);
