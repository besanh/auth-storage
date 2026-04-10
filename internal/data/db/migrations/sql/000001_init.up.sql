CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  email TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  name TEXT NOT NULL,
  avatar_url TEXT,
  role TEXT NOT NULL,
  scope TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS machine_clients (
  -- The public identifier (e.g., 'media-service-prod')
    client_id VARCHAR(255) PRIMARY KEY,
    
    -- The bcrypt hash of the secret. NEVER store the raw secret!
    client_secret_hash VARCHAR(255) NOT NULL,
    
    -- A human-readable name for your own dashboard/logs
    name VARCHAR(255) NOT NULL,
    
    -- Postgres Native Array: What is this service allowed to do?
    -- e.g., '{"read:users", "write:notifications"}'
    scopes TEXT[] NOT NULL DEFAULT '{}',
    
    -- Standard audit timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);