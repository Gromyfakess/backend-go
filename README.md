# Siro Backend - Simple Office Work Order System

A simple, secure Go backend API for managing work orders in a small office (~50 users).

## Features

- ✅ User authentication with JWT tokens
- ✅ Work order/ticket management
- ✅ Activity logging
- ✅ File uploads (avatars, work order evidence)
- ✅ Role-based access control (Admin/Staff)
- ✅ Simple and beginner-friendly code

## Quick Start

### 1. Install Dependencies

```bash
go mod download
```

### 2. Setup Environment

Copy `.env.example` to `.env` and fill in your settings:

```env
# Server Settings
PORT=8080
SERVER_HOST=localhost          # Use 'localhost' for security, '0.0.0.0' for network access

# Database Settings
DB_HOST=localhost
DB_PORT=3306
DB_USER=your_username
DB_PASSWORD=your_password
DB_NAME=workorder_db

# Security
JWT_SECRET=your_long_random_secret_key_here_at_least_32_characters

# Frontend URL (for CORS)
FRONTEND_URL=http://localhost:3000
```

### 3. Create Database

Run the migration files in `migrations/` folder to create tables.

### 4. Run Server

```bash
go run ./cmd/server/main.go
```

Server will start on `http://localhost:8080`

## Project Structure

```
├── cmd/server/          # Main application entry point
├── internal/
│   ├── controller/     # HTTP request handlers
│   ├── initialize/     # App initialization
│   ├── middlewares/    # Authentication middleware
│   ├── models/         # Data structures
│   ├── repo/          # Database queries
│   └── routers/       # Route definitions
├── pkg/
│   ├── logger/        # Logging utilities
│   ├── response/       # Response helpers
│   ├── setting/       # Database connection
│   └── utils/         # Utility functions (token, file)
├── global/            # Constants
├── migrations/        # Database migration SQL files
└── uploads/           # Uploaded files storage
```

## Security

### Default: Localhost Only

By default, the server only accepts connections from `localhost` (this computer).

**To change to network access:**
1. Edit `.env` file
2. Change `SERVER_HOST=0.0.0.0`
3. Configure Windows Firewall (see `SECURITY_GUIDE.md`)

**To restrict back to localhost:**
1. Edit `.env` file  
2. Change `SERVER_HOST=localhost`
3. Restart server

See `SECURITY_GUIDE.md` for more security information.

## Configuration

### .env vs config.yaml

- **`.env`**: We use this! Simple key=value format, easy to understand
- **`config.yaml`**: Not used - removed to keep things simple

See `CONFIG_EXPLANATION.md` for details.

## API Endpoints

### Authentication
- `POST /login` - Login user
- `POST /refresh` - Refresh access token
- `POST /logout` - Logout user

### User (requires authentication)
- `GET /me` - Get current user info
- `PUT /me` - Update current user
- `POST /upload` - Upload avatar
- `GET /staff` - Get staff list
- `PATCH /staff/:id/availability` - Update availability

### Work Orders (requires authentication)
- `GET /workorders/stats` - Get dashboard stats
- `GET /workorders` - List work orders (with filters)
- `POST /workorders` - Create work order
- `PATCH /workorders/:id/take` - Take/claim work order
- `PATCH /workorders/:id/assign` - Assign to staff (admin)
- `PATCH /workorders/:id/finalize` - Complete work order
- `POST /upload/workorder` - Upload work order evidence

### Activities
- `GET /activities` - Get activity logs

### Admin Only
- `GET /admin/users` - List all users
- `POST /admin/users` - Create user
- `PUT /admin/users/:id` - Update user
- `DELETE /admin/users/:id` - Delete user

## Code Style

- ✅ Simple, beginner-friendly code
- ✅ All errors properly handled
- ✅ Clear comments explaining what code does
- ✅ No complex patterns - easy to understand
- ✅ Security best practices followed

## Documentation

- `CONFIG_EXPLANATION.md` - Explains .env vs config.yaml
- `SECURITY_GUIDE.md` - Security settings and best practices
- `QUICK_START_NETWORK.md` - Network access setup
- `migrations/README.md` - Database migrations explained

## License

See LICENSE file for details.
