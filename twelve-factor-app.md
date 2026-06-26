# 📘 The Twelve-Factor App — Complete Guide

> A methodology for building modern, scalable, and maintainable software-as-a-service (SaaS) applications. Originally created by engineers at Heroku, the Twelve-Factor App is a set of best practices designed to help developers build apps that are **portable**, **resilient**, and **easy to scale**.

---

## 🌐 What is the Twelve-Factor App?

The Twelve-Factor App is a collection of 12 principles that describe how to build web applications and services that:

- Can be **deployed on any cloud platform** without code changes
- Are **easy to scale** horizontally (by adding more machines)
- Are **maintainable** by different developers over time
- Have **minimal divergence** between development and production environments
- Can **scale up** without significant changes to tools, architecture, or development practices

---

## The Twelve Factors

---

### I. 🗂️ Codebase
**"One codebase tracked in revision control, many deploys"**

#### What it means:
Your app should live in a **single code repository** (like Git), and that same codebase should be used to deploy to multiple environments (development, staging, production).

#### Key rules:
- **One app = One repo.** If you have multiple repos, it's a distributed system — each part should follow the 12-factor methodology independently.
- **Multiple deploys, one codebase.** The same code runs in dev, staging, and production — just with different configurations (more on that in Factor III).

#### ❌ Bad Practice:
- Having separate repos for `dev`, `staging`, and `production` branches that diverge over time.
- Copy-pasting code between multiple "versions" of an app.

#### ✅ Good Practice:
```
Repo: my-app
├── src/
├── tests/
└── ...

Deployed to:
  → production.myapp.com  (v1.0.3)
  → staging.myapp.com     (v1.0.4-rc1)
  → dev.myapp.com         (latest commit)
```

---

### II. 📦 Dependencies
**"Explicitly declare and isolate dependencies"**

#### What it means:
Never assume that a library or tool is pre-installed on the system. **Declare all dependencies explicitly** in a manifest file, and **isolate** them so the app only uses the declared versions.

#### Key rules:
- Use a **dependency declaration file** (e.g., `package.json`, `requirements.txt`, `Gemfile`, `pom.xml`).
- Use a **dependency isolation tool** (e.g., `npm`, `virtualenv`, `bundler`, `Maven`).
- No relying on system-wide packages like `curl`, `ImageMagick`, or Python libraries being globally installed.

#### ❌ Bad Practice:
```bash
# Assuming the user has ImageMagick installed globally
$ convert image.png output.jpg
```

#### ✅ Good Practice (Python example):
```
# requirements.txt
Flask==3.0.0
requests==2.31.0
Pillow==10.1.0
```
```bash
pip install -r requirements.txt
```

---

### III. ⚙️ Config
**"Store config in the environment"**

#### What it means:
Configuration is anything that **varies between deployments** (dev, staging, production). This includes database URLs, API keys, credentials, and feature flags. These should **never be hardcoded in code** — instead, store them in **environment variables**.

#### Key rules:
- **No secrets in source code** — ever.
- Use **environment variables** to pass configuration to the app at runtime.
- A good test: could you open-source your codebase right now without exposing any credentials? If yes, you're doing it right.

#### ❌ Bad Practice:
```javascript
// Hardcoded in code — dangerous!
const db = connect("postgres://admin:password123@prod-db.myapp.com/mydb");
const apiKey = "sk-abc123secretkey";
```

#### ✅ Good Practice:
```javascript
// Read from environment variables
const db = connect(process.env.DATABASE_URL);
const apiKey = process.env.API_KEY;
```
```bash
# Set in the environment (or a .env file locally, never committed to git)
DATABASE_URL=postgres://admin:password@localhost/mydb
API_KEY=sk-abc123secretkey
```

---

### IV. 🔌 Backing Services
**"Treat backing services as attached resources"**

#### What it means:
A **backing service** is any service the app consumes over the network as part of its normal operation — databases, message queues, email services, caching systems, etc. Treat all of them as **interchangeable attached resources**, accessed via a URL or credentials from the environment config.

#### Key rules:
- A local MySQL database and a third-party MySQL database should be **swappable without code changes** — just change the `DATABASE_URL` environment variable.
- There is **no distinction** between local and third-party services.

#### Examples of backing services:
| Type | Example |
|------|---------|
| Databases | PostgreSQL, MySQL, MongoDB |
| Message queues | RabbitMQ, Apache Kafka |
| Caching | Redis, Memcached |
| Email | SendGrid, Mailgun, SMTP |
| Storage | Amazon S3, Google Cloud Storage |

