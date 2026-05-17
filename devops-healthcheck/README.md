# 🚀 DevOps Health Check System

A lightweight HTTP server written in Go that monitors the health of services, stores results in PostgreSQL, and includes a beautiful dark-themed UI — all running together via Docker Compose.

---

## 🎨 UI Screenshots

### Dashboard Overview
![Dashboard Overview](https://github.com/dipakrasal2009/DevOps-Cohort-Go/blob/main/devops-healthcheck/dashboard1.png)

### Services & Health Status
![Services Health Status](https://github.com/dipakrasal2009/DevOps-Cohort-Go/blob/main/devops-healthcheck/dashboard2.png)

---

## 📁 Project Structure

```
devops-healthcheck/
├── main.go               # HTTP server & route handlers
├── go.mod
├── go.sum
├── Dockerfile            # Multi-stage build for Go backend
├── Dockerfile.ui         # Nginx image for the UI
├── docker-compose.yml    # Postgres + App + UI together
├── nginx.conf            # Nginx config for UI container
├── ui.html               # Frontend dashboard
├── db/
│   └── db.go             # PostgreSQL connection & table setup
├── models/
│   └── service.go        # Service struct & CheckHealth() method
└── colorprint/
    └── print.go          # Color terminal output utility
```

---

## ⚙️ Prerequisites

- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [Go](https://golang.org/dl/) 1.21+ *(only for local development)*

---

## 🐳 Run Everything with Docker Compose (Recommended)

One command starts **PostgreSQL + Go backend + UI** together:

```bash
docker-compose up --build
```

| Service     | URL                   | Description         |
|-------------|-----------------------|---------------------|
| UI          | http://localhost:3000 | Frontend dashboard  |
| Backend API | http://localhost:8080 | Go HTTP server      |
| PostgreSQL  | localhost:5432        | Database            |

Expected output:
```
devops-pg           | database system is ready to accept connections
devops-healthcheck  | 🚀 DevOps Health Check System
devops-healthcheck  | ✅ Connected to PostgreSQL
devops-healthcheck  | ✅ Table ready
devops-healthcheck  | Server running on localhost:8080
devops-ui           | nginx: ready to accept connections
```

Then open **http://localhost:3000** in your browser.

### Docker Compose Commands

```bash
# Build and start all services
docker-compose up --build

# Run in background
docker-compose up --build -d

# Check logs
docker-compose logs -f

# Check logs for a specific service
docker-compose logs -f app
docker-compose logs -f ui
docker-compose logs -f postgres

# Stop everything
docker-compose down

# Stop and delete volumes (fresh DB)
docker-compose down -v
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

### Step 3 — Run the Backend

```bash
go run main.go
```

### Step 4 — Open the UI

```bash
python3 -m http.server 3000
# Then open http://localhost:3000/ui.html
```

---

## 🎨 UI Dashboard Features

The UI is served at **http://localhost:3000** and provides:

- **Stats bar** — live count of total, healthy, and unhealthy services
- **Register Service** — enter a name and URL to check and store a service
- **Quick Actions** — refresh all services or re-run all health checks at once
- **Service Cards** — visual green/red status, URL link, ID, and last checked timestamp

---

## 🛣️ API Routes

| Method | Route          | Description                             |
|--------|----------------|-----------------------------------------|
| GET    | `/`            | Home                                    |
| POST   | `/healthcheck` | Register a service and check its health |
| GET    | `/services`    | Get all registered services             |
| POST   | `/runall`      | Re-run health checks on all services    |

---

## 🧪 Test with curl

```bash
# Home
curl http://localhost:8080/

# Register & health check a service
curl -X POST http://localhost:8080/healthcheck \
  -H "Content-Type: application/json" \
  -d '{"Name":"google","URL":"https://google.com"}'

# Add another service
curl -X POST http://localhost:8080/healthcheck \
  -H "Content-Type: application/json" \
  -d '{"Name":"github","URL":"https://github.com"}'

# Add a broken service
curl -X POST http://localhost:8080/healthcheck \
  -H "Content-Type: application/json" \
  -d '{"Name":"broken","URL":"https://thisdomaindoesnotexist99999.com"}'

# Get all services
curl http://localhost:8080/services

# Re-run all health checks
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

### Images

| Image      | Dockerfile       | Base                              | Purpose                  |
|------------|------------------|-----------------------------------|--------------------------|
| `app`      | `Dockerfile`     | `golang:1.21-alpine` + `alpine`   | Go backend (multi-stage) |
| `ui`       | `Dockerfile.ui`  | `nginx:alpine`                    | Serves UI on port 80     |
| `postgres` | official image   | `postgres:latest`                 | Database                 |

### docker-compose Services

| Service    | Container            | Port mapping |
|------------|----------------------|--------------|
| `postgres` | `devops-pg`          | 5432 → 5432  |
| `app`      | `devops-healthcheck` | 8080 → 8080  |
| `ui`       | `devops-ui`          | 3000 → 80    |

---

## 🔧 Troubleshooting

**App can't reach database (`connection refused`)**
```bash
docker-compose down
docker-compose up --build
```

**Table constraint error (`42P10`)**
```bash
docker exec -it devops-pg psql -U admin -d healthcheck -c "DROP TABLE IF EXISTS services;"
docker-compose down && docker-compose up --build
```

**Old volume data causing PostgreSQL errors**
```bash
docker-compose down -v
docker-compose up --build
```

**UI can't connect to backend**
```bash
docker-compose ps
curl http://localhost:8080/
```

**Module path error (local dev)**
```bash
go mod init github.com/dipakrasal2009/DevOps-Cohort-Go/devops-healthcheck
go mod tidy
```

---

## 📦 Dependencies

| Package                   | Purpose                  |
|---------------------------|--------------------------|
| `github.com/lib/pq`       | PostgreSQL driver for Go |
| `github.com/fatih/color`  | Colored terminal output  |
