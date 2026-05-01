# HireFlow API

A production-grade **Job Recruitment REST API** built with Go, following Clean Architecture principles. Companies post jobs, candidates apply, and recruiters manage hiring pipelines — all secured with JWT-based role-based access control.

---

## Table of Contents

- [Overview](#overview)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Database Design](#database-design)
- [API Endpoints](#api-endpoints)
- [Roles & Permissions](#roles--permissions)
- [Getting Started](#getting-started)
- [Environment Variables](#environment-variables)
- [Running Tests](#running-tests)
- [Application Pipeline](#application-pipeline)

---

## Overview

HireFlow is a backend REST API that powers a job recruitment platform. It supports three user roles — **Admin**, **Recruiter**, and **Candidate** — each with distinct permissions enforced at the service layer.

**Key features:**

- JWT authentication with role-based access control
- Full application pipeline management (`applied → screening → interview → offer → hired/rejected`)
- Clean Architecture with strict layer separation
- Soft delete for users, companies, and jobs
- Structured JSON logging with zerolog
- Graceful server shutdown
- Input validation with human-readable error messages
- PostgreSQL via Docker Compose

---

## Tech Stack

| Layer            | Technology                  |
| ---------------- | --------------------------- |
| Language         | Go 1.22+                    |
| HTTP Framework   | Gin                         |
| ORM              | GORM                        |
| Database         | PostgreSQL 16               |
| Authentication   | JWT (golang-jwt/jwt v5)     |
| Password Hashing | bcrypt                      |
| Validation       | go-playground/validator v10 |
| Logging          | zerolog                     |
| Environment      | godotenv                    |
| Testing          | testify                     |
| Containerisation | Docker + Docker Compose     |

---

## Architecture

HireFlow follows **Clean Architecture** with strict unidirectional dependency flow:

```
HTTP Request
     │
     ▼
Middleware          (JWT validation, role check, logger, CORS)
     │
     ▼
Handler             (parse request, validate input, call service)
     │
     ▼
Service             (business rules, orchestration — no HTTP, no SQL)
     │
     ▼
Repository          (CRUD queries only — no business logic)
     │
     ▼
PostgreSQL
```

**Key principles enforced:**

- Handlers depend on service **interfaces**, not concrete types
- Services depend on repository **interfaces**, not concrete types
- Only `main.go` knows about concrete implementations
- DTOs cross the HTTP boundary; Entities cross the database boundary
- Mappers convert between the two — `PasswordHash` never leaves the server

---

## Project Structure

```
hireflow/
├── cmd/
│   └── api/
│       └── main.go                  # Entry point
│
├── internal/
│   ├── config/
│   │   ├── config.go                # Env config loader
│   │   └── logger.go                # Zerolog initialiser
│   │
│   ├── database/
│   │   ├── postgres.go              # GORM connection + migration
│   │   └── seed.go                  # Admin user seeder
│   │
│   ├── domain/
│   │   ├── entity/                  # Database models (GORM structs)
│   │   ├── dto/                     # Request/response shapes
│   │   └── mapper/                  # Entity ↔ DTO converters
│   │
│   ├── repository/                  # Data access layer (GORM queries)
│   ├── service/                     # Business logic layer
│   ├── handler/                     # HTTP handlers (controllers)
│   ├── middleware/                  # Auth, logger, CORS middleware
│   └── router/                      # Route definitions
│
├── internal/mocks/                  # Testify mock repositories
├── internal/service/tests/          # Service layer unit tests
│
├── http/                            # .http test files
│   ├── auth.http
│   ├── companies.http
│   ├── jobs.http
│   └── applications.http
│
├── .env                             # Local environment variables (never commit)
├── .env.example                     # Safe env template
├── docker-compose.yml               # PostgreSQL container
├── go.mod
└── go.sum
```

---

## Database Design

### Entities

| Table          | Description                                        |
| -------------- | -------------------------------------------------- |
| `users`        | All platform users — admin, recruiter, candidate   |
| `companies`    | Employer organisations                             |
| `jobs`         | Job postings created by recruiters under a company |
| `applications` | A candidate's application to a job                 |

### Relationships

```
companies  ──< users        (one company has many recruiters)
companies  ──< jobs         (one company has many jobs)
users      ──< jobs         (one recruiter posts many jobs)
jobs       ──< applications (one job receives many applications)
users      ──< applications (one candidate submits many applications)
```

### Key Constraints

- `users.deleted_at` — soft delete (GORM `DeletedAt`)
- `companies.deleted_at` — soft delete
- `jobs.deleted_at` — soft delete
- `applications (job_id, candidate_id)` — unique constraint (no duplicate applications)
- `applications` — hard deleted on withdrawal (no audit trail needed)

---

## API Endpoints

All routes are prefixed with `/api/v1`.

### Auth

| Method | Route            | Description              | Auth   |
| ------ | ---------------- | ------------------------ | ------ |
| POST   | `/auth/register` | Register a new user      | Public |
| POST   | `/auth/login`    | Login and receive JWT    | Public |
| GET    | `/auth/me`       | Get current user profile | Any    |

### Companies

| Method | Route            | Description        | Auth      |
| ------ | ---------------- | ------------------ | --------- |
| GET    | `/companies`     | List all companies | Public    |
| GET    | `/companies/:id` | Get a company      | Public    |
| POST   | `/companies`     | Create a company   | Admin     |
| PUT    | `/companies/:id` | Update a company   | Recruiter |
| DELETE | `/companies/:id` | Delete a company   | Admin     |

### Jobs

| Method | Route                 | Description                | Auth                    |
| ------ | --------------------- | -------------------------- | ----------------------- |
| GET    | `/jobs`               | List all open jobs         | Public                  |
| GET    | `/jobs/:id`           | Get a job posting          | Public                  |
| POST   | `/companies/:id/jobs` | Post a job under a company | Recruiter               |
| PUT    | `/jobs/:id`           | Update a job               | Recruiter (poster only) |
| PATCH  | `/jobs/:id/close`     | Close a job                | Recruiter (poster only) |
| DELETE | `/jobs/:id`           | Delete a job               | Recruiter (poster only) |

### Applications

| Method | Route                      | Description                | Auth                                    |
| ------ | -------------------------- | -------------------------- | --------------------------------------- |
| POST   | `/jobs/:id/apply`          | Apply to a job             | Candidate                               |
| GET    | `/applications/my`         | My applications            | Candidate                               |
| GET    | `/jobs/:id/applications`   | All applications for a job | Recruiter (poster only)                 |
| GET    | `/applications/:id`        | Get a single application   | Candidate (own) / Recruiter (their job) |
| PATCH  | `/applications/:id/status` | Advance pipeline status    | Recruiter (poster only)                 |
| DELETE | `/applications/:id`        | Withdraw an application    | Candidate (own)                         |

---

## Roles & Permissions

| Action                     | Admin | Recruiter | Candidate |
| -------------------------- | :---: | :-------: | :-------: |
| Create company             |  ✅   |    ❌     |    ❌     |
| Delete company             |  ✅   |    ❌     |    ❌     |
| Update company             |  ❌   |    ✅     |    ❌     |
| Post a job                 |  ❌   |    ✅     |    ❌     |
| Close / delete own job     |  ❌   |    ✅     |    ❌     |
| Apply to a job             |  ❌   |    ❌     |    ✅     |
| View own applications      |  ❌   |    ❌     |    ✅     |
| View job's applications    |  ❌   |    ✅     |    ❌     |
| Advance application status |  ❌   |    ✅     |    ❌     |
| Withdraw application       |  ❌   |    ❌     |    ✅     |

> **Note:** Recruiter-level actions are additionally scoped — a recruiter can only manage jobs they personally posted and applications for those jobs.

---

## Getting Started

### Prerequisites

- Go 1.22+
- Docker and Docker Compose
- Git

### 1. Clone the repository

```bash
git clone https://github.com/gpslakshan/hireflow.git
cd hireflow
```

### 2. Set up environment variables

```bash
cp .env.example .env
```

Edit `.env` and fill in your values — especially `JWT_SECRET` and `ADMIN_PASSWORD`.

### 3. Start PostgreSQL

```bash
docker compose up -d
```

Verify it is running:

```bash
docker compose ps
```

### 4. Install dependencies

```bash
go mod download
```

### 5. Run the server

```bash
go run cmd/api/main.go
```

On first run you will see:

```
INF database connection established
INF database migration completed
INF default admin user created — email: admin@hireflow.com
INF server starting on port 8080
```

The API is now live at `http://localhost:8080`.

### 6. Login as admin

```http
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json

{
  "email": "admin@hireflow.com",
  "password": "<your ADMIN_PASSWORD from .env>"
}
```

---

## Environment Variables

| Variable           | Description                                | Default              |
| ------------------ | ------------------------------------------ | -------------------- |
| `APP_PORT`         | Server port                                | `8080`               |
| `APP_ENV`          | Environment (`development` / `production`) | `development`        |
| `DB_HOST`          | PostgreSQL host                            | `localhost`          |
| `DB_PORT`          | PostgreSQL port                            | `5432`               |
| `DB_USER`          | Database user                              | —                    |
| `DB_PASSWORD`      | Database password                          | —                    |
| `DB_NAME`          | Database name                              | `hireflow_db`        |
| `DB_SSLMODE`       | SSL mode                                   | `disable`            |
| `JWT_SECRET`       | Secret key for signing JWTs                | —                    |
| `JWT_EXPIRY_HOURS` | Token expiry in hours                      | `72`                 |
| `ADMIN_EMAIL`      | Seeded admin email                         | `admin@hireflow.com` |
| `ADMIN_PASSWORD`   | Seeded admin password                      | —                    |

> **Never commit your `.env` file.** It is listed in `.gitignore`. Use `.env.example` as the committed template.

---

## Running Tests

Unit tests cover the service layer business rules. No database or running server required.

```bash
go test ./internal/service/tests/... -v
```

### What is tested

| Test                                     | Rule verified                                   |
| ---------------------------------------- | ----------------------------------------------- |
| `TestApply_Success`                      | Application created with `applied` status       |
| `TestApply_JobNotFound`                  | Cannot apply to a non-existent job              |
| `TestApply_JobClosed`                    | Cannot apply to a closed job                    |
| `TestApply_AlreadyApplied`               | Cannot apply to the same job twice              |
| `TestUpdateStatus_Success`               | Recruiter can advance pipeline stage            |
| `TestUpdateStatus_UnauthorizedRecruiter` | Only the job poster can update status           |
| `TestUpdateStatus_ApplicationNotFound`   | Returns error for missing application           |
| `TestWithdraw_Success`                   | Candidate can withdraw their own application    |
| `TestWithdraw_NotOwner`                  | Cannot withdraw another candidate's application |
| `TestClose_Success`                      | Recruiter can close their open job              |
| `TestClose_AlreadyClosed`                | Cannot close an already closed job              |
| `TestClose_UnauthorizedRecruiter`        | Only the poster can close a job                 |
| `TestClose_JobNotFound`                  | Returns error for missing job                   |
| `TestDeleteJob_Success`                  | Recruiter can delete their own job              |
| `TestDeleteJob_Unauthorized`             | Cannot delete another recruiter's job           |
| `TestUpdateJob_Success`                  | Job fields update correctly                     |

---

## Application Pipeline

When a candidate applies, the application enters a linear pipeline that only a recruiter can advance:

```
applied → screening → interview → offer → hired
                                        ↘ rejected
```

- A recruiter can move to any forward stage or set `rejected` at any point
- The `applied` status is set exclusively by the system on submission — it cannot be set via the API
- Closing a job does not affect existing applications

---

## Stopping the Server

Press `Ctrl+C`. The server will finish any in-flight requests before shutting down cleanly:

```
INF shutting down server...
INF server stopped cleanly
```

To stop the database container:

```bash
docker compose down
```

To stop and remove all data volumes:

```bash
docker compose down -v
```
