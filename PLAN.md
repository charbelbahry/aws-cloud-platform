```markdown

---
description: "Implementation plan for AWS Cloud Platform"
created-date: 2026-08-02
status: not-started
target-repo: https://github.com/charbelbahry/aws-cloud-platform
---

# AWS Cloud Platform - Implementation Plan

## Project Goal

Build a Go-based cloud deployment platform that demonstrates proficiency in
AWS infrastructure, Terraform IaC, Docker containerization, GitHub Actions
CI/CD, and production monitoring. The application is a simple CRUD API. The
infrastructure is the learning objective.

## Target Outcome

A public GitHub repository containing a fully deployed, monitored,
CI/CD-automated Go application on AWS, provisioned entirely through Terraform,
with documentation and architecture diagrams. This repository will be linked
on a CV for a Cloud/DevOps Engineering internship application at Digico
Solutions (Beirut).

## Architecture Overview

```

Developer
    |
    | git push
    v
GitHub Repository
    |
    | GitHub Actions
    v
+-----------------------------------+
| CI Pipeline (every PR)            |
|  - go test -race                  |
|  - golangci-lint run              |
|  - go build                       |
+-----------------------------------+
    |
    | (on merge to main)
    v
+-----------------------------------+
| CD Pipeline                       |
|  - docker build (multi-stage)     |
|  - docker push -> Amazon ECR      |
|  - terraform plan                 |
|  - terraform apply                |
|  - deploy to EC2                  |
|  - curl /health -> 200?           |
|      yes -> done                  |
|      no  -> rollback              |
+-----------------------------------+
    |
    v
+-----------------------------------+
| AWS Infrastructure (Terraform)    |
|                                   |
|  VPC (10.0.0.0/16)                |
|  +-- Public Subnet (10.0.1.0/24)  |
|  |   +-- EC2 Instance             |
|  |       +-- Docker               |
|  |           +-- Go API (:8080)   |
|  +-- Private Subnet (10.0.2.0/24) |
|  |   +-- RDS PostgreSQL (:5432)   |
|  +-- Internet Gateway             |
|  +-- Security Groups              |
|                                   |
|  S3 Bucket (Terraform state)      |
|  DynamoDB Table (state lock)      |
|  ECR Repository (Docker images)   |
|  IAM Role + Policy (EC2)          |
|  CloudWatch (Logs + Alarms)       |
+-----------------------------------+

