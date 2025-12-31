# 🧪 Testing Guide - OMS Login UI

## ✅ Quick Start Testing

### Step 1: Start Your Docker Services
```powershell
cd d:\Laravel\Order-management-system
docker-compose up -d
```

Wait for all services to start (this may take 1-2 minutes).

### Step 2: Verify Gateway is Running
```powershell
curl http://localhost:8080/api/v1/ping
```

Expected response: `{"message":"pong"}`

### Step 3: Create a Test User

#### Option A: Using PowerShell (curl)
```powershell
$body = @{
    email = "test@example.com"
    password = "password123"
    name = "Test User"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/user/register" -Method POST -Body $body -ContentType "application/json"
```

#### Option B: Using Postman
1. Open Postman
2. Create a new POST request
3. URL: `http://localhost:8080/api/v1/auth/user/register`
4. Headers: `Content-Type: application/json`
5. Body (raw JSON):
```json
{
  "email": "test@example.com",
  "password": "password123",
  "name": "Test User"
}
```
6. Click Send

### Step 4: Open the Login Page

#### Method 1: Direct File (Easiest)
1. Navigate to: `d:\Laravel\Order-management-system\ui`
2. Double-click `index.html`

#### Method 2: Using a Local Server
```powershell
cd d:\Laravel\Order-management-system\ui
python -m http.server 3000
```
Then open: http://localhost:3000

### Step 5: Test Login
1. Make sure "User Login" is selected (blue button)
2. Enter email: `test@example.com`
3. Enter password: `password123`
4. Click "Sign In"
5. You should see a success message and be redirected to the dashboard!

---

## 🔍 Testing Checklist

### UI Features to Test:

- [ ] **User/Admin Toggle**
  - Click "Admin Login" button
  - Form title should change to "Admin Portal"
  - Click "User Login" button
  - Form title should change back to "Welcome Back"

- [ ] **Form Validation**
  - Try submitting empty form → Should show validation errors
  - Enter invalid email (e.g., "notanemail") → Should show error
  - Enter password less than 6 characters → Should show error
  - Enter valid credentials → Should proceed to login

- [ ] **Password Visibility**
  - Click the eye icon next to password field
  - Password should become visible
  - Click again → Password should be hidden

- [ ] **Remember Me**
  - Check the "Remember me" checkbox
  - (This is UI only for now, functionality can be added later)

- [ ] **Responsive Design**
  - Resize browser window
  - UI should adapt to different screen sizes
  - Try on mobile device if available

### API Integration Tests:

- [ ] **User Login**
  - Use test credentials created earlier
  - Should receive access token
  - Should be redirected to user dashboard
  - Check browser localStorage for tokens

- [ ] **Admin Login**
  - Switch to "Admin Login"
  - Try logging in (you'll need to create an admin first)
  - Should redirect to admin dashboard

- [ ] **Error Handling**
  - Try wrong password → Should show error message
  - Try non-existent email → Should show error message
  - Turn off Docker services → Should show connection error

- [ ] **Session Management**
  - After successful login, refresh the page
  - Should auto-redirect to dashboard (already logged in)
  - Click logout on dashboard
  - Should return to login page

---

## 🐛 Common Issues & Solutions

### Issue: "Failed to connect to server"
**Solution:**
1. Check if Docker is running: `docker ps`
2. Check gateway logs: `docker logs gateway-service`
3. Verify gateway is on port 8080: `netstat -ano | findstr :8080`

### Issue: "Login failed - Invalid credentials"
**Solution:**
1. Make sure you created a user first (Step 3)
2. Check the email and password are correct
3. Verify user exists in database

### Issue: CORS Error in Browser Console
**Solution:**
Your gateway already has CORS middleware, but if you see CORS errors:
1. Make sure you're accessing via `http://localhost` not `file://`
2. Use a local server (Python/Node) instead of opening file directly

### Issue: Page doesn't load properly
**Solution:**
1. Clear browser cache (Ctrl + Shift + Delete)
2. Hard refresh (Ctrl + F5)
3. Check browser console for JavaScript errors (F12)

### Issue: Redirect to dashboard fails
**Solution:**
The dashboard pages are placeholders. For now, you'll see a simple welcome page. This is expected!

---

## 📊 What to Check in Browser DevTools

### Console Tab (F12)
You should see:
```
🚀 Order Management System
Welcome to the OMS Login Portal
API Endpoint: http://localhost:8080/api/v1
```

### Network Tab
When you submit login:
1. Should see POST request to `/api/v1/auth/user/login`
2. Status should be 200 (success) or 401 (wrong credentials)
3. Response should contain `accessToken` and `refreshToken`

### Application Tab → Local Storage
After successful login, you should see:
- `accessToken`: Your JWT token
- `refreshToken`: Your refresh token
- `userType`: "user" or "admin"
- `userEmail`: Your email address
- `userData`: JSON object with user info

---

## 🎯 Expected Behavior

### Successful Login Flow:
1. Enter valid credentials
2. Click "Sign In"
3. Button shows loading spinner
4. Green success alert appears: "Login successful! Redirecting..."
5. After 1.5 seconds, redirects to dashboard
6. Dashboard shows welcome message with your email

### Failed Login Flow:
1. Enter invalid credentials
2. Click "Sign In"
3. Button shows loading spinner
4. Red error alert appears with error message
5. Form shakes slightly (error animation)
6. Can try again

---

## 🔐 Creating Admin User

To test admin login, you need to create an admin user. This typically requires:

1. **Check your auth service** for admin creation endpoint
2. **Or use database directly** to set user role to admin
3. **Or check if there's a seed/migration** that creates default admin

Example (if endpoint exists):
```powershell
$body = @{
    email = "admin@example.com"
    password = "admin123"
    name = "Admin User"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/admin/register" -Method POST -Body $body -ContentType "application/json"
```

---

## 📸 Screenshots

The login page should look like:
- **Left side:** Branding with logo, title, and feature cards
- **Right side:** Login form with user/admin toggle
- **Background:** Animated gradient orbs (purple/pink)
- **Theme:** Dark with glassmorphism effects

---

## 🚀 Next Steps After Testing

Once login works:
1. ✅ Build out the dashboard pages
2. ✅ Add order management interface
3. ✅ Add product management
4. ✅ Add inventory tracking
5. ✅ Implement token refresh logic
6. ✅ Add user profile management

---

## 📞 Need Help?

If something isn't working:
1. Check browser console (F12) for errors
2. Check Docker logs: `docker logs gateway-service`
3. Verify all services are running: `docker ps`
4. Check the README.md for more details

**Happy Testing! 🎉**
