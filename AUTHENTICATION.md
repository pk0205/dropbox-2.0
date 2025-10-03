# Authentication Guide

## Cookie-Based Authentication

This API uses **HTTP-only cookies** for authentication. No Bearer tokens or Authorization headers required!

## Why Cookies?

✅ **More Secure**

- HTTP-only cookies cannot be accessed by JavaScript (prevents XSS attacks)
- Automatic CSRF protection with SameSite policy
- Browser handles storage and transmission automatically

✅ **Simpler Client Code**

- No need to manually manage tokens
- No need to add Authorization headers to every request
- Just use `credentials: "include"` in fetch requests

❌ **Why Not Bearer Tokens?**

- Can be stolen via XSS attacks if stored in localStorage
- Require manual header management on every request
- More complex client-side code

## How It Works

### 1. Sign Up or Login

```javascript
// Sign up
const response = await fetch("/api/user/signup", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  credentials: "include", // Important!
  body: JSON.stringify({
    firstName: "John",
    lastName: "Doe",
    username: "johndoe",
    email: "john@example.com",
    password: "securepassword",
  }),
});

// Cookie is automatically set by server
```

### 2. Make Authenticated Requests

```javascript
// Upload file - cookie is sent automatically
const response = await fetch("/api/files/upload", {
  method: "POST",
  credentials: "include", // This sends the cookie
  body: formData,
});
```

### 3. Logout

```javascript
const response = await fetch("/api/user/logout", {
  method: "POST",
  credentials: "include",
});

// Cookie is cleared by server
```

## Cookie Details

The `AuthToken` cookie has these properties:

```go
Cookie{
  Name:     "AuthToken",
  Value:    "<JWT token>",
  Expires:  30 days,
  HTTPOnly: true,     // Cannot be accessed by JavaScript
  Secure:   false,    // Set to true in production (HTTPS only)
  SameSite: "Lax",    // CSRF protection
  Path:     "/",      // Available to all routes
}
```

## Client Examples

### JavaScript/Browser

```javascript
// Always include credentials: 'include'
async function uploadFile(file) {
  const formData = new FormData();
  formData.append("file", file);

  const response = await fetch("/api/files/upload", {
    method: "POST",
    credentials: "include", // ← This is the key!
    body: formData,
  });

  return response.json();
}
```

### cURL

```bash
# Login and save cookie
curl -X POST http://localhost:3000/api/user/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{"emailOrUsername":"user","password":"pass"}'

# Use cookie for subsequent requests
curl -X POST http://localhost:3000/api/files/upload \
  -b cookies.txt \
  -F "file=@document.pdf"
```

### Python (requests)

```python
import requests

# Create session to persist cookies
session = requests.Session()

# Login
session.post('http://localhost:3000/api/user/login', json={
    'emailOrUsername': 'user',
    'password': 'password'
})

# Cookie is automatically stored and sent
response = session.post('http://localhost:3000/api/files/upload',
    files={'file': open('document.pdf', 'rb')}
)
```

### Axios (JavaScript)

```javascript
import axios from "axios";

// Configure axios to send cookies
const api = axios.create({
  baseURL: "http://localhost:3000",
  withCredentials: true, // ← This enables cookies
});

// Login
await api.post("/api/user/login", {
  emailOrUsername: "user",
  password: "password",
});

// Upload - cookie sent automatically
const formData = new FormData();
formData.append("file", file);
await api.post("/api/files/upload", formData);
```

## Security Best Practices

### Production Configuration

Update your cookie settings for production:

```go
// In handlers/user.go
c.Cookie(&fiber.Cookie{
    Name:     "AuthToken",
    Value:    tokenString,
    Expires:  time.Now().Add(time.Hour * 24 * 30),
    HTTPOnly: true,
    Secure:   true,  // ← Enable for HTTPS
    SameSite: "Strict", // ← Stronger CSRF protection
    Path:     "/",
    Domain:   "yourdomain.com", // ← Set your domain
})
```