```

## Database Schema

```sql
-- migrations/001_create_services.sql
CREATE TABLE IF NOT EXISTS services (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL UNIQUE,
    repository  TEXT,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- migrations/002_create_deployments.sql
CREATE TABLE IF NOT EXISTS deployments (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id   UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    version      TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'pending'
                 CHECK (status IN (
                     'pending', 'building', 'running', 'failed', 'rolled_back'
                 )),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_deployments_service_id
    ON deployments(service_id);
```

## API Endpoints

| Method | Path                       | Description                     | Success Code |
|--------|----------------------------|---------------------------------|--------------|
| GET    | /health                    | Liveness check (always 200)     | 200          |
| GET    | /ready                     | Readiness check (DB ping)       | 200 / 503    |
| GET    | /metrics                   | Basic JSON metrics              | 200          |
| POST   | /services                  | Create a service                | 201          |
| GET    | /services                  | List all services               | 200          |
| GET    | /services/{id}             | Get one service by ID           | 200 / 404    |
| PUT    | /services/{id}             | Update a service                | 200 / 404    |
| DELETE | /services/{id}             | Delete a service                | 204 / 404    |
| POST   | /services/{id}/deployments | Create a deployment record      | 201 / 404    |
| GET    | /services/{id}/deployments | List deployments for a service  | 200 / 404    |

---

## Phase 1: Go API Foundation

**Goal:** A working Go HTTP server with PostgreSQL, running locally.
**Estimated time:** 3-5 days
**New concepts:** Go syntax, net/http, database/sql, slog, context, structs,
methods, interfaces, error handling, goroutines, graceful shutdown, embed.

### Prerequisites

- [ ] Install Go 1.26+ (verify with `go version`)
- [ ] Install golangci-lint (verify with `golangci-lint --version`)
- [ ] Install Docker Desktop (verify with `docker --version` and
      `docker compose version`)
- [ ] Create GitHub repository: `aws-cloud-platform` (public)
- [ ] Clone locally: `git clone git@github.com:charbelbahry/aws-cloud-platform.git`
- [ ] Create and checkout branch: `git checkout -b phase-1-go-api`

### Steps

- [ ] **Step 1.1: Project initialization**
  - **Objective:** Initialize the Go module, create the directory structure,
    and set up developer tooling.
  - **Tasks:**
    - Run `go mod init github.com/charbelbahry/aws-cloud-platform`
    - Create all directories from the project structure
    - Create `Makefile` with targets: `run`, `test`, `lint`, `build`,
      `docker-up`, `docker-down`
    - Create `.gitignore` (bin/, .env, *.tfstate,*.tfstate.backup,
      .terraform/, vendor/, *.tfvars except .tfvars.example)
    - Create `.env.example` with PORT, DATABASE_URL, LOG_LEVEL
  - **Files:** `go.mod`, `Makefile`, `.gitignore`, `.env.example`
  - **Learn:** Go modules, `go mod init`, Makefile syntax, `.gitignore`
    patterns.
  - **Validation:** `go mod init` succeeds. `make run` prints a placeholder
    message. Directory structure matches the plan.
  - **Commit:** `feat: initialize Go project structure and developer tooling`

- [ ] **Step 1.2: Configuration loading**
  - **Objective:** Load application configuration from environment variables
    with sensible defaults and validation.
  - **Tasks:**
    - Create `internal/config/config.go`
    - Define a `Config` struct with fields: `Port`, `DatabaseURL`, `LogLevel`
    - Load from `os.Getenv` with defaults (Port=8080, LogLevel=info)
    - Return error if `DATABASE_URL` is empty
    - Add a `Load() (*Config, error)` function
  - **Files:** `internal/config/config.go`
  - **Learn:** Go structs, `os.Getenv`, default value patterns, explicit
    error returns, why Go doesn't have constructors.
  - **Understanding checkpoint:** "Why do we return an error if DATABASE_URL
    is missing instead of defaulting to an empty string?"
  - **Commit:** `feat: add configuration loading from environment variables`

- [ ] **Step 1.3: Database connection pool**
  - **Objective:** Create a PostgreSQL connection pool with health checking
    and graceful cleanup.
  - **Tasks:**
    - Create `internal/database/db.go`
    - Open connection with `sql.Open("postgres", dsn)`
    - Configure pool: `SetMaxOpenConns(25)`, `SetMaxIdleConns(5)`,
      `SetConnMaxLifetime(5 * time.Minute)`
    - Implement `Ping(ctx)` for health checks
    - Implement `Close()` for shutdown
    - Add `github.com/lib/pq` dependency: `go get github.com/lib/pq`
  - **Files:** `internal/database/db.go`, `go.mod`, `go.sum`
  - **Learn:** Connection pooling, `sql.Open` vs `db.Ping`, why pooling
    matters, `defer` for cleanup, context propagation.
  - **Understanding checkpoint:** "What is the difference between `sql.Open()`
    and `db.Ping()`? Why is calling `sql.Open()` alone not enough to confirm
    the database is reachable?"
  - **Commit:** `feat: add PostgreSQL connection pool with health check`

- [ ] **Step 1.4: Database migrations**
  - **Objective:** Create SQL migration files and an embedded migration runner
    that applies them on startup.
  - **Tasks:**
    - Create `migrations/001_create_services.sql`
    - Create `migrations/002_create_deployments.sql`
    - Create `internal/database/migrations.go`
    - Use `//go:embed migrations/*.sql` to embed SQL files
    - Create a `schema_migrations` tracking table
    - Run unapplied migrations in order on startup
    - Make migrations idempotent with `IF NOT EXISTS`
  - **Files:** `migrations/001_create_services.sql`,
    `migrations/002_create_deployments.sql`,
    `internal/database/migrations.go`
  - **Learn:** `embed.FS`, SQL DDL, migration patterns, idempotent DDL,
    why migrations must be ordered and tracked.
  - **Understanding checkpoint:** "Why do we embed the SQL files into the
    binary using `//go:embed` instead of reading them from disk at runtime?"
  - **Commit:** `feat: add embedded SQL migration runner`

- [ ] **Step 1.5: Models and validation**
  - **Objective:** Define data structures for services and deployments with
    JSON serialization and input validation.
  - **Tasks:**
    - Create `internal/models/service.go` with `Service` struct, JSON tags,
      and `Validate() error` method
    - Create `internal/models/deployment.go` with `Deployment` struct,
      JSON tags, and `Validate() error` method
    - Use `time.Time` for timestamps with proper JSON marshaling
    - Validate: name is required, status is one of the allowed values
  - **Files:** `internal/models/service.go`, `internal/models/deployment.go`
  - **Learn:** Go structs, JSON struct tags, methods with pointer receivers,
    validation patterns without a framework, `time.Time` JSON handling.
  - **Understanding checkpoint:** "Why does `Validate()` use a pointer
    receiver `(s *Service)` instead of a value receiver `(s Service)`?
    What would change if we used a value receiver?"
  - **Commit:** `feat: add Service and Deployment models with validation`

- [ ] **Step 1.6: Services CRUD handlers**
  - **Objective:** Implement HTTP handlers for all service endpoints using
    only the standard library.
  - **Tasks:**
    - Create `internal/handlers/services.go`
    - Define a `ServiceHandler` struct holding a `*sql.DB` reference
    - Implement: `CreateService`, `ListServices`, `GetService`,
      `UpdateService`, `DeleteService`
    - Use `http.NewServeMux` with Go 1.22+ patterns:
      `mux.HandleFunc("POST /services", h.Create)`
      `mux.HandleFunc("GET /services/{id}", h.GetByID)`
    - Parse path params with `r.PathValue("id")`
    - Return proper status codes: 201, 200, 204, 400, 404, 500
    - Set `Content-Type: application/json` header BEFORE `WriteHeader`
  - **Files:** `internal/handlers/services.go`
  - **Learn:** `net/http` handler signature, `http.ResponseWriter`,
    `r.PathValue()`, JSON encode/decode, HTTP status code semantics,
    REST conventions, why header order matters.
  - **Understanding checkpoint:** "Why must we call
    `w.Header().Set(\"Content-Type\", \"application/json\")` BEFORE
    `w.WriteHeader(statusCode)`? What happens if we reverse the order?"
  - **Commit:** `feat: add services CRUD HTTP handlers`

- [ ] **Step 1.7: Deployments handlers**
  - **Objective:** Implement HTTP handlers for deployment records nested
    under services.
  - **Tasks:**
    - Create `internal/handlers/deployments.go`
    - Define a `DeploymentHandler` struct holding a `*sql.DB` reference
    - Implement: `CreateDeployment`, `ListDeployments`
    - Verify parent service exists before creating a deployment (return
      404 if not)
    - Use SQL JOIN for listing deployments with service context
  - **Files:** `internal/handlers/deployments.go`
  - **Learn:** Nested resource patterns, foreign key validation at the
    application layer, SQL parameterized queries, preventing SQL injection.
  - **Understanding checkpoint:** "Why do we check that the service exists
    before inserting the deployment, even though the database has a FOREIGN
    KEY constraint? Isn't the constraint enough?"
  - **Commit:** `feat: add deployments HTTP handlers`

- [ ] **Step 1.8: Health and metrics endpoints**
  - **Objective:** Add operational endpoints for liveness, readiness, and
    basic application metrics.
  - **Tasks:**
    - Create `internal/handlers/health.go`
    - `/health`: always returns `200 {"status": "ok"}`
    - `/ready`: pings the database, returns `200` or `503`
    - Create `internal/handlers/metrics.go`
    - `/metrics`: returns JSON with `uptime_seconds`, `goroutine_count`
      (via `runtime.NumGoroutine()`), `go_version`
  - **Files:** `internal/handlers/health.go`, `internal/handlers/metrics.go`
  - **Learn:** Liveness vs readiness probes, `runtime` package, why
    orchestrators need both probe types, what makes a service "ready"
    vs "alive."
  - **Understanding checkpoint:** "What is the difference between a liveness
    probe and a readiness probe? Describe a scenario where a service is
    alive but not ready."
  - **Commit:** `feat: add health, readiness, and metrics endpoints`

- [ ] **Step 1.9: Middleware**
  - **Objective:** Add request logging and panic recovery as composable
    HTTP middleware.
  - **Tasks:**
    - Create `internal/middleware/logging.go`
    - Log: method, path, status code, duration, request ID using `slog`
    - Use a `responseWriter` wrapper to capture the status code
    - Create `internal/middleware/recovery.go`
    - Recover from panics, log the stack trace with `slog.Error`,
      return `500 Internal Server Error`
    - Implement middleware as `func(http.Handler) http.Handler`
  - **Files:** `internal/middleware/logging.go`,
    `internal/middleware/recovery.go`
  - **Learn:** Middleware pattern in Go, `http.Handler` interface,
    decorator pattern, `defer` for panic recovery, `slog` structured
    logging with attributes, `runtime/debug.Stack()`.
  - **Understanding checkpoint:** "Draw the middleware chain for an incoming
    request: request -> ??? -> ??? -> handler -> response. Why does
    middleware take an `http.Handler` and return an `http.Handler`?"
  - **Commit:** `feat: add request logging and panic recovery middleware`

- [ ] **Step 1.10: Main entry point and graceful shutdown**
  - **Objective:** Wire all components together and implement graceful
    shutdown on SIGINT/SIGTERM.
  - **Tasks:**
    - Create `cmd/api/main.go`
    - Load config, connect to DB, run migrations
    - Create `http.NewServeMux`, register all routes
    - Wrap mux with recovery middleware, then logging middleware
    - Create `http.Server` with timeouts:
      `ReadTimeout: 5s`, `WriteTimeout: 10s`, `IdleTimeout: 120s`
    - Start server in a goroutine
    - Use `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)`
    - On signal: call `server.Shutdown(ctx)` with a 10-second timeout
    - Close database connection after server stops
    - Log startup and shutdown with `slog`
  - **Files:** `cmd/api/main.go`
  - **Learn:** Application wiring, `signal.NotifyContext`, `http.Server`
    vs `http.ListenAndServe`, graceful shutdown, context cancellation,
    goroutine for non-blocking server start, `sync` patterns.
  - **Understanding checkpoint:** "What happens to in-flight HTTP requests
    when we call `server.Shutdown(ctx)`? What happens if a request takes
    longer than the 10-second shutdown timeout?"
  - **Commit:** `feat: add main entry point with graceful shutdown`

- [ ] **Step 1.11: Tests**
  - **Objective:** Write table-driven tests for all handlers.
  - **Tasks:**
    - Create `internal/handlers/services_test.go`
    - Create `internal/handlers/health_test.go`
    - Use `httptest.NewRecorder` and `httptest.NewRequest`
    - Test cases: successful CRUD, validation error (400), not found (404),
      health returns 200, ready returns 503 when DB is down
    - Use table-driven pattern: `tests := []struct{ name string; ... }`
    - Run with `go test -v -race ./...`
  - **Files:** `internal/handlers/services_test.go`,
    `internal/handlers/health_test.go`
  - **Learn:** `testing` package, table-driven tests, `httptest`,
    `t.Run()` for subtests, `t.Helper()`, race detector, test isolation.
  - **Validation:** `go test -v -race -cover ./...` passes with 0 failures.
  - **Commit:** `test: add table-driven tests for handlers`

- [ ] **Step 1.12: Local PostgreSQL with docker compose**
  - **Objective:** Run the full stack locally with docker compose.
  - **Tasks:**
    - Create `docker-compose.yml` with PostgreSQL 16 service
    - Configure: volume for persistence, health check
      (`pg_isready`), environment variables
    - Add `depends_on` with `condition: service_healthy`
    - Create `.env.example` with `DATABASE_URL`
  - **Files:** `docker-compose.yml`, `.env.example`
  - **Learn:** Docker Compose networking (services reach each other by
    name), named volumes, health checks in compose, `depends_on` conditions.
  - **Validation:**
    - `docker compose up -d` starts PostgreSQL
    - `go run ./cmd/api` connects and runs migrations
    - `curl localhost:8080/health` returns `{"status":"ok"}`
    - `curl localhost:8080/ready` returns 200
    - CRUD operations work via curl
  - **Commit:** `feat: add docker-compose for local PostgreSQL`

### Phase 1 Completion Criteria

- [ ] `make run` starts the API on localhost:8080
- [ ] All CRUD endpoints work (verified with curl)
- [ ] `make test` passes with 0 failures and race detector enabled
- [ ] `make lint` passes with 0 warnings
- [ ] Graceful shutdown works (Ctrl+C exits cleanly with log message)
- [ ] All code committed on branch `phase-1-go-api`
- [ ] Branch merged to main

---

## Phase 2: Docker Containerization

**Goal:** The Go application runs as a production-ready Docker container.
**Estimated time:** 1 day
**New concepts:** Multi-stage builds, image layers, non-root users,
.dockerignore, HEALTHCHECK instruction, image size optimization.

### Steps

- [ ] **Step 2.1: Multi-stage Dockerfile**
  - **Objective:** Create a production Dockerfile that produces a minimal,
    secure container image.
  - **Tasks:**
    - Create `Dockerfile` with two stages:
      - Stage 1 (`builder`): `golang:1.26-alpine`, copy `go.mod`/`go.sum`
        first (layer caching), download deps, copy source, `CGO_ENABLED=0
        go build -o /api ./cmd/api`
      - Stage 2 (`runtime`): `alpine:3.24`, install `ca-certificates`,
        copy binary from builder, create non-root user, `USER 1000:1000`,
        `EXPOSE 8080`, `HEALTHCHECK`, `ENTRYPOINT ["/api"]`
    - Create `.dockerignore` (.git, bin/, terraform/, docs/, k8s/,
      *.md, .env)
  - **Files:** `Dockerfile`, `.dockerignore`
  - **Learn:** Multi-stage builds, Docker layer caching (why go.mod is
    copied before source), `CGO_ENABLED=0` for static binary, why alpine,
    why non-root, `HEALTHCHECK` instruction, `ENTRYPOINT` vs `CMD`.
  - **Understanding checkpoint:** "Why is the final image approximately
    15-20MB instead of approximately 800MB? What did we leave behind in
    the builder stage, and why doesn't it end up in the final image?"
  - **Commit:** `feat: add multi-stage Dockerfile with non-root user`

- [ ] **Step 2.2: Update docker-compose for full stack**
  - **Objective:** Run both the Go app and PostgreSQL as containers.
  - **Tasks:**
    - Update `docker-compose.yml` to build the Go app from the Dockerfile
    - Add `depends_on` with `condition: service_healthy` for PostgreSQL
    - Pass environment variables to the app container
    - Add a named volume for PostgreSQL data
  - **Files:** `docker-compose.yml`
  - **Learn:** Compose build context, service dependencies with health
    conditions, container-to-container networking, environment variable
    passing.
  - **Validation:**
    - `docker compose up --build` starts both containers
    - `curl localhost:8080/health` returns 200
    - `docker images aws-cloud-platform` shows image < 30MB
    - `docker exec <container> whoami` returns non-root user
    - `docker inspect <container>` shows HEALTHCHECK as healthy
  - **Commit:** `feat: containerize Go app in docker-compose full stack`

### Phase 2 Completion Criteria

- [ ] `docker compose up --build` starts the full stack
- [ ] Final image is under 30MB
- [ ] Container runs as non-root
- [ ] HEALTHCHECK passes
- [ ] All committed on branch `phase-2-docker`, merged to main

---

## Phase 3: GitHub Actions CI

**Goal:** Every pull request is automatically tested, linted, and built.
**Estimated time:** 1 day
**New concepts:** GitHub Actions workflows, jobs, steps, service containers,
caching, concurrency.

### Steps

- [ ] **Step 3.1: golangci-lint configuration**
  - **Objective:** Configure the Go linter for the project.
  - **Tasks:**
    - Create `.golangci.yml`
    - Enable linters: `govet`, `errcheck`, `staticcheck`, `unused`,
      `gosimple`, `ineffassign`, `typecheck`, `goconst`, `gofmt`
    - Exclude test files from some checks
    - Set `timeout: 5m`
  - **Files:** `.golangci.yml`
  - **Learn:** Go linting ecosystem, static analysis, common Go mistakes
    caught by linters, why linting matters in CI.
  - **Validation:** `golangci-lint run` passes with 0 warnings.
  - **Commit:** `ci: add golangci-lint configuration`

- [ ] **Step 3.2: CI workflow**
  - **Objective:** Create a GitHub Actions workflow that runs on every PR.
  - **Tasks:**
    - Create `.github/workflows/ci.yml`
    - Trigger: `pull_request` and `push` to non-main branches
    - Job 1 (`lint`): checkout, setup Go, install golangci-lint, run lint
    - Job 2 (`test`): checkout, setup Go, start PostgreSQL service
      container, run `go test -v -race -cover ./...`
    - Job 3 (`build`): checkout, setup Go, run `go build ./...`,
      run `docker build`
    - Use `actions/checkout@v4` and `actions/setup-go@v5`
    - Add Go module caching
  - **Files:** `.github/workflows/ci.yml`
  - **Learn:** GitHub Actions YAML syntax, `on` triggers, `jobs`, `steps`,
    `services` (sidecar containers for PostgreSQL), action versions,
    Go module caching, job independence.
  - **Understanding checkpoint:** "Why do we run tests with the `-race`
    flag? What category of bugs does the race detector catch, and why
    are they especially dangerous in a concurrent language like Go?"
  - **Validation:** Push a branch, open a PR, all three jobs pass (green).
  - **Commit:** `ci: add GitHub Actions workflow for lint, test, and build`

### Phase 3 Completion Criteria

- [ ] PR triggers CI pipeline automatically
- [ ] All three jobs (lint, test, build) pass
- [ ] Race detector is enabled in test job
- [ ] PostgreSQL service container is used for tests
- [ ] All committed on branch `phase-3-ci`, merged to main

---

## Phase 4: AWS Manual Exploration

**Goal:** Understand AWS services by creating them manually in the console
before automating with Terraform.
**Estimated time:** 3-4 days
**New concepts:** AWS console, VPC networking, subnets, internet gateways,
NAT gateways, security groups, EC2, RDS, IAM, ECR, S3, CloudWatch.

### IMPORTANT

This phase is about LEARNING, not building. You will create resources
manually, understand them, then DESTROY them all. Terraform (Phase 5)
will recreate everything as code. No code commits for this phase.

### Steps

- [ ] **Step 4.1: AWS account setup**
  - **Objective:** Set up a secure AWS account with billing protection.
  - **Tasks:**
    - Create AWS free tier account
    - Enable MFA on root account
    - Create IAM user for programmatic access (NOT root)
    - Set up CloudWatch billing alarm at $5
    - Install AWS CLI v2
    - Configure a named profile: `aws configure --profile clouddev`
  - **Learn:** AWS account structure, root vs IAM user, MFA, billing
    alerts, AWS CLI profiles, free tier limits.
  - **Understanding checkpoint:** "Why should you never use the root
    account for daily work? What could happen if root credentials are
    compromised?"

- [ ] **Step 4.2: VPC and networking**
  - **Objective:** Build a private network in AWS manually optimized for the AWS Free Tier.
  - **Tasks:**
    - Create VPC: `10.0.0.0/16`, name: `clouddev-vpc`
    - Create public subnet: `10.0.1.0/24`, auto-assign public IP
    - Create private subnet: `10.0.2.0/24`
    - Create internet gateway, attach to VPC
    - Create public route table: route `0.0.0.0/0` -> IGW
    - Create private route table (VPC local traffic only; NAT Gateway omitted to guarantee 100% Free Tier compliance)
    - Associate subnets with route tables
  - **Learn:** CIDR notation, public vs private subnets, internet gateway,
    route tables, network ACLs (conceptual), why databases go in private subnets,
    and AWS Free Tier cost trade-offs.
  - **Understanding checkpoint:** "Draw the network diagram from memory.
    Why can the EC2 instance in the public subnet reach the internet via IGW,
    while the RDS instance in the private subnet remains isolated? Why is a NAT
    Gateway unnecessary for RDS in standard usage, saving ~$32/month?"

- [ ] **Step 4.3: Security groups**
  - **Objective:** Create firewall rules for the application and database.
  - **Tasks:**
    - Create `app-sg`: inbound HTTP (80) from `0.0.0.0/0`, inbound
      custom (8080) from `0.0.0.0/0`, all outbound allowed
    - Create `db-sg`: inbound PostgreSQL (5432) from `app-sg` ONLY,
      all outbound allowed
  - **Learn:** Security groups as stateful firewalls, SG-to-SG references,
    why `0.0.0.0/0` on port 5432 is dangerous, stateful vs stateless
    (security groups vs NACLs).
  - **Understanding checkpoint:** "What would happen if you set the db-sg
    inbound rule to allow port 5432 from `0.0.0.0/0` instead of from
    `app-sg`? What attack does the SG-to-SG rule prevent?"

- [ ] **Step 4.4: RDS PostgreSQL**
  - **Objective:** Create a managed PostgreSQL database in the private subnet.
  - **Tasks:**
    - Create RDS PostgreSQL instance: `db.t3.micro`, PostgreSQL 16,
      free tier eligible
    - Place in private subnet via subnet group
    - Attach `db-sg` security group
    - Set master username and password (store securely, NOT in git)
    - Enable automated backups (1 day retention)
    - Disable public access
  - **Learn:** Managed vs self-managed databases, RDS instance classes,
    subnet groups, automated backups, multi-AZ (conceptual), why
    "disable public access" matters.
  - **Understanding checkpoint:** "Why is RDS in a private subnet with
    public access disabled? How does the application reach it if it has
    no public IP address?"

- [ ] **Step 4.5: EC2 instance**
  - **Objective:** Launch a virtual machine and deploy a container manually.
  - **Tasks:**
    - Launch EC2: `t3.micro`, Amazon Linux 2024, public subnet, `app-sg`
    - Create and download a key pair
    - SSH into the instance
    - Install Docker: `sudo yum install -y docker && sudo systemctl start docker`
    - Run `docker run hello-world` to verify
    - Test connectivity to RDS endpoint: `nc -zv <rds-endpoint> 5432`
  - **Learn:** EC2 instance types, AMIs, key pairs, SSH, security group
    enforcement, installing Docker on Linux, container networking from EC2,
    `user_data` scripts (conceptual).
  - **Understanding checkpoint:** "What is the difference between a key
    pair and a security group? You have both. Why do you need both, and
    what does each one protect against?"

- [ ] **Step 4.6: ECR repository**
  - **Objective:** Push and pull Docker images through AWS's container registry.
  - **Tasks:**
    - Create ECR repository: `aws-cloud-platform`
    - Authenticate Docker locally:
      `aws ecr get-login-password --profile clouddev | docker login --username AWS --password-stdin <account>.dkr.ecr.<region>.amazonaws.com`
    - Tag and push a local image
    - SSH into EC2, pull the image from ECR, run it
  - **Learn:** Container registries, ECR authentication flow, image
    tagging, IAM permissions for ECR, how ECR compares to GHCR (which
    Charbel already uses).
  - **Understanding checkpoint:** "How is ECR different from GHCR? If your
    infrastructure is on AWS, why might you choose ECR over GHCR?"

- [ ] **Step 4.7: IAM roles and policies**
  - **Objective:** Give EC2 permissions without static credentials.
  - **Tasks:**
    - Create IAM role with trust policy for EC2
    - Attach inline policy allowing:
      - `ecr:GetAuthorizationToken`, `ecr:BatchGetImage`,
        `ecr:GetDownloadUrlForLayer`
      - `logs:CreateLogStream`, `logs:PutLogEvents`
    - Create instance profile, attach to EC2
    - Verify: SSH into EC2, pull from ECR WITHOUT access keys
  - **Learn:** IAM roles vs users vs access keys, instance profiles,
    trust policies, least-privilege JSON policies, why roles are safer
    than static credentials.
  - **Understanding checkpoint:** "Why is an IAM role attached to EC2
    safer than putting AWS access keys in a `.env` file on the server?
    What happens if the server is compromised in each case?"

- [ ] **Step 4.8: CloudWatch basics**
  - **Objective:** Understand AWS monitoring and alerting.
  - **Tasks:**
    - View EC2 metrics in CloudWatch (CPU, network, status checks)
    - Create alarm: CPU > 80% for 5 minutes -> notify
    - Enable VPC Flow Logs
    - Explore CloudWatch Logs (create a log group manually)
  - **Learn:** CloudWatch metrics vs logs vs alarms, evaluation periods,
    VPC Flow Logs, monitoring vs logging vs alerting, the three pillars
    of observability.
  - **Understanding checkpoint:** "What is the difference between a metric
    and a log? Give one example of each from this project. When would you
    use an alarm vs just looking at a dashboard?"

- [ ] **Step 4.9: DESTROY EVERYTHING**
  - **Objective:** Clean up all resources and understand dependency ordering.
  - **Tasks:**
    - Delete in reverse order: EC2 -> RDS -> ECR -> IGW ->
      subnets -> route tables -> security groups -> VPC
    - Release any allocated Elastic IPs
    - Verify AWS console shows no active resources
    - Verify billing dashboard shows $0 ongoing charges
  - **Learn:** Resource dependency graphs, why deletion order matters,
    cost awareness, orphaned resources (Elastic IPs, unattached volumes).
  - **Understanding checkpoint:** "Why did you have to delete the EC2 instance
    and IGW before the VPC? What is a dependency graph, and how does
    Terraform handle this ordering automatically?"

### Phase 4 Completion Criteria

- [ ] All resources created manually and understood
- [ ] All resources destroyed
- [ ] AWS console shows no active resources
- [ ] Billing dashboard shows $0 ongoing charges
- [ ] You can draw the full network architecture from memory
- [ ] No code commits for this phase

---

## Phase 5: Terraform Infrastructure as Code

**Goal:** Recreate the entire AWS infrastructure from Phase 4 using Terraform.
**Estimated time:** 4-5 days
**New concepts:** HCL syntax, providers, resources, variables, outputs,
state management, S3 backend, DynamoDB locking, data sources, depends_on,
tags.

### Steps

- [ ] **Step 5.1: Terraform setup and S3 backend**
  - **Objective:** Initialize Terraform with remote state storage.
  - **Tasks:**
    - Install Terraform 1.15+
    - Create `terraform/providers.tf` with AWS provider (region as variable)
    - Create S3 bucket for state (manually or via bootstrap config):
      versioning enabled, encryption enabled
    - Create DynamoDB table for state locking (partition key: `LockID`)
    - Create `terraform/backend.tf` with `backend "s3"` configuration
    - Run `terraform init`
  - **Files:** `terraform/providers.tf`, `terraform/backend.tf`,
    `terraform/variables.tf` (initial)
  - **Learn:** Terraform providers, backend configuration, why remote
    state, why state locking, S3 versioning for state protection,
    `terraform init` lifecycle.
  - **Understanding checkpoint:** "What would happen if two engineers ran
    `terraform apply` at the same time without DynamoDB locking? What
    specific corruption could occur in the state file?"
  - **Commit:** `infra: initialize Terraform with S3 backend and DynamoDB lock`

- [ ] **Step 5.2: VPC and networking**
  - **Objective:** Provision the network layer as code.
  - **Tasks:**
    - Create resources: `aws_vpc`, `aws_subnet` (public + private),
      `aws_internet_gateway`, `aws_route_table` (public + private),
      `aws_route_table_association` (NAT Gateway omitted for 100% Free Tier)
    - Use variables for CIDR blocks, region, project name
    - Tag every resource: `Name`, `Project = "aws-cloud-platform"`,
      `Environment = "dev"`
    - Run `terraform plan`, review, then `terraform apply`
  - **Files:** `terraform/network.tf` (or add to `main.tf`),
    `terraform/variables.tf` (update)
  - **Learn:** HCL resource syntax, `var.` references, resource
    dependencies (implicit via references, explicit via `depends_on`),
    `terraform plan` output reading, resource tagging conventions.
  - **Validation:** `terraform plan` shows expected resources.
    `terraform apply` creates them. AWS console matches Phase 4.
  - **Commit:** `infra: add VPC, subnets, gateways, and route tables`

- [ ] **Step 5.3: Security groups**
  - **Objective:** Create firewall rules as code.
  - **Tasks:**
    - Create `aws_security_group` for app (HTTP 80, 8080 inbound,
      all outbound)
    - Create `aws_security_group` for database (PostgreSQL 5432 inbound
      from app SG only, all outbound)
    - Reference VPC ID via `aws_vpc.main.id`
  - **Files:** `terraform/network.tf` (add SG resources)
  - **Learn:** Security group resources, referencing other resource
    attributes, SG-to-SG rules in Terraform, `aws_vpc_security_group_ingress_rule`.
  - **Commit:** `infra: add application and database security groups`

- [ ] **Step 5.4: RDS PostgreSQL**
  - **Objective:** Provision the managed database as code.
  - **Tasks:**
    - Create `aws_db_subnet_group` for private subnets
    - Create `aws_db_instance`: PostgreSQL 16, `db.t3.micro`,
      20GB storage, `skip_final_snapshot = true` (dev only)
    - Attach `db-sg` security group
    - Use `random_password` resource for master password
    - Mark password variable/output as `sensitive = true`
    - Enable automated backups
  - **Files:** `terraform/database.tf`, `terraform/variables.tf` (update)
  - **Learn:** `aws_db_instance`, `aws_db_subnet_group`, `random_password`,
    `sensitive` flag, `skip_final_snapshot`, why you never hardcode
    passwords in HCL.
  - **Understanding checkpoint:** "Why is `sensitive = true` important on
    the password variable? What happens in the `terraform plan` output
    if you omit it?"
  - **Commit:** `infra: add RDS PostgreSQL in private subnet`

- [ ] **Step 5.5: ECR repository**
  - **Objective:** Create the container registry as code.
  - **Tasks:**
    - Create `aws_ecr_repository` with image scanning on push
    - Add lifecycle policy: keep last 10 images, expire the rest
  - **Files:** `terraform/ecr.tf`
  - **Learn:** `aws_ecr_repository`, `aws_ecr_lifecycle_policy`, image
    scanning, why limiting image count matters (storage cost).
  - **Commit:** `infra: add ECR repository with lifecycle policy`

- [ ] **Step 5.6: IAM role for EC2**
  - **Objective:** Give EC2 permissions via a role, not static keys.
  - **Tasks:**
    - Create `aws_iam_role` with assume-role trust policy for
      `ec2.amazonaws.com`
    - Create `aws_iam_policy` with least-privilege JSON:
      ECR pull + CloudWatch Logs write
    - Create `aws_iam_role_policy_attachment`
    - Create `aws_iam_instance_profile`
  - **Files:** `terraform/iam.tf`
  - **Learn:** `aws_iam_role`, trust policies vs permission policies,
    `aws_iam_instance_profile`, least-privilege JSON, why EC2 needs
    both a role and an instance profile.
  - **Understanding checkpoint:** "What is the difference between an IAM
    role and an IAM instance profile? Why does EC2 need both? What would
    happen if you attached the role directly without the instance profile?"
  - **Commit:** `infra: add IAM role and instance profile for EC2`

- [ ] **Step 5.7: EC2 instance**
  - **Objective:** Provision the compute layer as code.
  - **Tasks:**
    - Use `aws_ami` data source to find latest Amazon Linux 2024 AMI
    - Create `aws_instance`: `t3.micro`, public subnet, `app-sg`,
      IAM instance profile, key pair
    - Add `user_data` script: install Docker, pull image from ECR,
      run container with environment variables
    - Reference RDS endpoint and ECR URL via outputs/variables
  - **Files:** `terraform/compute.tf`, `terraform/user_data.sh`
  - **Learn:** `aws_ami` data source, `aws_instance`, `user_data`
    bootstrap scripts, AMI selection, key pairs, referencing outputs
    from other resources.
  - **Validation:** `terraform apply` creates the full stack. SSH into
    EC2. Verify Docker is running. Verify app container is up. Verify
    `curl localhost:8080/health` returns 200 from inside the instance.
  - **Commit:** `infra: add EC2 instance with Docker bootstrap`

- [ ] **Step 5.8: CloudWatch monitoring**
  - **Objective:** Add centralized logging and alerting as code.
  - **Tasks:**
    - Create `aws_cloudwatch_log_group` with 14-day retention
    - Create `aws_cloudwatch_metric_alarm` for CPU > 80% (5 min)
    - Create `aws_cloudwatch_metric_alarm` for status check failures
    - Configure CloudWatch agent in `user_data` to ship app logs
  - **Files:** `terraform/monitoring.tf`, `terraform/user_data.sh` (update)
  - **Learn:** `aws_cloudwatch_log_group`, `aws_cloudwatch_metric_alarm`,
    alarm thresholds, evaluation periods, CloudWatch agent configuration,
    log retention policies.
  - **Commit:** `infra: add CloudWatch log groups and metric alarms`

- [ ] **Step 5.9: Outputs and full validation**
  - **Objective:** Expose important values and validate the entire stack.
  - **Tasks:**
    - Create `terraform/outputs.tf`: EC2 public IP, RDS endpoint,
      ECR repository URL, VPC ID
    - Run `terraform fmt` and `terraform validate`
    - Verify end-to-end: `curl http://<ec2-ip>:8080/health` returns 200
    - Verify `terraform plan` shows zero drift
  - **Files:** `terraform/outputs.tf`
  - **Validation:**
    - `curl http://<ec2-ip>:8080/health` returns 200
    - `curl http://<ec2-ip>:8080/ready` returns 200
    - `terraform plan` shows "No changes"
    - `terraform fmt -check` passes
    - `terraform validate` passes
  - **Commit:** `infra: add Terraform outputs and validate full stack`

### Phase 5 Completion Criteria

- [ ] `terraform init && terraform plan && terraform apply` creates the
    entire infrastructure
- [ ] `curl http://<ip>:8080/health` returns 200
- [ ] State is stored in S3 with DynamoDB locking
- [ ] All resources have `Name`, `Project`, `Environment` tags
- [ ] No hardcoded secrets anywhere in `.tf` files
- [ ] `terraform validate` and `terraform fmt -check` pass
- [ ] `terraform plan` shows zero drift after apply
- [ ] All committed on branch `phase-5-terraform`, merged to main

---

## Phase 6: CI/CD Deployment Pipeline

**Goal:** Every push to main automatically builds, pushes, deploys, and
validates.
**Estimated time:** 2-3 days
**New concepts:** ECR authentication in CI, Terraform in CI, deployment
strategies, health-check gates, automated rollback, concurrency control.

### Steps

- [ ] **Step 6.1: CD workflow - build and push to ECR**
  - **Objective:** Automate Docker image building and pushing on merge to main.
  - **Tasks:**
    - Create `.github/workflows/deploy.yml`
    - Trigger: `push` to `main` only
    - Add `concurrency` group to prevent parallel deployments
    - Steps: checkout, setup Go, run tests, build Docker image,
      authenticate to ECR (`aws-actions/amazon-ecr-login@v2`),
      tag with git SHA and `latest`, push both tags
    - Use GitHub environment secrets for AWS credentials
  - **Files:** `.github/workflows/deploy.yml`
  - **Learn:** ECR login action, image tagging strategies (SHA for
    traceability, `latest` for convenience), GitHub environment secrets,
    `concurrency` groups, why parallel deployments are dangerous.
  - **Understanding checkpoint:** "Why do we tag with the git SHA in
    addition to `latest`? If a deployment fails, how does the SHA tag
    help with rollback?"
  - **Commit:** `ci: add deployment workflow with ECR build and push`

- [ ] **Step 6.2: CD workflow - Terraform apply**
  - **Objective:** Run Terraform in CI to ensure infrastructure stays in sync.
  - **Tasks:**
    - Add steps: `hashicorp/setup-terraform@v3`, `terraform init`,
      `terraform plan`, `terraform apply`
    - Use GitHub environment protection rules (require manual approval
      before deploy job runs)
    - Pass Terraform variables via environment variables
  - **Files:** `.github/workflows/deploy.yml` (update)
  - **Learn:** Terraform in CI, `hashicorp/setup-terraform` action,
    environment protection rules, why `plan` before `apply` even in CI,
    Terraform variables in CI context.
  - **Commit:** `ci: add Terraform plan and apply to deployment pipeline`

- [ ] **Step 6.3: CD workflow - deploy, health check, rollback**
  - **Objective:** Deploy the new image to EC2 and validate with automated
    rollback.
  - **Tasks:**
    - After Terraform apply, deploy new image to EC2 via AWS SSM
      Run Command (or SSH): pull new image, stop old container,
      start new container
    - Health check: `curl` the `/health` endpoint in a retry loop
      (max 5 attempts, 10-second interval)
    - If health check fails: redeploy previous image tag (rollback)
    - Log deployment result
  - **Files:** `.github/workflows/deploy.yml` (update)
  - **Learn:** Deployment strategies, health-check gates, rollback
    patterns, AWS SSM Run Command as SSH alternative, retry logic in
    bash, why automated rollback matters in production.
  - **Understanding checkpoint:** "Why is automated rollback important?
    Describe what would happen in a production system serving real users
    if a bad deploy went unnoticed for one hour."
  - **Commit:** `ci: add deployment with health check and automatic rollback`

### Phase 6 Completion Criteria

- [ ] Push to main triggers the full pipeline automatically
- [ ] Image is built, pushed to ECR, and deployed to EC2
- [ ] Health check passes after deployment
- [ ] Rollback works if health check fails (test this deliberately)
- [ ] `concurrency` group prevents parallel deployments
- [ ] Environment protection rules require approval
- [ ] All committed on branch `phase-6-cicd`, merged to main

---

## Phase 7: Production Hardening and Bonus

**Goal:** Add centralized logging, Kubernetes exploration, Terraform
modules, and documentation.
**Estimated time:** 3-4 days
**New concepts:** CloudWatch agent, kind, Kubernetes Deployments / Services /
ConfigMaps, Terraform modules, technical documentation.

### Steps

- [ ] **Step 7.1: Centralized logging to CloudWatch**
  - **Objective:** Ship application logs to CloudWatch for centralized
    monitoring.
  - **Tasks:**
    - Configure CloudWatch agent on EC2 (via `user_data` or SSM)
    - Ship structured JSON logs from `slog` to the CloudWatch log group
    - Create a metric filter for `ERROR` level entries
    - Create an alarm on the error metric filter
  - **Files:** `terraform/monitoring.tf` (update),
    `terraform/user_data.sh` (update)
  - **Learn:** CloudWatch agent configuration, log group retention,
    metric filters, log-based alarms, structured logging in production.
  - **Validation:** Trigger an error in the app. Verify it appears in
    CloudWatch Logs. Verify the error metric increments.
  - **Commit:** `infra: configure CloudWatch agent for application log shipping`

- [ ] **Step 7.2: Kubernetes local exploration (kind)**
  - **Objective:** Understand Kubernetes fundamentals by deploying the
    same app locally.
  - **Tasks:**
    - Install `kind` and `kubectl`
    - Create cluster: `kind create cluster --name clouddev`
    - Write `k8s/deployment.yaml`: 2 replicas of the Go app, resource
      limits, liveness/readiness probes
    - Write `k8s/service.yaml`: ClusterIP service on port 80 -> 8080
    - Write `k8s/configmap.yaml`: environment variables
    - Apply: `kubectl apply -f k8s/`
    - Verify: `kubectl get pods`, `kubectl get svc`,
      `kubectl port-forward svc/api 8080:80`
    - Scale: `kubectl scale deployment/api --replicas=3`
    - Observe: `kubectl get pods -w`
  - **Files:** `k8s/deployment.yaml`, `k8s/service.yaml`,
    `k8s/configmap.yaml`
  - **Learn:** Pods, Deployments, Services, ConfigMaps, labels and
    selectors, `kubectl` commands, replica management, service discovery,
    liveness/readiness probes in Kubernetes, resource limits.
  - **Understanding checkpoint:** "What is the difference between a Pod
    and a Deployment? Why do you create a Deployment instead of just
    creating Pods directly? What happens if a Pod crashes?"
  - **Commit:** `feat: add Kubernetes manifests for local kind deployment`

- [ ] **Step 7.3: Terraform modules refactor**
  - **Objective:** Refactor monolithic Terraform into reusable modules.
  - **Tasks:**
    - Create `terraform/modules/network/` (VPC, subnets, gateways,
      route tables, security groups)
    - Create `terraform/modules/database/` (RDS, subnet group)
    - Create `terraform/modules/compute/` (EC2, IAM, instance profile)
    - Create `terraform/modules/storage/` (S3, DynamoDB, ECR)
    - Each module: `main.tf`, `variables.tf`, `outputs.tf`
    - Refactor root `terraform/main.tf` to call modules
    - Pass configuration via module variables
  - **Files:** `terraform/modules/*/main.tf`,
    `terraform/modules/*/variables.tf`,
    `terraform/modules/*/outputs.tf`,
    `terraform/main.tf` (refactored)
  - **Learn:** Terraform module syntax, `module` blocks, passing
    variables into modules, module outputs as inputs to other modules,
    why modules improve maintainability and reusability.
  - **Understanding checkpoint:** "If you had three clients each needing
    a VPC + RDS + EC2 with different CIDR blocks and instance sizes,
    how would modules help compared to copying `main.tf` three times?"
  - **Validation:** `terraform plan` shows ZERO changes. The refactor
    is purely structural. No resources are created, modified, or destroyed.
  - **Commit:** `infra: refactor Terraform into reusable modules`

- [ ] **Step 7.4: README and architecture documentation**
  - **Objective:** Write comprehensive documentation for the repository.
  - **Tasks:**
    - Write `README.md`: project description, architecture diagram
      (ASCII or Mermaid), prerequisites, local setup guide, deployment
      guide, Terraform guide, cost estimation, future improvements
    - Write `docs/architecture.md`: detailed design decisions, trade-offs
      (EC2 vs ECS, why Go stdlib, why Terraform over CloudFormation),
      network diagram, security model
    - Add a `COST.md` or section estimating monthly AWS cost
  - **Files:** `README.md`, `docs/architecture.md`
  - **Learn:** Technical writing, architecture documentation, cost
    estimation, trade-off analysis, writing for an audience (recruiters,
    interviewers, future engineers).
  - **Commit:** `docs: add comprehensive README and architecture documentation`

### Phase 7 Completion Criteria

- [ ] Application logs appear in CloudWatch Logs
- [ ] ERROR-level logs trigger a metric filter and alarm
- [ ] kind cluster runs the app with 2+ replicas
- [ ] `kubectl get pods` shows running pods
- [ ] `kubectl scale` works and new pods start
- [ ] Terraform is refactored into modules with zero drift
- [ ] README is comprehensive with architecture diagram
- [ ] Architecture document explains design decisions and trade-offs
- [ ] All committed on branch `phase-7-hardening`, merged to main

---

## Final Checklist

- [ ] All 7 phases completed and merged to main
- [ ] `terraform apply` creates the full infrastructure from scratch
- [ ] Push to main triggers CI/CD and deploys successfully
- [ ] Health checks pass in production
- [ ] CloudWatch shows logs and metrics
- [ ] Kubernetes manifests work locally with kind
- [ ] Terraform is modularized
- [ ] README is polished with architecture diagram
- [ ] Repository is public
- [ ] CV is updated with the project link
- [ ] Application to Digico Solutions is submitted

---

## Cost Estimation

All resources are free tier eligible. Estimated monthly cost if left running:

| Resource | Free Tier | Estimated Cost |
| --- | --- | --- |
| EC2 t3.micro | 750 hrs/month | $0 (within free tier) |
| RDS db.t3.micro | 750 hrs/month, 20GB | $0 (within free tier) |
| S3 (state bucket) | 5GB | ~$0.00 |
| ECR | 500MB | $0 |
| CloudWatch | 5GB logs, 10 alarms | $0 |
| NAT Gateway | Omitted (Paid service) | $0 (NAT GW removed) |
| Elastic IP / Public IP | Auto-assigned | $0 (free) |

**NOTE ON FREE TIER COMPLIANCE:** NAT Gateway has been intentionally omitted from the architecture because it is a paid service (~$32/month). By running EC2 in a public subnet with IGW routing and keeping RDS isolated in private subnets, the entire infrastructure remains **100% free ($0/month)** within the AWS Free Tier. Always run `terraform destroy` when completing testing to avoid any accidental charges.
