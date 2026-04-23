# Admin User Setup Guide

## Initial Setup (First Time Only)

### Step 1: Start the Server
```bash
go run ./cmd/server
```

### Step 2: Create Admin User
Make a POST request to create the initial admin user:

```bash
curl -X POST http://localhost:3000/api/admin/create-admin \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Administrator",
    "email": "admin@example.com",
    "password": "YourSecurePassword123!"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Admin user created successfully. IMPORTANT: Comment out this endpoint in main.go after creation!",
  "data": {
    "id": "uuid-here",
    "name": "Administrator",
    "email": "admin@example.com",
    "role": "admin"
  }
}
```

### Step 3: Comment Out the Admin Endpoint

**⚠️ IMPORTANT SECURITY STEP:**

After creating your first admin user, **immediately** comment out the admin creation endpoint in `cmd/server/main.go`:

```go
// ADMIN SETUP - COMMENT OUT AFTER CREATING FIRST ADMIN
// adminGroup := app.Group("/api/admin")
// adminGroup.Post("/create-admin", userHandler.CreateAdminUser)
```

### Step 4: Login with Admin Credentials
```bash
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "YourSecurePassword123!"
  }'
```

This will return access and refresh tokens.

## Adding the Route in main.go

### Location: `cmd/server/main.go`

Add this code after your auth routes and before protected routes:

```go
// ========== ADMIN SETUP ENDPOINT - COMMENT OUT AFTER FIRST ADMIN IS CREATED ==========
// IMPORTANT: Only uncomment this for initial setup. Disable after creating the first admin.
adminGroup := app.Group("/api/admin")
adminGroup.Post("/create-admin", userHandler.CreateAdminUser)
// =====================================================================================
```

### Example in context:

```go
func main() {
    // ... database and handler initialization ...

    // Auth routes
    authGroup := app.Group("/api/auth")
    authGroup.Post("/login", authHandler.Login)
    authGroup.Post("/refresh", authHandler.Refresh)

    // ========== ADMIN SETUP ENDPOINT - COMMENT OUT AFTER FIRST ADMIN IS CREATED ==========
    adminGroup := app.Group("/api/admin")
    adminGroup.Post("/create-admin", userHandler.CreateAdminUser)
    // =====================================================================================

    // Protected routes
    api := app.Group("/api", middleware.JWTMiddleware())
    // ... rest of routes ...
}
```

## Security Best Practices

1. **Disable After Use**: Comment out or remove the admin creation endpoint after creating the first admin
2. **Use Strong Passwords**: Create a strong password for your first admin account
3. **Never Deploy with Active**: Do NOT deploy to production with this endpoint active
4. **Change Email**: Update the admin email from the example to your actual email
5. **Verify Swagger**: After commenting out, verify the endpoint no longer appears in Swagger UI at `/swagger/index.html`

## Troubleshooting

### "User may already exist" Error
- The admin user already exists in the database
- Comment out the endpoint if you haven't already
- Use the `/auth/login` endpoint to authenticate

### "No Go files" Warning
- This is safe to ignore; it occurs during Swagger generation
- Your application will still run and function normally

### Need to Create Another Admin?
- Uncomment the endpoint temporarily
- Create the new admin user
- Comment it back out immediately

## Alternative: Database Direct Insert

If you prefer, you can also create an admin user directly in PostgreSQL:

```sql
INSERT INTO users (id, name, email, password_hash, role, created_at)
VALUES (
  gen_random_uuid(),
  'Administrator',
  'admin@example.com',
  '$2a$10$...', -- bcrypt hash of password
  'admin',
  NOW()
);
```

Use an online bcrypt generator or Go's `golang.org/x/crypto/bcrypt` package to hash the password first.
