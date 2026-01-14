# Security Guide for Office Backend

## Quick Security Settings

### For Localhost Only (Most Secure) ✅

Edit your `.env` file:
```env
SERVER_HOST=localhost
DB_HOST=localhost
```

**What this does:**
- Only this computer can access the backend
- Most secure for office use
- No network exposure

### For Network Access (Other PCs)

Edit your `.env` file:
```env
SERVER_HOST=0.0.0.0
DB_HOST=172.25.0.70  # Your MySQL server IP
```

**What this does:**
- Other computers on your network can access the backend
- Less secure - only use if needed

## Security Features Already Implemented

✅ **Password Hashing**: All passwords are hashed using bcrypt
✅ **JWT Tokens**: Secure token-based authentication
✅ **CORS Protection**: Only allows requests from your frontend
✅ **Input Validation**: All inputs are validated
✅ **Error Handling**: No sensitive errors exposed to users
✅ **SQL Injection Protection**: Using parameterized queries

## Best Practices for Office Use

1. **Use Strong JWT Secret**
   - At least 32 characters long
   - Random string of letters, numbers, symbols
   - Don't share it with others

2. **Database Security**
   - Use strong database passwords
   - Don't use default MySQL root user
   - Limit database access to your app only

3. **Regular Updates**
   - Keep Go dependencies updated
   - Update MySQL regularly

4. **Backup**
   - Regular database backups
   - Keep backups secure

5. **Monitor Logs**
   - Check server logs regularly
   - Watch for suspicious activity

## Changing Network Access

### To Restrict to Localhost Only:

1. Open `.env` file
2. Change:
   ```
   SERVER_HOST=localhost
   ```
3. Restart server

### To Allow Network Access:

1. Open `.env` file
2. Change:
   ```
   SERVER_HOST=0.0.0.0
   ```
3. Configure Windows Firewall (see QUICK_START_NETWORK.md)
4. Restart server

## Important Notes

⚠️ **Never commit `.env` file to Git** - it contains secrets!
⚠️ **Use HTTPS in production** - HTTP is not secure for production
⚠️ **Limit admin users** - Only trusted people should be admins
⚠️ **Regular password changes** - Encourage users to change passwords