### HTTPS Only

Always use HTTPS in production:

```go
// main.go
if os.Getenv("ENV") == "production" {
    log.Fatal(app.ListenTLS(":443", "./cert.pem", "./key.pem"))
} else {
    log.Fatal(app.Listen(":" + PORT))
}
```

### CORS Configuration

Ensure CORS allows credentials:

```go
app.Use(cors.New(cors.Config{
    AllowOrigins:     "https://yourdomain.com", // Specific domain, not *
    AllowCredentials: true, // ← Required for cookies
    AllowHeaders:     "Origin, Content-Type, Accept",
}))
```

## Troubleshooting

### Cookies Not Being Sent

**Problem:** Requests return 401 Unauthorized

**Solutions:**

1. Add `credentials: 'include'` to fetch requests
2. Use `withCredentials: true` in axios
3. Check CORS has `AllowCredentials: true`
4. Ensure frontend and backend are on same domain (or use proper CORS)

### Cookie Not Set on Login

**Problem:** Cookie doesn't appear after login

**Solutions:**

1. Check browser developer tools → Application → Cookies
2. Ensure `SameSite` policy is compatible
3. Use HTTPS in production (required for Secure cookies)
4. Check cookie domain matches your domain

### CORS Errors

**Problem:** CORS policy blocking requests

**Solutions:**

```go
// Allow specific origin, not wildcard with credentials
AllowOrigins:     "http://localhost:5173",
AllowCredentials: true,
```

### Cross-Domain Cookies

**Problem:** Frontend and backend on different domains

**Solutions:**

1. Best: Deploy both on same domain
2. Alternative: Use subdomain (api.yourdomain.com)
3. Last resort: Proxy frontend requests to backend

## Token Expiration

Tokens expire after **30 days**. Handle this gracefully:

```javascript
async function makeAuthenticatedRequest(url, options) {
  const response = await fetch(url, {
    ...options,
    credentials: "include",
  });

  if (response.status === 401) {
    // Token expired, redirect to login
    window.location.href = "/login";
  }

  return response;
}
```

## Logout Implementation

Server clears the cookie:

```go
func Logout() fiber.Handler {
    return func(c *fiber.Ctx) error {
        c.Cookie(&fiber.Cookie{
            Name:     "AuthToken",
            Value:    "",
            Expires:  time.Now().Add(-time.Hour), // Past time
            HTTPOnly: true,
        })

        return c.Status(200).JSON(fiber.Map{
            "message": "Logged out successfully",
        })
    }
}
```

Client just needs to call the endpoint:

```javascript
async function logout() {
  await fetch("/api/user/logout", {
    method: "POST",
    credentials: "include",
  });

  // Redirect to login page
  window.location.href = "/login";
}
```

## Testing Authentication

### Test Cookie is Set

```bash
# Login and check cookie
curl -v -X POST http://localhost:3000/api/user/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{"emailOrUsername":"user","password":"pass"}'

# Look for "Set-Cookie: AuthToken=..." in response headers
```

### Test Authenticated Endpoint

```bash
# Should work with cookie
curl -b cookies.txt http://localhost:3000/api/files

# Should fail without cookie
curl http://localhost:3000/api/files
```

## FAQ

**Q: Can I use both cookies and Bearer tokens?**
A: You could, but we recommend sticking with one method for simplicity. Cookies are more secure for browser-based apps.

**Q: How do I use this in a mobile app?**
A: Mobile apps can still use cookies! Most HTTP clients support cookie jars. Or switch to Bearer tokens for mobile if preferred.

**Q: What about API testing tools like Postman?**
A: Postman automatically handles cookies. Just login first, then subsequent requests will include the cookie.

**Q: Can I refresh tokens?**
A: Current implementation uses 30-day tokens. For refresh tokens, implement a `/api/user/refresh` endpoint that validates the old cookie and issues a new one.

---

**Summary:** Use `credentials: 'include'` in all fetch requests, and the browser handles the rest! 🍪
