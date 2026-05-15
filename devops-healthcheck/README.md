# 🚀 DevOps Health Check System

A lightweight HTTP server written in Go that monitors the health of services and stores results in a PostgreSQL database running in Docker.

---

## 📁 Project Structure

```
devops-healthcheck/
├── main.go               # HTTP server & route handlers
├── go.mod
├── go.sum
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

---

## 🐘 Step 1 — Start PostgreSQL in Docker

```bash
docker run --name devops-pg \
  -e POSTGRES_USER=admin \
  -e POSTGRES_PASSWORD=admin123 \
  -e POSTGRES_DB=healthcheck \
  -p 5432:5432 \
  -d postgres
```

Verify the container is running:

```bash
docker ps
```

---

## 📦 Step 2 — Install Dependencies

```bash
go mod tidy
```

---

## ▶️ Step 3 — Run the Application

```bash
go run main.go
```

Expected output:

```
🚀 DevOps Health Check System
✅ Connected to PostgreSQL
✅ Table ready
Server running on localhost:8080
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

## 🧹 Cleanup

```bash
docker stop devops-pg
docker rm devops-pg
```

---

## 🔧 Troubleshooting

**Table constraint error (`42P10`)**
The old table exists without the `UNIQUE` constraint. Drop and recreate it:
```bash
docker exec -it devops-pg psql -U admin -d healthcheck -c "DROP TABLE IF EXISTS services;"
# Then restart: go run main.go
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
