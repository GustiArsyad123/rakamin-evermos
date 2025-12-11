# Go Microservices E-commerce Example

This repository is a starter scaffold for a Go microservices project using MySQL, JWT and a Clean Architecture approach.

Included:

- `auth` service: implemented (register, login) — register auto-creates a `store` for the user.
- skeletons for `account`, `store`, `address`, `category`, `product`, `transaction` services
- `docker-compose.yml` with MySQL
- `sql/schema.sql` with tables and constraints

Run (development):

1. Start MySQL with Docker Compose:

```bash
docker compose up -d
```

2. Build and run the auth service (example):

```bash
cd cmd/auth
go run main.go
```

API endpoints (auth service)

- POST /api/v1/auth/register {name,email,phone,password} -> returns JWT
- POST /api/v1/auth/login {email,password} -> returns JWT

Notes

- This is a scaffold focusing on Clean Architecture structure and an implemented Auth service. You can extend other services in `internal/services/*` following the same pattern.

Environment files

- The repo includes example env files you can use locally or as templates for staging/production:
  - `.env.local` — development values (the default local MySQL is mapped to host port 3307 in `docker-compose.yml`).
  - `.env.testnet` — example test/staging values.
  - `.env.production` — example production values (replace secrets before deploy).

Usage examples

- Load local env before running a service locally:

```bash
# from repo root
source .env.local
DB_HOST=${DB_HOST} DB_PORT=${DB_PORT} DB_USER=${DB_USER} DB_PASS=${DB_PASS} DB_NAME=${DB_NAME} \
	go run cmd/auth/main.go
```

- If you run services with Docker Compose, the auth/product/transaction containers connect to the internal `db` service automatically. Use the `.env.*` files for local development or to create your deployment configs.
