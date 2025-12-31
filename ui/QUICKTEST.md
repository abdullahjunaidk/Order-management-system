# 🚀 Quick Test - Login UI

## ✅ Your Docker is Running!

All services are up and running. Gateway is responding on port 8080.

---

## 📝 Step 1: Create a Test User

Copy and paste this command in PowerShell:

```powershell
$body = @{
    username = "testuser"
    name = "Test User"
    email = "test@example.com"
    phone = 1234567890
    password = "password123"
    companyIds = @()
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/user/register" -Method POST -Body $body -ContentType "application/json"
```

### Expected Response:
```json
{
  "message": "Employee Registered Successfully!",
  "user": {
    "id": "...",
    "name": "Test User",
    "username": "testuser",
    "email": "test@example.com",
    ...
  }
}
```

---

## 🔐 Step 2: Login via UI

1. **Open the login page** (already open in your browser)
   - Or navigate to: `d:\Laravel\Order-management-system\ui\index.html`

2. **Enter credentials:**
   - **Email/Username:** `test@example.com` (or `testuser`)
   - **Password:** `password123`

3. **Click "Sign In"**

4. **You should see:**
   - ✅ Green success message
   - ✅ Redirect to dashboard after 1.5 seconds
   - ✅ Welcome page with your email

---

## 🧪 Alternative: Test Login via PowerShell

If you want to test the API directly first:

```powershell
$body = @{
    identifier = "test@example.com"
    password = "password123"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/user/login" -Method POST -Body $body -ContentType "application/json"
```

### Expected Response:
```json
{
  "message": "User Logged In Successfully!",
  "user": { ... },
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

---

## 📋 Registration Requirements

Your API requires these fields:

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `username` | string | ✅ | 3-20 characters |
| `name` | string | ✅ | Any |
| `email` | string | ✅ | Valid email format |
| `phone` | int64 | ✅ | 10 digits |
| `password` | string | ✅ | Min 8 characters |
| `companyIds` | array | ❌ | Optional |

---

## 🔑 Login Requirements

Your API accepts:

| Field | Type | Description |
|-------|------|-------------|
| `identifier` | string | Email OR Username |
| `password` | string | User password |

**Note:** The UI currently shows "Email Address" but it accepts both email and username!

---

## ✨ What Happens After Login

1. **Tokens Stored:**
   - `accessToken` → localStorage
   - `refreshToken` → localStorage
   - `userType` → localStorage
   - `userEmail` → localStorage
   - `userData` → localStorage

2. **Redirect:**
   - User login → `/user/dashboard.html`
   - Admin login → `/admin/dashboard.html`

3. **Dashboard Shows:**
   - Welcome message
   - Your email
   - User type
   - Logout button

---

## 🎯 Quick Commands Reference

### Check Gateway Status:
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/ping" -Method GET
```

### View Docker Containers:
```powershell
docker ps
```

### Check Gateway Logs:
```powershell
docker logs gateway-service
```

### Check Auth Service Logs:
```powershell
docker logs auth-service
```

---

## 🐛 If Something Goes Wrong

### User Registration Fails:
- Check if phone is exactly 10 digits (as integer)
- Password must be at least 8 characters
- Username must be 3-20 characters
- Email must be valid format

### Login Fails:
- Make sure you registered the user first
- Check if you're using the correct password
- Try using username instead of email (or vice versa)
- Check browser console (F12) for errors

### Can't Connect:
```powershell
# Restart gateway service
docker restart gateway-service

# Check if it's running
docker ps | findstr gateway
```

---

## 🎉 Ready to Test!

Your login UI is ready and your Docker services are running!

**Just run the registration command above, then login via the UI!** 🚀

---

**Current Status:**
- ✅ Docker services running
- ✅ Gateway responding (port 8080)
- ✅ Login UI created and ready
- ✅ API integration complete
- ⏳ Waiting for test user creation

**Next:** Create a test user and login! 🎯
