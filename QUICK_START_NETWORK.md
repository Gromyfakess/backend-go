# Quick Start: Make Backend Accessible from Other PCs

## ✅ What's Been Done

1. **Server Configuration Updated**
   - Server now binds to `0.0.0.0:8080` (accepts connections from all network interfaces)
   - Updated `cmd/server/main.go` to listen on all interfaces

2. **Environment Variables Updated**
   - `DB_HOST=172.25.0.70` (your MySQL server IP)
   - `SERVER_HOST=0.0.0.0` (listen on all interfaces)
   - `BACKEND_URL=http://172.25.0.70:8080`

3. **Windows Firewall**
   - Firewall rule created for port 8080
   - ⚠️ **Note**: If verification shows firewall rule missing, run as Administrator:
     ```powershell
     New-NetFirewallRule -DisplayName "Siro Backend API" -Direction Inbound -LocalPort 8080 -Protocol TCP -Action Allow
     ```

## ⚠️ Important: IP Address Mismatch

Your actual IP is: **172.25.10.169**  
But configuration uses: **172.25.0.70**

### Option 1: Use Your Actual IP (Recommended)
Update `.env` file:
```env
DB_HOST=172.25.10.169
BACKEND_URL=http://172.25.10.169:8080
FRONTEND_URL=http://172.25.10.169:3000
```

### Option 2: Keep 172.25.0.70
If `172.25.0.70` is a different server (like a Docker container or remote MySQL), keep it as is.

## 🔧 Next Steps

### Step 1: Configure MySQL for Remote Access

**If MySQL is on this PC (172.25.10.169):**

1. Edit MySQL config file (`my.ini` or `my.cnf`):
   ```ini
   [mysqld]
   bind-address = 0.0.0.0
   ```

2. Grant remote access:
   ```sql
   mysql -u root -p
   ```
   ```sql
   GRANT ALL PRIVILEGES ON workorder_db.* TO 'david'@'172.25.%' IDENTIFIED BY 'david123';
   FLUSH PRIVILEGES;
   ```

3. Restart MySQL:
   ```powershell
   Restart-Service mysql
   ```

**Or use the provided script:**
```powershell
mysql -u root -p < scripts/setup_mysql_remote_access.sql
```

### Step 2: Start Your Server

```powershell
go run ./cmd/server/main.go
```

You should see:
```
Server running on 0.0.0.0:8080 (Allowed Origin: http://172.25.0.70:3000)
Accessible from other PCs at: http://172.25.0.70:8080
Database Connected Successfully via database/sql!
```

### Step 3: Test from Another PC

From another computer on the same network:

1. **Test connectivity:**
   ```powershell
   Test-NetConnection -ComputerName 172.25.0.70 -Port 8080
   # Or use your actual IP:
   Test-NetConnection -ComputerName 172.25.10.169 -Port 8080
   ```

2. **Open in browser:**
   ```
   http://172.25.0.70:8080
   # Or:
   http://172.25.10.169:8080
   ```

3. **Test API endpoint:**
   ```powershell
   curl http://172.25.0.70:8080/me
   ```

## 🔍 Verify Setup

Run the verification script:
```powershell
powershell -ExecutionPolicy Bypass -File scripts/verify_setup.ps1
```

## 📝 Frontend Configuration

Update your frontend to use:
```javascript
const API_URL = "http://172.25.0.70:8080";
// Or use your actual IP:
const API_URL = "http://172.25.10.169:8080";
```

## 🛡️ Security Notes

- ✅ Firewall allows port 8080
- ⚠️ MySQL should only allow connections from your network (172.25.%)
- ⚠️ Use strong passwords in production
- ⚠️ Consider HTTPS for production

## 🐛 Troubleshooting

### Can't connect from other PC?

1. **Check firewall:**
   ```powershell
   Get-NetFirewallRule -DisplayName "Siro Backend API"
   ```

2. **Verify server is running:**
   ```powershell
   netstat -an | findstr 8080
   ```

3. **Test MySQL connection:**
   ```powershell
   Test-NetConnection -ComputerName 172.25.0.70 -Port 3306
   ```

4. **Check server logs** for connection errors

### Database connection fails?

1. Verify MySQL allows remote connections
2. Check user privileges: `SHOW GRANTS FOR 'david'@'172.25.%';`
3. Ensure firewall allows port 3306
4. Verify DB_HOST in `.env` matches MySQL server IP

## 📚 More Information

See `NETWORK_SETUP.md` for detailed instructions.