#### ✅ Good Practice:
```bash
# Swapping from local Postgres to AWS RDS is just a config change:
DATABASE_URL=postgres://user:pass@localhost/mydb        # local
DATABASE_URL=postgres://user:pass@aws-rds.endpoint/mydb # production
```

---

### V. 🔨 Build, Release, Run
**"Strictly separate build and run stages"**

#### What it means:
The process of turning code into a running app should be split into **three distinct, separate stages**:

| Stage | What happens |
|-------|-------------|
| **Build** | Code is compiled/transformed into an executable bundle (assets, binaries, dependencies) |
| **Release** | The build artifact is combined with the current environment config |
| **Run** | The release is executed in the target environment |

#### Key rules:
- **Never change code at runtime.** If you need a fix, go through Build → Release → Run again.
- Each release should have a **unique ID** (timestamp or version number) and be **immutable**.
- Releases can be rolled back to a previous version if something goes wrong.

#### Visualized:
```
Source Code
    │
    ▼ (Build stage)
Build Artifact  ←──── dependencies compiled
    │
    ▼ (Release stage)
Release = Artifact + Config  ←── env vars injected
    │
    ▼ (Run stage)
Running App in Production
```

---

### VI. ⚡ Processes
**"Execute the app as one or more stateless processes"**

#### What it means:
The app should run as **one or more stateless processes**. "Stateless" means the process stores **no data locally** between requests. Any data that needs to persist must be stored in a **backing service** (like a database or cache).

#### Key rules:
- **No sticky sessions.** Don't store session data in the process's memory or local filesystem.
- Each request should be **self-contained** — it shouldn't depend on data from a previous request being in memory.
- Use **Redis** or a database for session storage instead of in-memory session objects.

#### ❌ Bad Practice:
```javascript
// Storing user session in memory — breaks when you have multiple processes
const sessions = {};
app.post('/login', (req, res) => {
  sessions[req.body.userId] = { loggedIn: true };
});
```

#### ✅ Good Practice:
```javascript
// Store sessions in Redis (a backing service)
app.use(session({ store: new RedisStore({ client: redisClient }) }));
```

---

### VII. 🌐 Port Binding
**"Export services via port binding"**

#### What it means:
The app should be **self-contained** and expose its functionality via a **port**. It should not rely on an external web server (like Apache or Nginx) being injected at runtime. The app itself includes a web server library and binds to a port to serve requests.

#### Key rules:
- The app **listens on a port** specified by an environment variable (usually `PORT`).
- One app can become the **backing service** of another app — just point to its URL.

#### ✅ Good Practice (Node.js):
```javascript
const express = require('express');
const app = express();

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
});
```

```python
# Python (Flask)
if __name__ == '__main__':
    port = int(os.environ.get('PORT', 5000))
    app.run(host='0.0.0.0', port=port)
```

---

### VIII. 🔀 Concurrency
**"Scale out via the process model"**

#### What it means:
Scale your app by running **more processes** (horizontal scaling), not by making a single process bigger (vertical scaling). Different types of work should be handled by **different process types**.

#### Key rules:
- Define different **process types** for different workloads (e.g., `web` for HTTP requests, `worker` for background jobs).
- Scale each type independently based on demand.
- This works because processes are stateless (Factor VI) — you can just spin up more of them.

#### Example process types:
```
web:    node server.js       # Handles HTTP requests
worker: node worker.js       # Processes background jobs
clock:  node scheduler.js    # Runs cron-like scheduled tasks
```

#### Scaling visually:
```
Low traffic:    [web x1] [worker x1]
High traffic:   [web x5] [worker x3]
```

---

### IX. 🔄 Disposability
**"Maximize robustness with fast startup and graceful shutdown"**

#### What it means:
Processes should be **disposable** — they can be started or stopped at any moment without causing problems. This enables fast deployments, easy scaling, and resilience to hardware failures.

#### Key rules:
- **Fast startup:** Processes should be ready to serve requests within seconds of starting.
- **Graceful shutdown:** On receiving a `SIGTERM` signal, the process should stop accepting new requests, finish processing current requests, and exit cleanly.
- **Crash resilience:** The app should be designed to handle unexpected crashes — use a process manager that automatically restarts crashed processes.

#### ✅ Good Practice (Node.js graceful shutdown):
```javascript
process.on('SIGTERM', () => {
  console.log('Shutting down gracefully...');
  server.close(() => {
    console.log('All connections closed. Exiting.');
    process.exit(0);
  });
});
```

---

### X. 🔁 Dev/Prod Parity
**"Keep development, staging, and production as similar as possible"**

#### What it means:
Minimize the **gaps** between development and production environments. The three gaps to watch for are:

