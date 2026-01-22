# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Gophish is an open-source phishing toolkit written in Go. This is a custom fork modified to work with the Evilginx phishing framework. It provides a web-based admin interface for managing phishing campaigns, target groups, email templates, and landing pages.

## Build and Run Commands

```bash
# Build the application (requires Go 1.22+)
go build

# Run the application (reads config.json by default)
./gophish

# Run with custom config
./gophish --config /path/to/config.json

# Run with mailer disabled (for multi-system deployments)
./gophish --disable-mailer

# Run in specific mode (all, admin, or phish)
./gophish --mode admin

# Run tests
go test ./...

# Run tests for a specific package
go test ./models/...
go test ./controllers/...
```

## Architecture

### Core Packages

- **gophish.go** - Main entry point. Initializes config, database, and servers. Supports `--mode` flag for running admin-only, phish-only, or both.

- **models/** - Database models and ORM layer using GORM. Handles campaigns, groups, pages, templates, users, SMTP profiles, results, and webhooks. Uses goose for migrations (`db/db_sqlite3/` and `db/db_mysql/`).

- **controllers/** - HTTP handlers split into:
  - `route.go` - AdminServer with web UI routes (login, campaigns, templates, etc.)
  - `api/server.go` - REST API server mounted at `/api/`
  - `phish.go` - Phishing server handlers (currently commented out in main)

- **worker/** - Background worker that polls for queued maillogs and processes campaign emails. Runs on 1-minute intervals.

- **mailer/** - SMTP email sending with connection pooling per campaign.

- **middleware/** - Authentication, CSRF protection, rate limiting, session management.

- **evilginx/** - Helpers for Evilginx integration including URL parameter encryption/decryption using RC4 and AES-CBC.

### Request Flow

1. Admin UI requests go through `AdminServer` (controllers/route.go) which serves HTML templates from `templates/`
2. API requests route through `/api/` to `api.Server` (controllers/api/server.go)
3. Both require authentication via session cookies (web) or API key header
4. Campaign launching triggers the worker to queue and send emails via the mailer

### Database

- Default: SQLite3 (`gophish.db`)
- Also supports MySQL with TLS
- Migrations in `db/db_sqlite3/migrations/` and `db/db_mysql/migrations/`
- Initial admin user created on first run with randomly generated password (logged to console)

### Configuration

Configuration via `config.json`:
- `admin_server` - Admin web interface settings (default: https://127.0.0.1:3333)
- `phish_server` - Phishing server settings (default: http://0.0.0.0:80)
- `db_name` - Database type: "sqlite3" or "mysql"
- `db_path` - Database connection string

Environment variables:
- `GOPHISH_INITIAL_ADMIN_PASSWORD` - Set initial admin password instead of random
- `GOPHISH_INITIAL_ADMIN_API_TOKEN` - Set initial admin API token instead of random

### Static Assets

- `static/` - Frontend JavaScript, CSS, images
- `templates/` - Go HTML templates for admin UI
