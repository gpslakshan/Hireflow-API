# HireFlow API

A production-grade **Job Recruitment REST API** built with Go, following Clean Architecture principles. Companies post jobs, candidates apply with CVs, and recruiters manage hiring pipelines — all secured with JWT-based role-based access control.

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
- [CV Upload Flow](#cv-upload-flow)
- [Running Tests](#running-tests)
- [Application Pipeline](#application-pipeline)

---

## Overview

HireFlow is a backend REST API that powers a job recruitment platform. It supports three user roles — **Admin**, **Recruiter**, and **Candidate** — each with distinct permissions enforced at the service layer.

**Key features:**

- JWT authentication with role-based access control
- Full application pipeline management (`applied → screening → interview → offer → hired/rejected`)
- CV upload via AWS S3 pre-signed URLs — server never touches binary files
- Pre-signed download URLs generated fresh on every application fetch
- Automatic CV cleanup from S3 when an application is withdrawn
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
| File Storage     | AWS S3 (pre-signed URLs)    |
| AWS SDK          | aws-sdk-go-v2               |
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
- Services depend on storage **interfaces**, not concrete AWS types
- Only `main.go` knows about concrete implementations
- DTOs cross the HTTP boundary; Entities cross the database boundary
- Mappers convert between the two — `PasswordHash` never leaves the server
- S3 keys are stored in the database — never the full URL (URLs expire, keys don't)

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
│   ├── router/                      # Route definitions
│   └── storage/
│       ├── interfaces.go            # CVStorage interface
│       └── s3.go                    # AWS S3 pre-signed URL implementation
│
├── internal/mocks/                  # Testify mock repositories + storage
├── internal/service/tests/          # Service layer unit tests
│
├── http/                            # .http test files
│   ├── auth.http
│   ├── companies.http
│   ├── jobs.http
│   └── applications.http            # Includes CV upload flow tests
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

| Table          | Description                                                         |
| -------------- | ------------------------------------------------------------------- |
| `users`        | All platform users — admin, recruiter, candidate                    |
| `companies`    | Employer organisations                                              |
| `jobs`         | Job postings created by recruiters under a company                  |
| `applications` | A candidate's application to a job, including optional CV reference |

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
- `applications.cv_key` — stores S3 object key, not the URL (URLs expire, keys don't)
- `applications` — hard deleted on withdrawal, CV cleaned up from S3 automatically

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

| Method | Route                      | Description                       | Auth                                    |
| ------ | -------------------------- | --------------------------------- | --------------------------------------- |
| POST   | `/jobs/:id/apply`          | Apply to a job (with optional CV) | Candidate                               |
| GET    | `/applications/my`         | My applications                   | Candidate                               |
| GET    | `/jobs/:id/applications`   | All applications for a job        | Recruiter (poster only)                 |
| GET    | `/applications/:id`        | Get a single application          | Candidate (own) / Recruiter (their job) |
| PATCH  | `/applications/:id/status` | Advance pipeline status           | Recruiter (poster only)                 |
| DELETE | `/applications/:id`        | Withdraw an application           | Candidate (own)                         |

### Uploads

| Method | Route                    | Description                           | Auth      |
| ------ | ------------------------ | ------------------------------------- | --------- |
| POST   | `/uploads/cv-upload-url` | Get a pre-signed S3 URL for CV upload | Candidate |

---

## Roles & Permissions

| Action                     | Admin | Recruiter | Candidate |
| -------------------------- | :---: | :-------: | :-------: |
| Create company             |  ✅   |    ❌     |    ❌     |
| Delete company             |  ✅   |    ❌     |    ❌     |
| Update company             |  ❌   |    ✅     |    ❌     |
| Post a job                 |  ❌   |    ✅     |    ❌     |
| Close / delete own job     |  ❌   |    ✅     |    ❌     |
| Get CV upload URL          |  ❌   |    ❌     |    ✅     |
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
- AWS account with an S3 bucket (for CV upload feature)

### 1. Clone the repository

```bash
git clone https://github.com/yourusername/hireflow.git
cd hireflow
```

### 2. Set up environment variables

```bash
cp .env.example .env
```

Edit `.env` and fill in your values — especially `JWT_SECRET`, `ADMIN_PASSWORD`, and all `AWS_*` variables.

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

| Variable                         | Description                                | Default              |
| -------------------------------- | ------------------------------------------ | -------------------- |
| `APP_PORT`                       | Server port                                | `8080`               |
| `APP_ENV`                        | Environment (`development` / `production`) | `development`        |
| `DB_HOST`                        | PostgreSQL host                            | `localhost`          |
| `DB_PORT`                        | PostgreSQL port                            | `5432`               |
| `DB_USER`                        | Database user                              | —                    |
| `DB_PASSWORD`                    | Database password                          | —                    |
| `DB_NAME`                        | Database name                              | `hireflow_db`        |
| `DB_SSLMODE`                     | SSL mode                                   | `disable`            |
| `JWT_SECRET`                     | Secret key for signing JWTs                | —                    |
| `JWT_EXPIRY_HOURS`               | Token expiry in hours                      | `72`                 |
| `ADMIN_EMAIL`                    | Seeded admin email                         | `admin@hireflow.com` |
| `ADMIN_PASSWORD`                 | Seeded admin password                      | —                    |
| `AWS_ACCESS_KEY_ID`              | IAM user access key                        | —                    |
| `AWS_SECRET_ACCESS_KEY`          | IAM user secret key                        | —                    |
| `AWS_REGION`                     | S3 bucket region                           | `ap-south-1`         |
| `AWS_S3_BUCKET`                  | S3 bucket name                             | —                    |
| `CV_UPLOAD_URL_EXPIRY_MINUTES`   | Pre-signed upload URL expiry               | `15`                 |
| `CV_DOWNLOAD_URL_EXPIRY_MINUTES` | Pre-signed download URL expiry             | `15`                 |

> **Never commit your `.env` file.** It is listed in `.gitignore`. Use `.env.example` as the committed template.

---

## CV Upload Flow

HireFlow uses **AWS S3 pre-signed URLs** for CV file handling. Your server never touches the binary file — uploads and downloads go directly between the client and S3.

### AWS Setup

**Create an S3 bucket:**

1. Go to AWS Console → S3 → Create bucket
2. Choose a globally unique name e.g. `hireflow-cvs-yourname`
3. Select your region
4. Keep **Block all public access** enabled — files are accessed only via pre-signed URLs

**Create an IAM user:**

1. Go to IAM → Users → Create user
2. Attach a policy allowing `s3:PutObject`, `s3:GetObject`, `s3:DeleteObject` on your bucket ARN
3. Create an access key and copy the credentials into your `.env`

### Three-Step Upload Flow

```
Step 1 — Candidate requests a pre-signed upload URL
         POST /api/v1/uploads/cv-upload-url
         → receives { upload_url, cv_key }

Step 2 — Candidate uploads PDF directly to S3
         PUT <upload_url>  (no Authorization header — auth is in the URL)
         Content-Type: application/pdf
         Body: binary PDF file
         → S3 stores the file, server not involved

Step 3 — Candidate submits application with cv_key
         POST /api/v1/jobs/:id/apply
         Body: { "cover_letter": "...", "cv_key": "cvs/..." }
         → cv_key stored in the database
```

### Download Flow

```
Recruiter fetches application
         GET /api/v1/applications/:id
         → service generates a fresh pre-signed GET URL from the stored cv_key
         → response includes cv_download_url (valid for 15 minutes)

Recruiter opens cv_download_url in browser or Postman
         → PDF downloads directly from S3
```

### Design Decisions

**Why store the key and not the URL?** Pre-signed URLs expire after 15 minutes. Storing the key (`cvs/candidate-id/filename.pdf`) is permanent — a fresh download URL is generated on every application fetch.

**Why pre-signed URLs instead of server-side upload?** The server never handles binary data — no memory pressure, no upload timeouts, no multipart parsing. S3 handles the upload directly and scales to thousands of concurrent uploads.

**Why PDF only?** The `GenerateUploadURL` call sets `ContentType: application/pdf` in the S3 signature. Uploading with any other content type returns `403 SignatureDoesNotMatch` from S3, enforcing the file type constraint at the infrastructure level.

**What happens when an application is withdrawn?** The service calls `s3.DeleteObject(cv_key)` before deleting the application row. S3 cleanup is best-effort — if it fails, the withdrawal still succeeds.

---

## Running Tests

Unit tests cover the service layer business rules. No database, no AWS credentials, and no running server required — all external dependencies are mocked.

```bash
go test ./internal/service/tests/... -v
```

### Application Service Tests

| Test                                             | Rule verified                                            |
| ------------------------------------------------ | -------------------------------------------------------- |
| `TestApply_Success`                              | Application created with `applied` status                |
| `TestApply_WithCVKey`                            | CV key stored on application, S3 not called during apply |
| `TestApply_JobNotFound`                          | Cannot apply to a non-existent job                       |
| `TestApply_JobClosed`                            | Cannot apply to a closed job                             |
| `TestApply_AlreadyApplied`                       | Cannot apply to the same job twice                       |
| `TestGetByID_WithCV_GeneratesDownloadURL`        | S3 download URL generated when CV key exists             |
| `TestGetByID_WithoutCV_NoDownloadURL`            | S3 not called when no CV was uploaded                    |
| `TestGetByID_CandidateCannotSeeOtherApplication` | Candidates only see their own applications               |
| `TestUpdateStatus_Success`                       | Recruiter can advance pipeline stage                     |
| `TestUpdateStatus_UnauthorizedRecruiter`         | Only the job poster can update status                    |
| `TestUpdateStatus_ApplicationNotFound`           | Returns error for missing application                    |
| `TestWithdraw_Success_WithCV`                    | S3 DeleteObject called when CV exists on withdrawal      |
| `TestWithdraw_Success_WithoutCV`                 | S3 not called when no CV was uploaded                    |
| `TestWithdraw_NotOwner`                          | Cannot withdraw another candidate's application          |

### Job Service Tests

| Test                              | Rule verified                         |
| --------------------------------- | ------------------------------------- |
| `TestClose_Success`               | Recruiter can close their open job    |
| `TestClose_AlreadyClosed`         | Cannot close an already closed job    |
| `TestClose_UnauthorizedRecruiter` | Only the poster can close a job       |
| `TestClose_JobNotFound`           | Returns error for missing job         |
| `TestDeleteJob_Success`           | Recruiter can delete their own job    |
| `TestDeleteJob_Unauthorized`      | Cannot delete another recruiter's job |
| `TestUpdateJob_Success`           | Job fields update correctly           |

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
- A candidate can withdraw their application at any stage

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
