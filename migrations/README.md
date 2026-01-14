# Database Migrations Explained

## What are Database Migrations?

**Database migrations** are scripts (SQL files) that help you manage changes to your database structure over time. Think of them as a version control system for your database schema.

### Simple Analogy
Imagine you're building a house:
- **Migration 1**: Build the foundation (create tables)
- **Migration 2**: Add walls (add new columns)
- **Migration 3**: Add a second floor (add new tables)
- **Migration 4**: Renovate a room (modify existing columns)

Each migration is a step-by-step instruction that changes your database from one state to another.

## Why Use Migrations?

### 1. **Version Control for Database**
Just like Git tracks code changes, migrations track database changes. You can see the history of all database modifications.

### 2. **Team Collaboration**
When a teammate adds a new feature that requires a new database column, they create a migration file. You can run it to get the same database structure.

### 3. **Consistent Environments**
- **Development**: Your local database
- **Staging**: Testing database
- **Production**: Live database

Migrations ensure all environments have the same database structure.

### 4. **Rollback Capability**
If something goes wrong, you can "undo" migrations to revert database changes.

## How Migrations Work

### Basic Flow:
```
1. Developer creates a migration file (e.g., `001_create_users_table.sql`)
2. Migration file contains SQL commands
3. Migration tool runs the SQL against the database
4. Tool tracks which migrations have been applied
5. Next time, only new migrations are run
```

## Example Migration Files

Based on your project structure, here are example migrations:

### Migration 1: Create Users Table
```sql
-- File: migrations/001_create_users_table.sql
CREATE TABLE IF NOT EXISTS users (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'Staff',
    unit VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    avatar_url VARCHAR(500),
    availability VARCHAR(50) DEFAULT 'Offline',
    can_crud BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### Migration 2: Create User Tokens Table
```sql
-- File: migrations/002_create_user_tokens_table.sql
CREATE TABLE IF NOT EXISTS user_tokens (
    user_id INT UNSIGNED PRIMARY KEY,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    at_expires_at TIMESTAMP NOT NULL,
    rt_expires_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### Migration 3: Create Work Orders Table
```sql
-- File: migrations/003_create_work_orders_table.sql
CREATE TABLE IF NOT EXISTS work_orders (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    priority VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'Pending',
    unit VARCHAR(255) NOT NULL,
    photo_url VARCHAR(500),
    requester_id INT UNSIGNED NOT NULL,
    assignee_id INT UNSIGNED,
    taken_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    completed_by_id INT UNSIGNED,
    completion_note TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (requester_id) REFERENCES users(id),
    FOREIGN KEY (assignee_id) REFERENCES users(id),
    FOREIGN KEY (completed_by_id) REFERENCES users(id)
);
```

### Migration 4: Create Activity Logs Table
```sql
-- File: migrations/004_create_activity_logs_table.sql
CREATE TABLE IF NOT EXISTS activity_logs (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    user_name VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL,
    request_id INT UNSIGNED NOT NULL,
    details TEXT NOT NULL,
    status VARCHAR(50) NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (request_id) REFERENCES work_orders(id)
);
```

### Migration 5: Add Index for Performance
```sql
-- File: migrations/005_add_indexes.sql
-- Add indexes to improve query performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_work_orders_status ON work_orders(status);
CREATE INDEX idx_work_orders_unit ON work_orders(unit);
CREATE INDEX idx_activity_logs_timestamp ON activity_logs(timestamp);
```

### Migration 6: Add New Column (Example: Future Feature)
```sql
-- File: migrations/006_add_notification_preference.sql
-- Example: Adding a new feature - notification preferences
ALTER TABLE users 
ADD COLUMN notification_enabled BOOLEAN DEFAULT TRUE AFTER availability;
```

## Migration Naming Convention

Common naming patterns:
- `001_create_users_table.sql`
- `002_create_user_tokens_table.sql`
- `003_add_email_index.sql`
- `YYYYMMDDHHMMSS_description.sql` (timestamp-based)

## Migration Tools

Popular Go migration tools:
1. **golang-migrate** - Most popular
2. **sql-migrate** - Simple and easy
3. **goose** - Feature-rich

## Running Migrations

### Using golang-migrate (Example):
```bash
# Install tool
go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Create a new migration
migrate create -ext sql -dir migrations -seq create_users_table

# Run migrations up
migrate -path migrations -database "mysql://user:pass@tcp(localhost:3306)/dbname" up

# Rollback last migration
migrate -path migrations -database "mysql://user:pass@tcp(localhost:3306)/dbname" down 1
```

## Best Practices

1. **One Change Per Migration**: Each migration should do one thing
2. **Reversible**: Write both "up" and "down" migrations when possible
3. **Test First**: Test migrations on a copy of production data
4. **Never Edit Old Migrations**: Create new migrations to fix issues
5. **Document Changes**: Add comments explaining why changes were made

## Migration States

Your database keeps track of which migrations have been applied:
- ✅ Applied migrations: Already run
- ⏳ Pending migrations: Not yet run
- ❌ Failed migrations: Need attention

## Example: Adding a Feature

Let's say you want to add "user preferences":

1. **Create Migration File**: `007_add_user_preferences.sql`
```sql
CREATE TABLE user_preferences (
    user_id INT UNSIGNED PRIMARY KEY,
    theme VARCHAR(50) DEFAULT 'light',
    language VARCHAR(10) DEFAULT 'en',
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

2. **Run Migration**: Apply to database
3. **Update Code**: Modify your Go models and repository
4. **Commit**: Add migration file to Git

Now everyone on your team can run this migration to get the new feature!

## Summary

- **Migrations** = Version control for your database
- **Purpose** = Track and apply database changes consistently
- **Benefit** = Team collaboration and consistent environments
- **Location** = `migrations/` folder in your project

Think of migrations as instructions that say: "To get from database version A to version B, run these SQL commands."
