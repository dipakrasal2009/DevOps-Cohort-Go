# 🐳 Dockerfile — The Complete Guide

> **Bad Dockerfiles, Good Dockerfiles, Multistage builds, entrypoint.sh, SIGTERM vs SIGKILL — everything explained.**

[![Docker](https://img.shields.io/badge/Docker-Containers-2496ED?style=flat&logo=docker&logoColor=white)](https://docker.com)
[![Best Practices](https://img.shields.io/badge/Guide-Best_Practices-brightgreen?style=flat)](#good-dockerfile--everything-right)
[![Multistage](https://img.shields.io/badge/Build-Multistage-orange?style=flat)](#multistage-dockerfile)

---

## 📖 Table of Contents

- [What is a Dockerfile?](#what-is-a-dockerfile)
- [All Dockerfile Instructions](#all-dockerfile-instructions)
- [❌ Bad Dockerfile — Everything Wrong](#-bad-dockerfile--everything-wrong)
- [✅ Good Dockerfile — Everything Right](#-good-dockerfile--everything-right)
- [Docker Layer Cache — The Cache Trick](#docker-layer-cache--the-cache-trick)
- [Shell Form vs Exec Form](#shell-form-vs-exec-form)
- [ENTRYPOINT vs CMD](#entrypoint-vs-cmd)
- [entrypoint.sh — The Shell Script Pattern](#entrypointsh--the-shell-script-pattern)
- [SIGTERM vs SIGKILL — The Signal Story](#sigterm-vs-sigkill--the-signal-story)
- [Multistage Dockerfile](#multistage-dockerfile)
  - [Go Multistage Example](#go-multistage-example)
  - [Python Multistage Example](#python-multistage-example)
  - [Three-Stage with Testing](#three-stage-with-testing)
- [.dockerignore](#dockerignore)
- [Complete Production Example](#complete-production-example)
- [Quick Reference Cheatsheet](#quick-reference-cheatsheet)
- [Resources](#resources)

---

## What is a Dockerfile?

A Dockerfile is a plain text file containing sequential instructions that Docker reads top-to-bottom to **build a container image**. Each instruction creates a new **layer** in the image — and because Docker uses OverlayFS under the hood, each layer is a lowerdir stacked on top of the previous one.

```
Dockerfile instruction  →  image layer (lowerdir)  →  stacked via OverlayFS
```

```
FROM python:3.12-slim        →  layer 1  (base OS + Python)
RUN pip install flask        →  layer 2  (flask added)
COPY . .                     →  layer 3  (your code)
CMD ["python3", "app.py"]    →  layer 4  (metadata only)
                                   │
                              docker run
                                   │
                                   ▼
                             container upper/  ← your writes at runtime
```

---

## All Dockerfile Instructions

| Instruction | Purpose |
|-------------|---------|
| `FROM` | Base image to start from |
| `RUN` | Execute a command during build |
| `COPY` | Copy files from host into image |
| `ADD` | Like COPY but also handles URLs and auto-extracts tarballs |
| `WORKDIR` | Set working directory (creates it if missing) |
| `ENV` | Set environment variables (available at build + runtime) |
| `ARG` | Build-time variables (NOT in final image) |
| `EXPOSE` | Document which port the app listens on |
| `VOLUME` | Declare a mount point |
| `USER` | Set which user runs subsequent instructions and the container |
| `LABEL` | Add metadata as key=value pairs |
| `HEALTHCHECK` | Define how Docker checks if the container is healthy |
| `ENTRYPOINT` | Main process — harder to override at runtime |
| `CMD` | Default arguments — easily overridden at runtime |
| `SHELL` | Override the default shell used for shell-form instructions |
| `ONBUILD` | Trigger instruction when this image is used as a base |
| `STOPSIGNAL` | Set the signal sent to the container on `docker stop` |

---

## ❌ Bad Dockerfile — Everything Wrong

```dockerfile
FROM ubuntu:latest

RUN apt-get update
RUN apt-get install -y curl
RUN apt-get install -y python3
RUN apt-get install -y python3-pip
RUN pip3 install flask
RUN pip3 install gunicorn
RUN pip3 install requests

COPY . .

RUN chmod 777 /app

CMD python3 app.py
```

### What's wrong — every mistake explained

---

#### ❌ `FROM ubuntu:latest` — Unpinned base image

`latest` is a moving tag. When Ubuntu releases 24.04 next month, your build silently pulls a different OS. Packages change, APIs break, builds become non-reproducible.

```dockerfile
# ❌ Wrong
FROM ubuntu:latest

# ✅ Right — pin to an exact version
FROM ubuntu:22.04
# or better — use a purpose-built image
FROM python:3.12-slim
```

---

#### ❌ Separate `RUN apt-get update` — Stale cache bug

```dockerfile
# ❌ Wrong — these are two separate layers
RUN apt-get update          # cached from 3 weeks ago
RUN apt-get install -y curl # uses stale package lists → install failure
```

Docker caches each `RUN` layer. The `apt-get update` layer gets cached on day 1. Three weeks later when you add a new package, Docker reuses the old cached `update` layer but tries to install from outdated package lists. This causes cryptic "package not found" errors.

```dockerfile
# ✅ Right — always chain update + install in ONE RUN
RUN apt-get update && apt-get install -y --no-install-recommends \
      curl \
      python3 \
    && rm -rf /var/lib/apt/lists/*
```

---

#### ❌ Seven separate `RUN pip3 install` — Layer explosion

```dockerfile
# ❌ Wrong — 7 layers for pip installs
RUN pip3 install flask
RUN pip3 install gunicorn
RUN pip3 install requests
RUN pip3 install sqlalchemy
RUN pip3 install celery
RUN pip3 install redis
RUN pip3 install boto3
```

Each `RUN` creates a new image layer. More layers = larger image = slower push/pull. And each layer stores the full diff, including pip's temporary download files if you don't clean up in the same layer.

```dockerfile
# ✅ Right — one layer, use requirements.txt
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
```

---

#### ❌ `COPY . .` with no `.dockerignore` — Secret leakage

```dockerfile
# ❌ Wrong — copies EVERYTHING
COPY . .
# Includes: .git/, .env, secrets/, node_modules/, *.pyc, test data...
```

Without `.dockerignore`, you copy your entire project into the image — including secrets, git history, test fixtures, and local configs. Once in a layer, they're in the image **forever** even if you delete them in a later `RUN`.

```dockerfile
# ✅ Right — selective copy, with .dockerignore in place
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY src/ ./src/
COPY config/ ./config/
```

---

#### ❌ `chmod 777` — Security disaster

```dockerfile
# ❌ Wrong — world-writable application files
RUN chmod 777 /app
```

777 means any user on the system (including attackers who escape the container) can read, write, and execute your app files. This is almost never correct.

```dockerfile
# ✅ Right — restrictive permissions, non-root owner
RUN addgroup --system appgroup \
    && adduser --system --ingroup appgroup appuser \
    && chown -R appuser:appgroup /app \
    && chmod 755 /app
USER appuser
```

---

#### ❌ `CMD python3 app.py` — Shell form, broken signals

```dockerfile
# ❌ Wrong — shell form
CMD python3 app.py
# Becomes: /bin/sh -c "python3 app.py"
# PID 1 = sh
# PID 2 = python3 (never receives SIGTERM from Docker)
```

When Docker stops a container, it sends SIGTERM to PID 1. With shell form, PID 1 is `/bin/sh`, which does nothing with SIGTERM. After 10 seconds, Docker sends SIGKILL — force killing your app with no chance for graceful shutdown.

```dockerfile
# ✅ Right — exec form
CMD ["python3", "app.py"]
# PID 1 = python3 (receives SIGTERM directly) ✅
```

---

#### ❌ No `WORKDIR` — Chaotic filesystem

```dockerfile
# ❌ Wrong — no WORKDIR set
COPY . .  # copies to /  (root of container filesystem!)
```

Without `WORKDIR`, files land in `/` — the root of the container's filesystem. You'll overwrite system files and create a mess that's impossible to reason about.

```dockerfile
# ✅ Right
WORKDIR /app
COPY . .  # copies to /app/
```

---

#### ❌ Running as root — Container escape risk

Without a `USER` instruction, everything runs as `root`. If an attacker exploits your app, they have root access inside the container — and potentially beyond.

```dockerfile
# ✅ Right
RUN addgroup --system appgroup && adduser --system --ingroup appgroup appuser
USER appuser
```

---

## ✅ Good Dockerfile — Everything Right

```dockerfile
# ── Pin exact version — reproducible builds ────────────────
FROM python:3.12-slim

# ── Metadata ──────────────────────────────────────────────
LABEL maintainer="you@example.com"
LABEL version="1.0.0"
LABEL description="Production Python application"

# ── Working directory ─────────────────────────────────────
WORKDIR /app

# ── Copy dependencies FIRST (cache optimization) ──────────
COPY requirements.txt .

# ── Install in ONE layer, clean up in same layer ──────────
RUN apt-get update && apt-get install -y --no-install-recommends \
      curl \
      gcc \
    && rm -rf /var/lib/apt/lists/* \
    && pip install --no-cache-dir -r requirements.txt \
    && apt-get purge -y gcc \
    && apt-get autoremove -y

# ── Copy app code AFTER dependencies ──────────────────────
COPY . .

# ── Create non-root user ───────────────────────────────────
RUN addgroup --system appgroup \
    && adduser --system --ingroup appgroup appuser \
    && chown -R appuser:appgroup /app

# ── Switch to non-root ─────────────────────────────────────
USER appuser

# ── Document port ──────────────────────────────────────────
EXPOSE 8000

# ── Health check ───────────────────────────────────────────
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:8000/health || exit 1

# ── exec form — PID 1, correct signal handling ─────────────
ENTRYPOINT ["./entrypoint.sh"]
CMD ["gunicorn", "app:app", "--bind", "0.0.0.0:8000", "--workers", "4"]
```

---

## Docker Layer Cache — The Cache Trick

The single most impactful optimization in a Dockerfile. Docker caches each layer and reuses it if nothing above it changed. The strategy: **put things that change rarely near the top, things that change often near the bottom.**

```
FROM python:3.12-slim          ← almost never changes     → cached forever
WORKDIR /app                   ← never changes            → cached forever
COPY requirements.txt .        ← changes rarely           → cached until deps change
RUN pip install -r req...      ← changes with deps        → cached until deps change
COPY . .                       ← changes on every commit  → always rebuilds from here
```

```
First build:           Second build (code change only):
FROM    ✅ pull        FROM    ✅ cache hit
WORKDIR ✅ create      WORKDIR ✅ cache hit
COPY    ✅ copy        COPY    ✅ cache hit (requirements.txt unchanged)
RUN pip ✅ install     RUN pip ✅ cache hit ← saves 3 minutes!
COPY .  ✅ copy        COPY .  🔄 rebuild (code changed)
```

**Wrong order (cache-busting pattern):**
```dockerfile
# ❌ Code changes bust the pip install cache every time
COPY . .                    ← changes every commit
RUN pip install -r req.txt  ← reinstalls everything every build
```

---

## Shell Form vs Exec Form

Every `RUN`, `CMD`, and `ENTRYPOINT` supports two syntaxes with very different behavior:

```dockerfile
# Shell form — invokes /bin/sh -c
RUN apt-get update
CMD python3 app.py
ENTRYPOINT python3 app.py

# Exec form — runs directly, no shell involved
RUN ["apt-get", "update"]
CMD ["python3", "app.py"]
ENTRYPOINT ["python3", "app.py"]
```

| Aspect | Shell Form | Exec Form |
|--------|-----------|-----------|
| Shell invoked | ✅ `/bin/sh -c` | ❌ direct exec |
| PID 1 | `sh` | your binary |
| SIGTERM received by app | ❌ No | ✅ Yes |
| Variable expansion | ✅ `$HOME` works | ❌ manual `["sh","-c","echo $HOME"]` |
| Use for CMD/ENTRYPOINT | ❌ Never | ✅ Always |
| Use for RUN | fine | fine |

**Critical for CMD and ENTRYPOINT — always use exec form.**

---

## ENTRYPOINT vs CMD

| | `ENTRYPOINT` | `CMD` |
|--|-------------|-------|
| Purpose | Define the main process | Define default arguments |
| Override at runtime | Needs `--entrypoint` flag | Just pass args after image name |
| If both defined | ENTRYPOINT runs, CMD becomes its args | — |

```dockerfile
# Pattern 1: CMD only
CMD ["python3", "app.py"]

docker run myimage               # → python3 app.py
docker run myimage bash          # → bash  (CMD completely replaced)

# ─────────────────────────────────────────────────

# Pattern 2: ENTRYPOINT only
ENTRYPOINT ["python3"]

docker run myimage               # → python3 (no args — probably wrong)
docker run myimage app.py        # → python3 app.py
docker run myimage bash          # → python3 bash (not what you want)

# ─────────────────────────────────────────────────

# Pattern 3: ENTRYPOINT + CMD (best for production)
ENTRYPOINT ["python3"]
CMD ["app.py"]

docker run myimage               # → python3 app.py       (default)
docker run myimage other.py      # → python3 other.py     (CMD overridden)
docker run --entrypoint bash myimage  # → bash            (both overridden)

# ─────────────────────────────────────────────────

# Pattern 4: entrypoint.sh + CMD (most flexible)
ENTRYPOINT ["./entrypoint.sh"]
CMD ["gunicorn", "app:app", "--bind", "0.0.0.0:8000"]

docker run myimage                          # → entrypoint.sh gunicorn app:app ...
docker run myimage uvicorn app:app          # → entrypoint.sh uvicorn app:app
```

---

## `entrypoint.sh` — The Shell Script Pattern

For production containers, an `entrypoint.sh` shell script handles setup before the main process starts: validating environment variables, waiting for dependencies, running migrations.

```bash
#!/bin/sh
# entrypoint.sh
set -e   # exit immediately on any error
set -u   # treat unset variables as errors

echo "[entrypoint] Container starting..."

# ── 1. Validate required environment variables ─────────────
required_vars="DATABASE_URL SECRET_KEY APP_ENV"

for var in $required_vars; do
  eval val=\$$var
  if [ -z "$val" ]; then
    echo "[entrypoint] ERROR: Required variable $var is not set"
    exit 1
  fi
done

echo "[entrypoint] Environment variables OK"

# ── 2. Wait for dependent services ─────────────────────────
if [ -n "$DB_HOST" ]; then
  echo "[entrypoint] Waiting for database at $DB_HOST:${DB_PORT:-5432}..."
  until pg_isready -h "$DB_HOST" -p "${DB_PORT:-5432}" -q; do
    echo "[entrypoint] Database not ready — retrying in 2s..."
    sleep 2
  done
  echo "[entrypoint] Database is ready!"
fi

# ── 3. Run database migrations ─────────────────────────────
if [ "$APP_ENV" = "production" ] || [ "$RUN_MIGRATIONS" = "true" ]; then
  echo "[entrypoint] Running database migrations..."
  python3 manage.py migrate --noinput
  echo "[entrypoint] Migrations complete"
fi

# ── 4. Create runtime directories ──────────────────────────
mkdir -p /tmp/app/logs /tmp/app/uploads

# ── 5. Hand off to the main process ────────────────────────
echo "[entrypoint] Starting: $*"

# THIS IS THE CRITICAL LINE — exec replaces the shell with
# the main process. Your app becomes PID 1 and receives signals.
exec "$@"
```

### The `exec "$@"` — Why It Is Everything

```
WITHOUT exec "$@":

  docker stop
       │
       ▼ SIGTERM
  PID 1 = /bin/sh (entrypoint.sh)
  PID 2 = gunicorn

  sh does nothing with SIGTERM.
  10 seconds pass.
  SIGKILL → both processes force-killed.
  → Requests dropped, database connections broken,
    buffers not flushed, data potentially corrupted.


WITH exec "$@":

  docker stop
       │
       ▼ SIGTERM
  PID 1 = gunicorn   ← sh was replaced by exec

  gunicorn catches SIGTERM:
  → finishes in-flight requests
  → closes database connections gracefully
  → flushes write buffers
  → exits cleanly with code 0
```

`exec` replaces the current shell process with the given command. No new PID is created — the PID stays the same, but the process image changes from `sh` to `gunicorn`. The shell is gone. Your app is now PID 1.

**`"$@"` expands to all arguments passed to the script.** In the Dockerfile:
```dockerfile
ENTRYPOINT ["./entrypoint.sh"]
CMD ["gunicorn", "app:app", "--bind", "0.0.0.0:8000"]
```
At runtime, Docker calls: `./entrypoint.sh gunicorn app:app --bind 0.0.0.0:8000`
Inside the script: `"$@"` = `gunicorn app:app --bind 0.0.0.0:8000`
`exec "$@"` = `exec gunicorn app:app --bind 0.0.0.0:8000`

---

## SIGTERM vs SIGKILL — The Signal Story

```
docker stop <container>  (or  docker-compose down)
      │
      ├─► sends SIGTERM (signal 15) to PID 1
      │         │
      │         │  app has time to:
      │         │  ✅ finish in-flight HTTP requests
      │         │  ✅ close database connections
      │         │  ✅ flush write buffers
      │         │  ✅ checkpoint state
      │         │  ✅ deregister from service discovery
      │         │
      │    waits --stop-timeout (default: 10 seconds)
      │
      └─► if still running → sends SIGKILL (signal 9)
                   │
                   │  kernel force-terminates immediately
                   │  ❌ no cleanup possible
                   │  ❌ requests dropped mid-flight
                   │  ❌ database connections forcibly closed
                   │  ❌ potential data corruption
```

| Signal | Number | Catchable | Behavior |
|--------|--------|-----------|----------|
| `SIGTERM` | 15 | ✅ Yes | Polite "please shut down" — app can handle |
| `SIGKILL` | 9 | ❌ Never | Kernel-level force kill — instant, no escape |
| `SIGINT` | 2 | ✅ Yes | Ctrl+C — usually same as SIGTERM |
| `SIGHUP` | 1 | ✅ Yes | Terminal closed / reload config |

**Handling SIGTERM in Python:**

```python
import signal
import sys

def graceful_shutdown(signum, frame):
    print("SIGTERM received — starting graceful shutdown...")
    server.stop(grace=5)      # finish current requests (5s max)
    db_pool.close()           # close all DB connections
    cache.close()             # close Redis connections
    print("Shutdown complete")
    sys.exit(0)

signal.signal(signal.SIGTERM, graceful_shutdown)
signal.signal(signal.SIGINT, graceful_shutdown)
```

**Handling SIGTERM in Node.js:**

```javascript
process.on('SIGTERM', async () => {
  console.log('SIGTERM received — shutting down gracefully');
  await server.close();
  await db.end();
  process.exit(0);
});
```

**Adjust the grace period:**

```bash
docker stop --time=30 mycontainer   # give 30s instead of 10s

# In docker-compose.yml:
services:
  app:
    stop_grace_period: 30s
```

---

## Multistage Dockerfile

Multistage builds let you use **multiple `FROM` instructions** in one Dockerfile. Each `FROM` starts a new stage. You `COPY --from=<stage>` to pull artifacts between stages. Only the **final stage** is shipped.

**The problem multistage solves:**

```
Single stage build:
  Go source code + Go compiler + all build tools + final binary = 800 MB image

Multistage build:
  Stage 1 (builder): Go compiler + source → compile → binary
  Stage 2 (runtime): binary only = 8 MB image

The compiler never ships. The source never ships.
```

---

### Go Multistage Example

```dockerfile
# ══════════════════════════════════════════════════════════
# Stage 1: builder
# Large image with full Go toolchain
# ══════════════════════════════════════════════════════════
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Download dependencies first (cache optimization)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and compile
COPY . .

# Build a static binary (no external libc dependencies)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o app ./cmd/app

# ══════════════════════════════════════════════════════════
# Stage 2: runtime
# FROM scratch = absolutely empty image (0 MB base)
# ══════════════════════════════════════════════════════════
FROM scratch

# Copy only the compiled binary
COPY --from=builder /build/app /app

# Copy TLS certificates (needed for HTTPS calls)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data if your app uses time.LoadLocation
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

EXPOSE 8080

ENTRYPOINT ["/app"]
```

**Image sizes:**
```
golang:1.22-alpine + source + build tools  →  ~600 MB
Final image (scratch + binary)             →  ~8 MB
```

---

### Python Multistage Example

```dockerfile
# ══════════════════════════════════════════════════════════
# Stage 1: builder
# Full Python image with gcc for compiling C extensions
# ══════════════════════════════════════════════════════════
FROM python:3.12 AS builder

WORKDIR /build

COPY requirements.txt .

# Install packages into /install prefix (not system Python)
RUN pip install --no-cache-dir --prefix=/install -r requirements.txt

# ══════════════════════════════════════════════════════════
# Stage 2: runtime
# Slim image — no gcc, no header files, no build tools
# ══════════════════════════════════════════════════════════
FROM python:3.12-slim AS runtime

WORKDIR /app

# Copy the installed packages from builder
COPY --from=builder /install /usr/local

# Copy only application code
COPY src/ ./src/
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

# Non-root user
RUN addgroup --system appgroup \
    && adduser --system --ingroup appgroup appuser \
    && chown -R appuser:appgroup /app

USER appuser

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD curl -f http://localhost:8000/health || exit 1

ENTRYPOINT ["./entrypoint.sh"]
CMD ["gunicorn", "src.app:app", "--bind", "0.0.0.0:8000", "--workers", "4"]
```

---

### Three-Stage with Testing

Run tests in CI, build production — all in one Dockerfile:

```dockerfile
# ══════════════════════════════════════════════════════════
# Stage 1: base — shared by test and production
# ══════════════════════════════════════════════════════════
FROM python:3.12-slim AS base

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# ══════════════════════════════════════════════════════════
# Stage 2: test — used in CI pipeline, never shipped
# ══════════════════════════════════════════════════════════
FROM base AS test

COPY requirements-dev.txt .
RUN pip install --no-cache-dir -r requirements-dev.txt

COPY . .

# Run tests during build — build fails if tests fail
RUN python -m pytest tests/ -v --tb=short

# ══════════════════════════════════════════════════════════
# Stage 3: production — the only stage that ships
# ══════════════════════════════════════════════════════════
FROM base AS production

COPY src/ ./src/
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

RUN addgroup --system appgroup \
    && adduser --system --ingroup appgroup appuser \
    && chown -R appuser:appgroup /app

USER appuser

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD curl -f http://localhost:8000/health || exit 1

ENTRYPOINT ["./entrypoint.sh"]
CMD ["gunicorn", "src.app:app", "--bind", "0.0.0.0:8000"]
```

**Using targeted builds:**

```bash
# Run only tests (CI)
docker build --target test -t myapp:test .

# Build only production image (deploy)
docker build --target production -t myapp:latest .

# Run tests then build production in CI pipeline
docker build --target test -t myapp:test . \
  && docker build --target production -t myapp:latest .
```

---

## `.dockerignore`

Always create a `.dockerignore` file. It works like `.gitignore` — patterns of files/dirs to exclude from the build context sent to the Docker daemon.

```
# .dockerignore

# Version control
.git/
.gitignore

# Python artifacts
__pycache__/
*.pyc
*.pyo
*.pyd
.Python
*.egg-info/
dist/
build/
.eggs/

# Virtual environments
.venv/
venv/
env/

# Tests and dev tools
tests/
.pytest_cache/
.coverage
htmlcov/
.mypy_cache/
.ruff_cache/

# Secrets and local config
.env
.env.*
*.key
*.pem
secrets/

# Docker files (no need to copy these into image)
Dockerfile
Dockerfile.*
.dockerignore
docker-compose*.yml

# Editor and OS artifacts
.idea/
.vscode/
*.swp
.DS_Store
Thumbs.db

# Logs
*.log
logs/
```

**Why it matters:**
- Smaller build context → faster `docker build`
- Prevents secrets from leaking into image layers
- Prevents cache invalidation from irrelevant file changes

---

## Complete Production Example

A full production-ready setup with all files:

### `Dockerfile`

```dockerfile
FROM python:3.12-slim AS base

LABEL maintainer="team@example.com"
LABEL version="1.0.0"

WORKDIR /app

COPY requirements.txt .
RUN apt-get update && apt-get install -y --no-install-recommends \
      curl \
      postgresql-client \
    && rm -rf /var/lib/apt/lists/* \
    && pip install --no-cache-dir -r requirements.txt

FROM base AS production

COPY src/ ./src/
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

RUN addgroup --system appgroup \
    && adduser --system --ingroup appgroup --no-create-home appuser \
    && chown -R appuser:appgroup /app

USER appuser

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD curl -f http://localhost:8000/health || exit 1

ENTRYPOINT ["./entrypoint.sh"]
CMD ["gunicorn", "src.app:app", "--bind", "0.0.0.0:8000", "--workers", "4", "--timeout", "30"]
```

### `entrypoint.sh`

```bash
#!/bin/sh
set -e
set -u

log() { echo "[entrypoint] $*"; }

log "Starting container..."

# Validate environment
for var in DATABASE_URL SECRET_KEY; do
  eval val=\$$var
  if [ -z "$val" ]; then
    log "ERROR: $var is required but not set"
    exit 1
  fi
done

# Wait for database
if [ -n "${DB_HOST:-}" ]; then
  log "Waiting for database at $DB_HOST..."
  until pg_isready -h "$DB_HOST" -p "${DB_PORT:-5432}" -q; do
    sleep 2
  done
  log "Database ready"
fi

# Migrations
if [ "${RUN_MIGRATIONS:-false}" = "true" ]; then
  log "Running migrations..."
  python3 -m alembic upgrade head
fi

log "Handing off to: $*"
exec "$@"
```

### `requirements.txt`

```
flask==3.0.0
gunicorn==21.2.0
psycopg2-binary==2.9.9
alembic==1.13.0
sqlalchemy==2.0.23
```

### `.dockerignore`

```
.git/
.venv/
__pycache__/
*.pyc
.env
tests/
*.log
.dockerignore
docker-compose*.yml
```

### Build and run

```bash
# Build
docker build --target production -t myapp:1.0.0 .

# Run
docker run -d \
  --name myapp \
  -p 8000:8000 \
  -e DATABASE_URL=postgresql://user:pass@db:5432/mydb \
  -e SECRET_KEY=supersecret \
  -e DB_HOST=db \
  -e RUN_MIGRATIONS=true \
  myapp:1.0.0

# Check logs
docker logs myapp

# Graceful stop (SIGTERM → 30s grace → SIGKILL if needed)
docker stop --time=30 myapp
```

---

## Quick Reference Cheatsheet

```dockerfile
# ── Base image — always pin ────────────────────────────────
FROM python:3.12-slim          # ✅ pinned
FROM ubuntu:latest             # ❌ unpinned

# ── Layer cache — deps before code ────────────────────────
COPY requirements.txt .        # ✅ cache deps separately
RUN pip install -r req.txt
COPY . .

# ── One RUN, clean in same layer ──────────────────────────
RUN apt-get update && apt-get install -y pkg \
    && rm -rf /var/lib/apt/lists/*

# ── Exec form for CMD/ENTRYPOINT ──────────────────────────
CMD ["python3", "app.py"]      # ✅ exec form → PID 1
CMD python3 app.py             # ❌ shell form → sh is PID 1

# ── entrypoint.sh — always end with exec ──────────────────
exec "$@"                      # ✅ replaces shell with app

# ── Non-root user ─────────────────────────────────────────
RUN adduser --system appuser
USER appuser

# ── Multistage — builder + runtime ────────────────────────
FROM golang:1.22 AS builder
RUN go build -o app .

FROM scratch
COPY --from=builder /app /app

# ── Build a specific stage ─────────────────────────────────
docker build --target production -t myapp .

# ── Graceful stop with grace period ───────────────────────
docker stop --time=30 mycontainer
```

---

## Resources

- 📘 [Dockerfile reference](https://docs.docker.com/engine/reference/builder/)
- 📘 [Docker best practices](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
- 📘 [Multistage builds](https://docs.docker.com/build/building/multi-stage/)
- 📘 [.dockerignore reference](https://docs.docker.com/engine/reference/builder/#dockerignore-file)
- 🔗 [Docker storage drivers (OverlayFS)](https://docs.docker.com/storage/storagedriver/overlayfs-driver/)
- 🔗 [Graceful shutdown in containers](https://cloud.google.com/blog/products/containers-kubernetes/kubernetes-best-practices-terminating-with-grace)

---

<div align="center">

**Star ⭐ if this saved you from a bad Dockerfile in production!**

</div>
