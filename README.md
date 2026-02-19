# UK Home Office Sponsor Licence Tracker

Tracks UK Home Office sponsor licence data, synced from gov.uk. Built with Go, PostgreSQL, and React.

## Prerequisites

- [Go](https://go.dev/) 1.22+
- [PostgreSQL](https://www.postgresql.org/) 15+
- [Node.js](https://nodejs.org/) 20+
- [goose](https://github.com/pressly/goose) (database migrations)

Install goose:
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

## Setup

### 1. Create databases

```sql
CREATE DATABASE sponsor_licence;
CREATE DATABASE sponsor_licence_test;
```

### 2. Configure environment

```bash
cd backend
cp .env.example .env
```

Edit `.env` and set `DATABASE_PASSWORD` to your PostgreSQL password.

### 3. Run migrations

```bash
cd backend
goose -dir migrations postgres "postgres://postgres:PASSWORD@localhost:5432/sponsor_licence" up
```

Replace `PASSWORD` with your actual password.

### 4. Install frontend dependencies

```bash
cd frontend
npm install
```

## Building

```bash
cd backend
go build ./...
```

## Testing

Backend tests use the `sponsor_licence_test` database (see `config.yaml`).

```bash
cd backend
go test ./...
```

## Running

### API server (port 8080)

```bash
cd backend
go run ./cmd/api
```

### Frontend dev server

```bash
cd frontend
npm run dev
```

## Create a user

The first admin user must be created using the CLI tool. Run from the `backend/` directory:

```bash
cd backend
go run ./cmd/createuser
```

You will be prompted for a username and password. The user is created with role 10 (admin) by default. To create a viewer:

```bash
go run ./cmd/createuser -role 50
```

## Trigger a sync

Fetches the latest sponsor licence CSV from gov.uk and updates the database. Run from the `backend/` directory:

```bash
cd backend
go run ./cmd/sync
```

## API Reference

### Authentication

| Method | Path | Auth required | Description |
|--------|------|---------------|-------------|
| `POST` | `/api/auth/login` | None | Login. Sets `session_token` cookie. |
| `POST` | `/api/auth/logout` | None | Logout. Clears `session_token` cookie. |
| `GET` | `/api/auth/me` | Any | Returns the current authenticated user. |

**POST /api/auth/login** — request body:
```json
{ "username": "admin", "password": "secret" }
```

**GET /api/auth/me** — response:
```json
{ "id": 1, "username": "admin", "role": 10 }
```

Sessions expire after 15 minutes of inactivity. Each authenticated request extends the session.

---

### Data

| Method | Path | Auth required | Description |
|--------|------|---------------|-------------|
| `GET` | `/api/data` | None | Returns paginated sponsor licence data. |
| `POST` | `/api/sync` | Admin (role ≤ 10) | Fetches latest data from gov.uk and updates the database. |

**GET /api/data** — query parameters:

| Parameter | Required | Constraints | Description |
|-----------|----------|-------------|-------------|
| `from` | Yes | 1 – 1,000,000,000 | First row index (1-based). |
| `to` | Yes | ≥ `from`, `to − from + 1 ≤ 100` | Last row index. Maximum page size is 100. |
| `search` | No | Max 200 characters | Filters by organisation name or town/city (case-insensitive). |

Example:
```
GET /api/data?from=1&to=50&search=london
```

## Roles

| Value | Name | Access |
|-------|------|--------|
| `10` | Admin | All endpoints, including `/api/sync`. |
| `50` | Viewer | Read-only data access. |

## Project structure

```
backend/
  cmd/
    api/            Main API server
    createuser/     CLI tool for creating users
    sync/           CLI tool for triggering a data sync
  internal/
    api/            HTTP handlers and middleware
    auth/           Authentication service and session management
    config/         Configuration loading (config.yaml + .env)
    csvfetch/       Gov.uk CSV discovery and parsing
    database/       Database types and queries
    sync/           Data sync orchestration
  migrations/       Goose SQL migrations
frontend/
  src/              React + TypeScript frontend
```