| Gap | Problem |
|-----|---------|
| **Time gap** | Code sits in dev for weeks before being deployed |
| **Personnel gap** | Developers write code, operations team deploys it |
| **Tools gap** | Dev uses SQLite, production uses PostgreSQL |

#### Key rules:
- **Deploy often** — ideally continuously (CI/CD pipelines).
- **Same tools in all environments** — if production uses PostgreSQL, dev should too. Don't use SQLite in dev "just because it's easier."
- **Use containers** (Docker) to ensure environment consistency.

#### ❌ Bad Practice:
```
Dev:        SQLite + Python 3.8 + local SMTP
Production: PostgreSQL + Python 3.11 + SendGrid
```

#### ✅ Good Practice:
```
Dev:        PostgreSQL + Python 3.11 + SendGrid (sandbox mode)
Production: PostgreSQL + Python 3.11 + SendGrid (live mode)
```

---

### XI. 📋 Logs
**"Treat logs as event streams"**

#### What it means:
A twelve-factor app **never manages its own log files**. Instead, it writes log events as an **unbuffered stream to `stdout`** (standard output). The environment (dev machine, cloud platform) is responsible for capturing and routing those logs.

#### Key rules:
- **Write to stdout only.** Don't open log files yourself.
- In development, the developer just watches the terminal output.
- In production, the infrastructure routes the stream to a log aggregation service (e.g., Datadog, Splunk, ELK Stack, Papertrail).

#### ❌ Bad Practice:
```python
# Writing logs to a file — you manage it, you rotate it, you lose it on crashes
logging.basicConfig(filename='app.log', level=logging.INFO)
```

#### ✅ Good Practice:
```python
# Write to stdout — let the platform handle the rest
import sys
logging.basicConfig(stream=sys.stdout, level=logging.INFO)
logger.info("User logged in: user_id=123")
```

---

### XII. 🛠️ Admin Processes
**"Run admin/management tasks as one-off processes"**

#### What it means:
Administrative and management tasks (database migrations, running scripts, inspecting data) should be run as **one-off processes** in the same environment as the app, using the same codebase and config — not as part of the running app itself.

#### Key rules:
- One-off admin tasks should run in an **identical environment** to the production app (same release, same config).
- They should be **run manually** or via automation, not baked into the app startup.
- **REPL access** (like `rails console` or `python manage.py shell`) to the production environment is acceptable for one-time inspections.

#### Common examples:
```bash
# Database migrations
python manage.py migrate

# Seed data
node scripts/seed-database.js

# One-time data cleanup
python manage.py fix_corrupted_records

# Open an interactive shell
rails console --environment=production
```

---

## 📊 Quick Reference Summary

| # | Factor | One-line Summary |
|---|--------|-----------------|
| I | Codebase | One repo, many deploys |
| II | Dependencies | Declare and isolate all dependencies |
| III | Config | Store secrets/config in environment variables |
| IV | Backing Services | Databases, queues, etc. are interchangeable resources |
| V | Build, Release, Run | Strictly separate these three stages |
| VI | Processes | Stateless processes — no local state |
| VII | Port Binding | App exposes itself via a port |
| VIII | Concurrency | Scale by running more processes |
| IX | Disposability | Fast start, graceful shutdown |
| X | Dev/Prod Parity | Dev and production should look the same |
| XI | Logs | Write to stdout, let the platform handle logs |
| XII | Admin Processes | Run one-off tasks in the same environment |

---

## 🎯 Why Does This Matter?

Following the twelve-factor methodology leads to apps that are:

- ✅ **Easy to onboard** — new developers can get started quickly
- ✅ **Cloud-native** — deploy anywhere (AWS, GCP, Azure, Heroku, etc.)
- ✅ **Resilient** — crashes are recoverable; processes are disposable
- ✅ **Scalable** — spin up more processes to handle load
- ✅ **Secure** — no secrets in code, config isolated in environment

---

## 📚 Further Reading

- [Official 12factor.net Website](https://12factor.net)
- [I. Codebase](https://12factor.net/codebase)
- [II. Dependencies](https://12factor.net/dependencies)
- [III. Config](https://12factor.net/config)
- [IV. Backing Services](https://12factor.net/backing-services)
- [V. Build, Release, Run](https://12factor.net/build-release-run)
- [VI. Processes](https://12factor.net/processes)
- [VII. Port Binding](https://12factor.net/port-binding)
- [VIII. Concurrency](https://12factor.net/concurrency)
- [IX. Disposability](https://12factor.net/disposability)
- [X. Dev/Prod Parity](https://12factor.net/dev-prod-parity)
- [XI. Logs](https://12factor.net/logs)
- [XII. Admin Processes](https://12factor.net/admin-processes)
