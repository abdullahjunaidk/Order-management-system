# 🚀 Order Management System - UI

## 📋 Overview
This is the frontend UI for the Order Management System microservices architecture. The system provides a modern, secure login interface for both users and administrators.

## 🎨 Features

### ✨ Modern UI/UX
- **Glassmorphism Design** - Beautiful glass-effect panels with backdrop blur
- **Gradient Animations** - Smooth animated gradient orbs in the background
- **Responsive Layout** - Works perfectly on desktop, tablet, and mobile
- **Dark Theme** - Eye-friendly dark color scheme
- **Micro-animations** - Smooth transitions and hover effects

### 🔐 Authentication
- **Dual Login System** - Separate login for Users and Admins
- **Real-time Validation** - Instant feedback on form inputs
- **Password Visibility Toggle** - Show/hide password feature
- **Session Management** - Automatic token storage and session handling
- **Error Handling** - Clear, user-friendly error messages

### 🛠️ Technical Features
- **API Integration** - Connected to your Go microservices backend
- **Token Management** - Stores access and refresh tokens
- **Form Validation** - Client-side validation before API calls
- **Loading States** - Visual feedback during API requests
- **Alert System** - Success, error, and warning notifications

## 📁 File Structure

```
ui/
├── index.html          # Main login page
├── styles.css          # All styling and animations
├── script.js           # Authentication logic and API calls
└── README.md           # This file
```

## 🚀 How to Use

### Option 1: Direct File Access (Simplest)
1. **Navigate to the UI folder:**
   ```bash
   cd d:\Laravel\Order-management-system\ui
   ```

2. **Open in browser:**
   - Simply double-click `index.html` OR
   - Right-click `index.html` → Open with → Your browser

### Option 2: Using a Local Server (Recommended)

#### Using Python (if installed):
```bash
cd d:\Laravel\Order-management-system\ui
python -m http.server 3000
```
Then open: http://localhost:3000

#### Using Node.js (if installed):
```bash
cd d:\Laravel\Order-management-system\ui
npx http-server -p 3000
```
Then open: http://localhost:3000

#### Using PHP (if installed):
```bash
cd d:\Laravel\Order-management-system\ui
php -S localhost:3000
```
Then open: http://localhost:3000

### Option 3: Using Live Server (VS Code Extension)
1. Install "Live Server" extension in VS Code
2. Right-click on `index.html`
3. Select "Open with Live Server"

## 🔧 Configuration

### API Endpoint
The default API endpoint is set to `http://localhost:8080/api/v1`

To change this, edit `script.js`:
```javascript
const API_BASE_URL = 'http://localhost:8080/api/v1';
```

### Authentication Endpoints
- **User Login:** `POST /api/v1/auth/user/login`
- **Admin Login:** `POST /api/v1/auth/admin/login`

## 📝 API Request Format

### Login Request
```json
{
  "email": "user@example.com",
  "password": "yourpassword"
}
```

### Expected Response (Success)
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIs...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "123",
    "email": "user@example.com",
    "name": "John Doe"
  }
}
```

## 🧪 Testing the Login

### Before Testing:
1. **Make sure your Docker services are running:**
   ```bash
   cd d:\Laravel\Order-management-system
   docker-compose up -d
   ```

2. **Verify the gateway is running:**
   ```bash
   curl http://localhost:8080/api/v1/ping
   ```
   Should return: `{"message":"pong"}`

### Test Credentials:
You'll need to create a test user first using Postman or curl:

#### Create a User (via Postman):
```
POST http://localhost:8080/api/v1/auth/user/register
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "password123",
  "name": "Test User"
}
```

Then use these credentials in the UI to login!

## 🎯 Features Breakdown

### User Type Toggle
- Switch between **User Login** and **Admin Login**
- Different form titles and subtitles for each type
- Separate API endpoints for each user type

### Form Validation
- **Email:** Must be valid email format
- **Password:** Minimum 6 characters
- Real-time validation on blur
- Clear error messages

### Session Management
- Stores `accessToken` in localStorage
- Stores `refreshToken` in localStorage
- Stores `userType` (user/admin)
- Stores `userEmail` and `userData`
- Auto-redirect if already logged in

### Alert System
- ✓ Success alerts (green)
- ✕ Error alerts (red)
- ⚠ Warning alerts (yellow)
- Auto-dismiss for non-error messages

## 🎨 Customization

### Colors
Edit CSS variables in `styles.css`:
```css
:root {
    --primary-gradient: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    --bg-primary: #0f0f23;
    --text-primary: #ffffff;
    /* ... more variables */
}
```

### Animations
All animations are defined in `styles.css`:
- `float` - Background orb animation
- `fadeInUp` - Element entrance animation
- `pulse` - Logo pulse effect
- `spin` - Loading spinner
- `shake` - Error shake effect

## 🐛 Troubleshooting

### CORS Issues
If you get CORS errors, make sure your gateway has CORS middleware enabled (it should already be configured in your Go code).

### API Connection Failed
1. Check if Docker services are running: `docker ps`
2. Check gateway logs: `docker logs gateway-service`
3. Verify API endpoint in `script.js`

### Login Not Working
1. Check browser console for errors (F12)
2. Verify API response in Network tab
3. Make sure user exists in database
4. Check backend logs for authentication errors

## 📱 Responsive Breakpoints

- **Desktop:** > 1024px (Full two-column layout)
- **Tablet:** 768px - 1024px (Stacked layout)
- **Mobile:** < 768px (Optimized mobile view)

## 🔒 Security Notes

- Passwords are sent over HTTPS (use HTTPS in production!)
- Tokens are stored in localStorage (consider httpOnly cookies for production)
- No sensitive data in console logs (remove in production)
- Input validation on both client and server side

## 🚀 Next Steps

After successful login, you can:
1. Create dashboard pages (`/user/dashboard.html`, `/admin/dashboard.html`)
2. Add protected routes
3. Implement token refresh logic
4. Add user profile management
5. Create order management interfaces

## 📞 Support

For issues or questions:
- Check the browser console (F12) for errors
- Review the API documentation
- Check Docker logs: `docker logs gateway-service`

---

**Built with ❤️ for Order Management System**
