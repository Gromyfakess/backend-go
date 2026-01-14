-- Migration: Create User Tokens Table
-- Description: Stores JWT access and refresh tokens for user sessions
-- Date: Initial migration

CREATE TABLE IF NOT EXISTS user_tokens (
    user_id INT UNSIGNED PRIMARY KEY,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    at_expires_at TIMESTAMP NOT NULL,
    rt_expires_at TIMESTAMP NOT NULL,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_expires_at (rt_expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
