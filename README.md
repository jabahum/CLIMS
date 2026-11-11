# CLIMS Full-Stack Application

A containerized full-stack application built with **Next.js**, **Go (Golang)**, **MariaDB**, and **Nginx**.

---

## 🏗️ Architecture Overview

The application is structured into four main isolated components managed via Docker Compose:

1. **Frontend**: Next.js App Router (TypeScript, Tailwind CSS).
2. **Backend**: Go API server.
3. **Database**: MariaDB instance pre-initialized with `clims_dump_11Nov2026.sql`.
4. **Reverse Proxy**: Nginx acting as a unified entry point routing traffic between the frontend and backend.

```
                  ┌──────────────────────┐
                  │     Client / Browser │
                  └──────────┬───────────┘
                             │
                             │ (Port 80)
                             ▼
                  ┌──────────────────────┐
                  │     Nginx Proxy      │
                  └──────┬────────┬──────┘
                         │        │
           / (Frontend)  │        │  /api/ (Backend)
                         ▼        ▼
        ┌──────────────────┐    ┌──────────────────┐
        │ Next.js Frontend │    │    Go Backend    │
        └──────────────────┘    └────────┬─────────┘
                                         │
                                         │ (Port 3306)
                                         ▼
                                ┌──────────────────┐
                                │ MariaDB Database │
                                └──────────────────┘
```

---

## 📂 Directory Structure

```text
clims/
├── backend/                  # Go Backend Service
│   ├── cmd/
│   │   └── server/          # Main entry point (main.go)
│   ├── internal/            # Business logic, database handlers, models
│   ├── Dockerfile           # Multi-stage Docker build for Go
│   ├── go.mod
│   └── go.sum
├── frontend/                 # Next.js Frontend Application
│   ├── src/
│   │   ├── app/             # App Router pages and layouts
│   │   ├── components/      # UI components
│   │   └── lib/             # API helpers and client utilities
│   ├── Dockerfile           # Multi-stage Docker build for Next.js
│   ├── package.json
│   └── tailwind.config.js
├── nginx/                    # Reverse Proxy Configurations
│   ├── app.conf              # Server block routing rules
│   └── nginx.conf            # Global Nginx directives
├── clims_dump_11Nov2026.sql  # Initial database initialization dump
├── dev.env                   # Development environment variables
├── prod.env                  # Production environment variables
├── dev.env.example           # Example environment template
└── docker-compose.yml        # Master Docker Compose orchestration file
```

---

## ⚙️ Environment Configuration

The application uses environment files (`dev.env` and `prod.env`) to manage environment settings safely across environments.

### Required Environment Variables

| Variable              | Description              | Dev Default             | Prod Default                 |
| :-------------------- | :----------------------- | :---------------------- | :--------------------------- |
| `MYSQL_ROOT_PASSWORD` | MariaDB root password    | `root_password`         | *Secure Password*            |
| `MYSQL_DATABASE`      | Database name            | `clims_db`              | `clims_db`                   |
| `MYSQL_USER`          | Database username        | `clims_user`            | `clims_user`                 |
| `MYSQL_PASSWORD`      | Database user password   | `clims_password`        | *Secure Password*            |
| `DB_PORT`             | Host exposed DB port     | `3307`                  | `3306`                       |
| `DB_HOST`             | Database host container  | `mariadb`               | `mariadb`                    |
| `PORT`                | Backend container port   | `8080`                  | `8080`                       |
| `NEXT_PUBLIC_API_URL` | Public API base URL      | `http://localhost:8080` | `https://yourdomain.com/api` |
| `NGINX_PORT`          | Host HTTP listening port | `80`                    | `80`                         |

---

## 🚀 Quick Start Guide

### 1. Clone & Set Up Environment Files

Copy `dev.env.example` to create your active environment files:

```bash
cp dev.env.example dev.env
cp dev.env.example prod.env
```

### 2. Running in Development Mode

Run the stack using `dev.env` (maps MariaDB to host port `3307` to avoid port collisions with any local MySQL instances):

```bash
docker compose --env-file dev.env up --build -d
```

### 3. Running in Production Mode

Run the stack using `prod.env`:

```bash
docker compose --env-file prod.env up --build -d
```

### 4. Verify Active Containers

Check that all containers are running:

```bash
docker compose ps
```

---

## 🌐 Port Mappings & Access Points

| Service         | Internal Port | Host Port (Dev) | Access Endpoint         |
| :-------------- | :------------ | :-------------- | :---------------------- |
| **Nginx Proxy** | `80`          | `80`            | `http://localhost`      |
| **Frontend**    | `3000`        | Via Nginx       | `http://localhost/`     |
| **Backend API** | `8080`        | Via Nginx       | `http://localhost/api/` |
| **MariaDB**     | `3306`        | `3307`          | `localhost:3307`        |

---

## 🛠️ Database Management & Tooling

### Connecting via DBeaver / Database Clients

* **Host**: `localhost`
* **Port**: `3307` (Development)
* **Database**: `clims_db`
* **User**: `clims_user`
* **Password**: `clims_password`

### Database Logs & Inspection

To inspect container logs or database dump initialization state:

```bash
docker logs mariadb_clims
```

---

## 🧹 Maintenance & Commands

* **Stop the Stack**:
  ```bash
  docker compose down
  ```
* **Stop and Remove Volumes (Data Reset)**:
  ```bash
  docker compose down -v
  ```
* **Tail Container Logs**:
  ```bash
  docker compose logs -f
  ```
* **Rebuild Specific Service**:
  ```bash
  docker compose --env-file dev.env up --build -d backend
  ```