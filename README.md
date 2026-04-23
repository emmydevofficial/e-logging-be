# E-Logging API

A REST API for electronic logging built with Go, Fiber, and PostgreSQL.

## Setup

1. Set up PostgreSQL database and run the migration:
   ```sql
   -- Run the SQL in migrations/001_initial_schema.sql
   ```

2. Set environment variables:
   - `DB_URL`: PostgreSQL connection string
   - `JWT_SECRET`: Secret key for JWT tokens
   - `OPENAI_API_KEY`: API key for OpenAI Whisper
   - `PORT`: Server port (default 3000)

3. Run the server:
   ```bash
   go run cmd/server/main.go
   ```

## API Endpoints

- `POST /api/auth/login` - Login and get JWT tokens
- `POST /api/auth/refresh` - Refresh access token
- `GET /api/logs` - Get paginated logs
- `POST /api/logs` - Create a new log (operator/admin + device fingerprint)
- `PUT /api/logs/:id` - Update log (creator only, within 24 hours)
- `GET /api/logs/export` - Export logs as CSV (downloader/admin)
- `GET /api/stations` - List stations
- `POST /api/stations` - Create station (admin)
- `GET /api/devices` - List devices (admin)
- `POST /api/devices` - Register device (admin)
- `DELETE /api/devices/:id` - Deactivate device (admin)
- `GET /api/users` - List users (admin)
- `POST /api/users` - Create user (admin)
- `POST /api/stt` - Transcribe audio using OpenAI Whisper