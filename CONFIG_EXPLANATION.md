# Simple Explanation: .env vs config.yaml

## What's the Difference?

### .env File (We Use This) ✅
- **What it is**: A simple text file with key=value pairs
- **Where**: In your project root (hidden file)
- **Used by**: Your Go application directly
- **Example**:
  ```
  DB_HOST=localhost
  DB_USER=root
  PORT=8080
  ```

**Why we use it:**
- Simple and easy to understand
- Can hide secrets (don't commit to Git)
- Works directly with Go's `os.Getenv()`

### config.yaml (We DON'T Need This) ❌
- **What it is**: A YAML configuration file (more complex format)
- **Problem**: We're not using it! It's just sitting there doing nothing
- **Why remove**: Redundant - we already use .env

## What We'll Do

**Remove config.yaml** - We don't need it because:
1. We use .env for all configuration
2. It's simpler for beginners
3. Less files = easier to understand

## Your .env File Structure

```env
# Server Settings
PORT=8080
SERVER_HOST=localhost          # Change to 0.0.0.0 for network access

# Database Settings  
DB_HOST=localhost
DB_PORT=3306
DB_USER=your_username
DB_PASSWORD=your_password
DB_NAME=workorder_db

# Security
JWT_SECRET=your_secret_key_here

# Frontend URL (for CORS)
FRONTEND_URL=http://localhost:3000
```

## How to Change Settings

### For Localhost Only (Most Secure)
```env
SERVER_HOST=localhost
DB_HOST=localhost
```

### For Network Access (Other PCs can connect)
```env
SERVER_HOST=0.0.0.0
DB_HOST=172.25.0.70  # Your MySQL server IP
```

That's it! Simple and easy! 🎉
