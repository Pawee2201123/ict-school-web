# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go web application for a mock lesson reservation system (模擬授業予約システム). The application uses:
- Go 1.23.8 with standard library HTTP server
- PostgreSQL 15 database
- Nix for reproducible builds
- Docker for deployment
- HTML templates (no frontend framework)

## Development Commands

### Setup
1. **Copy environment template**:
```bash
cp .env.example .env
```

2. **Edit .env** with your actual credentials (NEVER commit this file)

3. **Enter the Nix development shell** (automatically installs Go and PostgreSQL):
```bash
nix develop
```
The shell will automatically load environment variables from `.env` if it exists.

### Running Locally
```bash
go run ./cmd/server/main.go
```
Server starts on `http://localhost:8080`

### Building for Production
Build the Docker image using Nix (creates byte-for-byte reproducible image):
```bash
nix build .#docker
```
This creates a `result` symlink pointing to the compressed Docker image archive.

### Deployment
1. Transfer files to server:
```bash
scp -i key.pem result ec2-user@SERVER_IP:/home/ec2-user/ict-web.tar.gz
scp -i key.pem docker-compose.yml init.sql .env ec2-user@SERVER_IP:/home/ec2-user/
```

2. On server, **edit .env** with production credentials, then:
```bash
docker load < ict-web.tar.gz
docker compose up -d
```

### Docker Operations
```bash
docker compose ps          # Check service status
docker compose logs web    # View web app logs
docker compose logs db     # View database logs
docker compose down        # Stop all services
```

## Architecture

### Project Structure
```
cmd/server/main.go          - Application entry point, routes setup
internal/
  auth/                     - Password hashing and session management
  config/                   - Environment variable loading
  database/                 - Database connection setup
  handlers/                 - HTTP handlers (admin, student, auth)
  models/                   - Database models and queries
  template/                 - HTML template rendering
web/
  templates/                - HTML template files
  static/                   - CSS, JS, and uploaded files
```

### Application Flow

**Entry Point**: `cmd/server/main.go`
1. Loads config from environment variables via `config.Load()`
2. Connects to PostgreSQL via `database.Connect()`
3. Ensures admin user exists via `handlers.EnsureAdmin()`
4. Loads HTML templates from `web/templates`
5. Sets up HTTP routes with middleware
6. Starts server on port 8080

**Route Protection**:
- `RequireLogin` middleware: Validates session cookie
- `RequireAdmin` middleware: Checks `is_admin` flag
- Admin routes use both: `h.RequireLogin(h.RequireAdmin(next))`

**Session Management**:
- Uses `gorilla/securecookie` for encrypted cookies
- Session key: "session"
- Stores `user_id`, `email`, `is_admin` in cookie payload
- No server-side session storage

**Database Schema** (see `init.sql`):
- `users` - Authentication (email, password_hash, is_admin)
- `user_profiles` - Student/guardian information
- `classes` - Course definitions with registration windows
- `class_sessions` - Individual class sessions with capacity tracking (has `day_sequence` field: 1 or 2)
- `session_enrollments` - Student registrations
- `instructors` - Teacher information
- `system_settings` - Key-value store for event dates

**Enrollment Business Rules**:
- Students can enroll in maximum **2 classes per day**
- Students can enroll in maximum **3 classes total** across both days
- Validation happens in `models.CheckEnrollmentLimits()` before enrollment
- Japanese error messages are shown when limits are exceeded

### Key Patterns

**Handler Pattern**:
All handlers are methods on `Handler` struct which holds:
- `db *sql.DB` - Database connection
- `tpl *template.Renderer` - Template renderer
- `cfg *config.Config` - Configuration

**Model Pattern**:
Models in `internal/models/` contain:
- Struct definitions
- Database query functions (e.g., `GetAllClasses`, `CreateUser`)
- No business logic (that stays in handlers)

**Template Rendering**:
```go
h.tpl.Render(w, "template_name.html", data)
```
Templates located in `web/templates/`. Uses Go's `html/template`.

### Secret Management

All secrets and configuration are stored in a `.env` file which is **never committed to git**.

**Setup**:
1. Copy `.env.example` to `.env`
2. Edit `.env` with your actual credentials
3. `.env` is automatically loaded by:
   - `nix develop` (for local development)
   - `docker compose` (for deployment)

**Security Notes**:
- `.env` is in `.gitignore` - never commit it
- Use `.env.example` as a template for new environments
- For production: Use strong passwords and real SMTP credentials
- For cookie keys: Generate with `openssl rand -hex 32`

### Environment Variables

**Required**:
- `DATABASE_URL` - PostgreSQL connection string (format: `postgres://user:pass@host:5432/dbname?sslmode=disable`)

**Optional**:
- `LISTEN_ADDR` - Server address (default: `:8080`)
- `ADMIN_EMAIL` - Initial admin email
- `ADMIN_PASSWORD` - Initial admin password
- `UPLOAD_DIR` - File upload directory (default: `./web/static/uploads`)
- `SMTP_HOST` - SMTP server hostname (e.g., `smtp.gmail.com`)
- `SMTP_PORT` - SMTP server port (default: `587`)
- `SMTP_USERNAME` - SMTP authentication username
- `SMTP_PASSWORD` - SMTP authentication password (use App Password for Gmail)
- `SMTP_FROM` - Email sender address

### Nix Configuration

**flake.nix** defines:
- Dev shell with Go + PostgreSQL
- Application build (compiles Go, bundles `web/` directory)
- Docker image (includes app, SSL certs, timezone data)

**Important**: When Go dependencies change:
1. Set `vendorHash = pkgs.lib.fakeHash` in flake.nix
2. Run `nix build`
3. Copy the real hash from error message
4. Update `vendorHash` with real value

### Database Connection

The app expects PostgreSQL to be available. In docker-compose, the service is named `db`, so use `@db:5432` not `@localhost:5432` in `DATABASE_URL`.

Initial admin user is created automatically on first run using `ADMIN_EMAIL` and `ADMIN_PASSWORD` environment variables (see `handlers.EnsureAdmin`).

### File Uploads

Syllabus PDFs are uploaded to `UPLOAD_DIR` (default: `./web/static/uploads`). In Docker, this is mounted to a persistent volume at `/data/uploads`.

Files are served via `/uploads/` route which maps to the upload directory.

### Module Name

The Go module is `example.com/myapp` (see `go.mod`). All internal imports use this prefix:
```go
import "example.com/myapp/internal/handlers"
```

### Email Notifications

The application sends enrollment confirmation emails to students after they successfully register for a class session.

**Implementation**:
- Located in `internal/email/` package
- Uses `gomail` library for SMTP
- Sends emails asynchronously (non-blocking)
- Supports Japanese content with UTF-8 encoding

**Email Flow**:
1. Student completes enrollment via `StudentApplication` handler
2. After successful database insertion, email is sent in background goroutine
3. If email fails, error is logged but enrollment still succeeds
4. Email contains: class name, date/time, room, instructor, and student name

**Configuration**:
- If SMTP settings are not configured, emails are skipped (logged only)
- For Gmail: Use App Password, not regular password
- For production: Consider AWS SES (already on AWS infrastructure)

**Testing Locally**:
Use Mailhog for local SMTP testing:
```bash
# Run Mailhog in Docker
docker run -d -p 1025:1025 -p 8025:8025 mailhog/mailhog

# Configure in docker-compose.yml or .env
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_FROM=test@localhost

# View emails at http://localhost:8025
```
