# TaskFlow API

A REST API for task management built with Go, PostgreSQL, and Docker.

## Overview

TaskFlow lets users register, log in, create projects, add tasks, and assign them. Built for the Greening India backend engineering assignment.

**Stack:** Go, chi router, PostgreSQL, golang-migrate, JWT, bcrypt, Docker

---

## Architecture Decisions

- **`internal/` package layout** — Go convention to keep packages private to the module. Handlers, repositories, models and middleware are fully separated.
- **Repository pattern** — All SQL lives in `internal/repository/`. Handlers never touch SQL directly. This makes the code testable and the DB swappable.
- **`chi` router** — Lightweight, idiomatic Go router. Supports middleware groups cleanly, which made protecting all non-auth routes simple.
- **`slog`** — Go's built-in structured logger (added in Go 1.21). No external dependency needed.
- **`golang-migrate`** — Explicit up/down migrations over ORM auto-migrate. Gives full control over schema changes and is production-safe.
- **JWT in middleware** — Auth is handled in a single middleware applied to a route group. No handler needs to repeat auth logic.
- **Graceful shutdown** — The server listens for SIGTERM/SIGINT and drains in-flight requests before exiting.

**Tradeoffs:**
- No ORM — raw `database/sql` is more verbose but explicit. You see exactly what SQL runs.
- No refresh tokens — JWT is 24hr only. Good enough for this scope.
- Seed runs via migrations — simple and automatic, no separate script needed.

---

## Running Locally

Requirements: Docker Desktop (nothing else needed)

```bash
git clone https://github.com/saisushantht/taskflow
cd taskflow
docker compose up
```

API is available at `http://localhost:8080`

To stop:
```bash
docker compose down
```

To reset the database completely:
```bash
docker compose down -v
docker compose up
```

---

## Running Migrations

Migrations run **automatically on startup**. No manual steps needed.

Migration files are in `migrations/`. Each has an up and down file.

---

## Test Credentials

Seeded automatically on first run:

```
Email:    test@example.com
Password: password123
```

---

## API Reference

All endpoints return `Content-Type: application/json`.

Protected endpoints require: `Authorization: Bearer <token>`

### Auth

#### POST /auth/register
```json
// Request
{ "name": "Jane Doe", "email": "jane@example.com", "password": "secret123" }

// Response 201
{ "token": "<jwt>", "user": { "id": "uuid", "name": "Jane Doe", "email": "jane@example.com", "created_at": "..." } }
```

#### POST /auth/login
```json
// Request
{ "email": "jane@example.com", "password": "secret123" }

// Response 200
{ "token": "<jwt>", "user": { ... } }
```

### Projects

#### GET /projects
Returns projects the user owns or has tasks in.
```json
{ "projects": [ { "id": "uuid", "name": "...", "description": "...", "owner_id": "uuid", "created_at": "..." } ] }
```

#### POST /projects
```json
// Request
{ "name": "My Project", "description": "Optional" }
// Response 201
{ "id": "uuid", "name": "My Project", ... }
```

#### GET /projects/:id
Returns project with all its tasks.
```json
{ "id": "uuid", "name": "...", "tasks": [ ... ] }
```

#### PATCH /projects/:id
Owner only. Partial update supported.
```json
// Request
{ "name": "Updated Name" }
// Response 200 — updated project
```

#### DELETE /projects/:id
Owner only. Deletes project and all its tasks.
```
Response 204 No Content
```

#### GET /projects/:id/stats
```json
{
  "by_status": { "todo": 2, "in_progress": 1, "done": 3 },
  "by_assignee": { "Jane Doe": 4, "John Smith": 2 }
}
```

### Tasks

#### GET /projects/:id/tasks
Supports filters and pagination.
```
?status=todo|in_progress|done
?assignee=<user_uuid>
?page=1&limit=20
```
```json
{
  "tasks": [ { "id": "uuid", "title": "...", "status": "todo", "priority": "high", ... } ],
  "pagination": { "page": 1, "limit": 20, "total": 5 }
}
```

#### POST /projects/:id/tasks
```json
// Request
{ "title": "Fix bug", "description": "...", "priority": "high", "status": "todo", "assignee_id": "uuid", "due_date": "2026-04-20" }
// Response 201 — created task
```

#### PATCH /tasks/:id
All fields optional.
```json
// Request
{ "status": "done" }
// Response 200 — updated task
```

#### DELETE /tasks/:id
Project owner only.
```
Response 204 No Content
```

### Error Responses

```json
{ "error": "validation failed", "fields": { "email": "is required" } }
{ "error": "unauthorized" }
{ "error": "forbidden" }
{ "error": "not found" }
```

---

## What I'd Do With More Time

- **Refresh tokens** — current JWT is 24hr with no refresh. A refresh token flow would be needed for production.
- **Input sanitization** — add more thorough validation (email format, enum validation for status/priority at handler level).
- **Integration tests** — would add full test coverage for auth, project ownership rules, and task filters.
- **Rate limiting** — add per-IP rate limiting on auth endpoints to prevent brute force.
- **Pagination on projects** — currently only tasks are paginated.
- **Task creator tracking** — deletion currently checks project ownership but doesn't track who created each task.

---

## Project Structure

```
taskflow/
├── cmd/api/main.go          # Entry point — wires everything together
├── internal/
│   ├── config/              # Env var loading
│   ├── db/                  # DB connection + migrations runner
│   ├── handler/             # HTTP handlers (auth, projects, tasks)
│   ├── middleware/          # JWT auth middleware
│   ├── model/               # Structs (User, Project, Task)
│   ├── repository/          # All SQL queries
│   └── util/                # JSON response helpers
├── migrations/              # Up + down SQL files
├── Dockerfile               # Multi-stage build
├── docker-compose.yml
└── .env.example
```
