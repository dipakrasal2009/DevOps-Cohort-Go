# 🚀 DevOps Health Check System

A lightweight HTTP server written in Go that monitors the health of services and stores results in a PostgreSQL database running in Docker.

---

## 📁 Project Structure

```
devops-healthcheck/
├── main.go               # HTTP server & route handlers
├── go.mod
├── go.sum
├── Dockerfile            # Multi-stage Docker build
├── docker-compose.yml    # App + PostgreSQL together
├── db/
│   └── db.go             # PostgreSQL connection & table setup
├── models/
│   └── service.go        # Service struct & CheckHealth() method
└── colorprint/
    └── print.go          # Color terminal output utility
```

---

## ⚙️ Prerequisites

- [Go](https://golang.org/dl/) 1.21+
- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)

---

## 🐳 Run with Docker Compose (Recommended)

This is the easiest way — it starts both the Go app and PostgreSQL together.

```bash
# Build and start everything
docker-compose up --build

# Run in background
docker-compose up --build -d

# Check logs
docker-compose logs -f

# Stop everything
docker-compose down

# Stop and delete volumes (fresh DB)
docker-compose down -v
```

Expected output:
```
devops-pg           | database system is ready to accept connections
devops-healthcheck  | 🚀 DevOps Health Check System
devops-healthcheck  | ✅ Connected to PostgreSQL
devops-healthcheck  | ✅ Table ready
devops-healthcheck  | Server running on localhost:8080
```

---

## 🖥️ Run Locally (Without Docker)

### Step 1 — Start PostgreSQL in Docker

```bash
docker run --name devops-pg \
  -e POSTGRES_USER=admin \
  -e POSTGRES_PASSWORD=admin123 \
  -e POSTGRES_DB=healthcheck \
  -p 5432:5432 \
  -d postgres
```

### Step 2 — Install Dependencies

```bash
go mod tidy
```

### Step 3 — Run the Application

```bash
go run main.go
```

---

## 🛣️ API Routes

| Method | Route | Description |
|--------|-------|-------------|
| GET | `/` | Home — health check of the server |
| POST | `/healthcheck` | Register a service and check its health |
| GET | `/services` | Get all registered services |
| POST | `/runall` | Re-run health checks on all stored services |

---

## 🧪 Test with curl

### Home
```bash
curl http://localhost:8080/
```

### Register & health check a service
```bash
curl -X POST http://localhost:8080/healthcheck \
  -H "Content-Type: application/json" \
  -d '{"Name":"google","URL":"https://google.com"}'
```

### Add another service
```bash
curl -X POST http://localhost:8080/healthcheck \
  -H "Content-Type: application/json" \
  -d '{"Name":"github","URL":"https://github.com"}'
```

### Add a broken/unhealthy service
```bash
curl -X POST http://localhost:8080/healthcheck \
  -H "Content-Type: application/json" \
  -d '{"Name":"broken","URL":"https://thisdomaindoesnotexist99999.com"}'
```

### Get all services
```bash
curl http://localhost:8080/services
```

### Re-run health checks on all stored services
```bash
curl -X POST http://localhost:8080/runall
```

---

## 🗄️ Database — Verify Data

```bash
# View all services
docker exec -it devops-pg psql -U admin -d healthcheck -c "SELECT * FROM services;"

# Check table structure
docker exec -it devops-pg psql -U admin -d healthcheck -c "\d services"

# Count total services
docker exec -it devops-pg psql -U admin -d healthcheck -c "SELECT COUNT(*) FROM services;"

# View only healthy services
docker exec -it devops-pg psql -U admin -d healthcheck -c "SELECT * FROM services WHERE healthy=true;"

# View only unhealthy services
docker exec -it devops-pg psql -U admin -d healthcheck -c "SELECT * FROM services WHERE healthy=false;"
```

---

## 🗃️ Database Schema

```sql
CREATE TABLE IF NOT EXISTS services (
    id        SERIAL PRIMARY KEY,
    name      VARCHAR(100) UNIQUE,
    url       VARCHAR(255),
    healthy   BOOLEAN,
    timestamp VARCHAR(50)
);
```

---

## 🐳 Docker Details

### Dockerfile (Multi-stage build)

| Stage | Base Image | Purpose |
|-------|-----------|---------|
| builder | `golang:1.21-alpine` | Compiles the Go binary |
| run | `alpine:latest` | Runs the binary (~15MB final image) |

### docker-compose.yml Services

| Service | Image | Port |
|---------|-------|------|
| `postgres` | `postgres:latest` | 5432 |
| `app` | built from `Dockerfile` | 8080 |

The app reads `DB_HOST` from the environment variable set in `docker-compose.yml`. When running locally without Docker it falls back to `localhost`.

---

## 🔧 Troubleshooting

**App can't reach database (`connection refused`)**
Make sure `db.go` reads `DB_HOST` from env and force rebuild:
```bash
docker-compose down
docker-compose up --build
```

**Table constraint error (`42P10`)**
The old table exists without the `UNIQUE` constraint. Drop and recreate:
```bash
docker exec -it devops-pg psql -U admin -d healthcheck -c "DROP TABLE IF EXISTS services;"
# Then restart
docker-compose down && docker-compose up --build
```

**Old volume data causing PostgreSQL errors**
Delete the volume and start fresh:
```bash
docker-compose down -v
docker-compose up --build
```

**Module path error**
Initialize the module with the correct path:
```bash
go mod init github.com/dipakrasal2009/DevOps-Cohort-Go/devops-healthcheck
go mod tidy
```

---

## 📦 Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/lib/pq` | PostgreSQL driver for Go |
| `github.com/fatih/color` | Colored terminal output |
