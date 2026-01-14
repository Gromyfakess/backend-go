-- EXAMPLE MIGRATION: How to Add a New Feature
-- This is just an example - don't run this in production!
-- It shows how you would add a new feature (notification preferences)

-- Step 1: Add a new column to existing table
-- Migration: Add notification preferences to users table
-- Description: Allows users to enable/disable notifications

ALTER TABLE users 
ADD COLUMN notification_enabled BOOLEAN DEFAULT TRUE AFTER availability,
ADD COLUMN notification_email BOOLEAN DEFAULT TRUE AFTER notification_enabled;

-- Step 2: Create a new table for notification settings (if needed)
-- CREATE TABLE IF NOT EXISTS notification_settings (
--     user_id INT UNSIGNED PRIMARY KEY,
--     email_notifications BOOLEAN DEFAULT TRUE,
--     push_notifications BOOLEAN DEFAULT TRUE,
--     FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
-- );

-- Step 3: Add indexes if needed for performance
-- CREATE INDEX idx_notification_enabled ON users(notification_enabled);

-- ROLLBACK (if you need to undo this migration):
-- ALTER TABLE users 
-- DROP COLUMN notification_enabled,
-- DROP COLUMN notification_email;
