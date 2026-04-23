# Quick Start Setup Guide

## Prerequisites
- PostgreSQL running locally
- Go 1.20+
- All dependencies installed (`go mod download` completed)

## 1. Database Setup

```bash
# Create database
createdb e_logging_db

# Apply schema
psql -U postgres -d e_logging_db -f migrations/001_initial_schema.sql

# Verify tables created
psql -U postgres -d e_logging_db -c "\dt"
```

## 2. Configure Environment

```bash
# Copy example and update
cp .env.example .env

# Edit .env with your credentials
nano .env
```

**Update these values in .env:**
```
DB_URL=postgresql://postgres:your_password@localhost:5432/e_logging_db?sslmode=disable
JWT_SECRET=your-super-secret-key-min-32-chars
OPENAI_API_KEY=sk-your-key (optional for now)
PORT=3000
```

## 3. Create Initial Admin User

### Option A: Using the API (Recommended)

1. **Start the server:**
   ```bash
   go run ./cmd/server
   ```

2. **Uncomment admin endpoint in `cmd/server/main.go`:**
   
   Find this section:
   ```go
   // ========== ADMIN SETUP ENDPOINT - COMMENT OUT AFTER FIRST ADMIN IS CREATED ==========
   // adminGroup := app.Group("/api/admin")
   // adminGroup.Post("/create-admin", userHandler.CreateAdminUser)
   ```
   
   Uncomment it to:
   ```go
   // ========== ADMIN SETUP ENDPOINT - COMMENT OUT AFTER FIRST ADMIN IS CREATED ==========
   adminGroup := app.Group("/api/admin")
   adminGroup.Post("/create-admin", userHandler.CreateAdminUser)
   ```

3. **Rebuild and start:**
   ```bash
   go build ./cmd/server
   go run ./cmd/server
   ```

4. **Create admin in another terminal:**
   ```bash
   curl -X POST http://localhost:3000/api/admin/create-admin \
     -H "Content-Type: application/json" \
     -d '{
       "name": "Administrator",
       "email": "admin@example.com",
       "password": "Admin@123456"
     }'
   ```

5. **Comment out the endpoint again** in `cmd/server/main.go`:
   ```go
   // adminGroup := app.Group("/api/admin")
   // adminGroup.Post("/create-admin", userHandler.CreateAdminUser)
   ```

6. **Rebuild and restart server**

### Option B: Direct Database Insert

```bash
# Generate bcrypt hash of your password using Go
go run -c 'package main; import ("fmt"; "golang.org/x/crypto/bcrypt"); func main() { h, _ := bcrypt.GenerateFromPassword([]byte("your-password"), 10); fmt.Println(string(h)) }'
```

Then insert into database:
```sql
INSERT INTO users (id, name, email, password_hash, role, created_at)
VALUES (
  gen_random_uuid(),
  'Administrator',
  'admin@example.com',
  '$2a$10$...',  -- paste hash from above
  'admin',
  NOW()
);
```

## 4. Login and Verify

```bash
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "Admin@123456"
  }'
```

Expected response:
```json
{
  "success": true,
  "data": {
    "access_token": "eyJ...",
    "refresh_token": "eyJ..."
  }
}
```

## 5. Access Swagger Documentation

Visit: `http://localhost:3000/swagger/index.html`

All endpoints will be documented with:
- Request/Response formats
- Authorization requirements
- Query parameters
- Error responses

## Security Checklist

- ✅ Admin endpoint is commented out after first admin created
- ✅ .env file added to .gitignore
- ✅ JWT_SECRET is strong and unique
- ✅ Database only accepts local connections
- ✅ Never commit .env to git

## Next Steps

1. Create other users (operator, downloader, etc.) using the authenticated `/api/users` endpoint
2. Set up devices with device fingerprints
3. Create stations for log tracking
4. Start creating logs via the API

## Useful Commands

```bash
# Build the project
go build ./cmd/server

# Run the server
go run ./cmd/server

# Run tests (if added)
go test ./...

# Check for lint issues
go vet ./...

# Regenerate Swagger docs (after adding new handlers)
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/server/main.go
```

## Troubleshooting

### "Failed to connect to database"
- Verify PostgreSQL is running: `psql -U postgres`
- Check DB_URL in .env file
- Ensure database and schema are created

### "Admin user already exists"
- The email already has an account
- Use different email or login with existing credentials

### "JWT_SECRET is required"
- Add JWT_SECRET to .env file
- Restart the server

See `ADMIN_SETUP.md` for detailed admin creation instructions.
