-- init_postgres.sql
-- Schema for a basic auth users table (PostgreSQL)

-- NOTE: CREATE DATABASE cannot be done inside this file reliably.
-- Create the database separately (see commands below), then run this file
-- against that database.

-- Create users table if missing
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash CHAR(64) NOT NULL,    -- if you're using SHA-256 hex
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Example safe seed: will not create duplicates if username already exists
INSERT INTO users (username, email, password_hash)
VALUES (
    'testuser',
    'test@example.com',
    '34819d7beeabb9260a5c854bc85b3e44f7928a9a5e4b9e5d0d5e4b7f6895f3c3'
)
ON CONFLICT (username) DO NOTHING;
