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

## Database Schema

This section explains several safe and repeatable ways to run the SQL schema located at `sql/schema.sql` against your MySQL instance used by this project.

### 1. Quick local (mysql client) — when MySQL is reachable from your machine

- If using the bundled Docker DB with host port mapped to `3307` (default for this repo):

```bash
# from project root
mysql -h 127.0.0.1 -P 3307 -u ${DB_USER:-user} -p${DB_PASS:-password} ${DB_NAME:-ecommerce} < sql/schema.sql
```

Notes:

- This reads `${DB_USER}`, `${DB_PASS}`, `${DB_NAME}` from your environment; fallbacks are shown.
- If your password contains special characters, avoid passing it directly on the command line. Instead run `mysql -h 127.0.0.1 -P 3307 -u user -p ${DB_NAME} < sql/schema.sql` and enter the password when prompted.

### 2. Using Docker (container already running)

- Find the running MySQL container name:

```bash
docker ps --format "{{.Names}}\t{{.Image}}\t{{.Ports}}" | grep -i mysql
```

- Import using `docker exec` (good when the SQL file is on the host):

```bash
# replace <container> with the container name or id (e.g. myproject_db_1)
docker exec -i <container> sh -c 'mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME"' < sql/schema.sql
```

This runs the `mysql` client inside the container and pipes the host file into it. The command uses container environment variables if they're set inside the container; otherwise you can pass explicit credentials.

### 3. Using `docker-compose` (service name approach)

- If you run the stack with `docker-compose.yml`, the DB service is typically called `db` (check your compose file). You can run:

```bash
docker-compose exec db sh -c 'mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME"' < sql/schema.sql
```

Note: `docker-compose exec` runs the command inside an _already running_ container.

### 4. Secure ways to provide credentials

- Avoid placing raw passwords in commands or code.
- Use the project `.env.local` only for local development and do not commit secrets.
- For production or managed environments, prefer mounting Docker secrets and using `DB_PASS_FILE` (this project supports reading `DB_PASS_FILE` in the `internal/pkg/db` helper).

Example with password file (host):

```bash
# create a temporary file with password (secure with proper perms)
echo -n "mysecretpass" > /tmp/dbpass
# pass file path to container (example)
docker exec -i <container> sh -c 'export DB_PASS_FILE=/run/secrets/db_pass; export DB_USER=user; export DB_NAME=ecommerce; mysql -u"$DB_USER" -p"$(cat $DB_PASS_FILE)" "$DB_NAME"' < sql/schema.sql
```

### 5. Running seeds

- If you have a separate `sql/seed.sql` file, run it the same way after schema is applied:

```bash
mysql -h 127.0.0.1 -P 3307 -u user -ppassword ecommerce < sql/seed.sql
```

### 6. Troubleshooting

- "Access denied" — check username/password, and that the user has privileges on the database.
- "Can't connect to MySQL server" — check host/port and whether the container is running. Use `docker logs <container>` for errors.
- If you see SQL errors, inspect `sql/schema.sql` and verify your MySQL server version is compatible.

### 7. Example: full flow for local dev (start stack, apply schema)

```bash
# start db (docker-compose should be configured in this repo)
docker-compose up -d db
# wait for db to be healthy (or sleep a few seconds)
sleep 5
# import schema
mysql -h 127.0.0.1 -P 3307 -u ${DB_USER:-user} -p${DB_PASS:-password} ${DB_NAME:-ecommerce} < sql/schema.sql
```

## API Documentation

This section lists the main routes, methods, request body and query parameters for the implemented services (Auth, Product, Transaction). Use this as a quick reference.

Base URLs when running locally via Docker Compose:

- Auth: `http://localhost:8080`
- Product: `http://localhost:8081`
- Transaction: `http://localhost:8082`

### 1. Auth

- POST /api/v1/auth/register

  - Body (JSON): { "name": string, "email": string, "phone": string, "password": string }
  - Response: { "token": string }

- POST /api/v1/auth/login

  - Body (JSON): { "email": string, "password": string }
  - Response: { "token": string }

- POST /api/v1/auth/forgot-password

  - Body (JSON): { "email": string }
  - Response: { "reset_token": string, "expires_in": 3600 } (development only — token is returned; in production this would be emailed)

- POST /api/v1/auth/reset-password

  - Body (JSON): { "token": string, "password": string }
  - Response: 204 No Content

- POST /api/v1/auth/sso/google

  - Body (JSON): { "id_token": string }
  - Response: { "token": string, "expires_in": 3600 }
  - Notes: Accepts a Google ID token (JWT). In this development setup the server validates the token via `https://oauth2.googleapis.com/tokeninfo?id_token=<token>` and will create a new user automatically if the email does not exist. In production, verify audience (`aud`) against your Google OAuth Client ID and send the ID token to the backend securely.

- GET /api/v1/auth/users

  - Headers: `Authorization: Bearer <admin-token>`
  - Query params: `page` (int, default 1), `limit` (int, default 10, max 100), `search` (string, optional, filters name or email)
  - Response: { "data": [...], "pagination": { "page": int, "limit": int, "total": int } }
  - Notes: Admin-only

- GET /api/v1/users/:id

  - Headers: `Authorization: Bearer <token>`
  - Response: user object (owner or admin only)

### 2. Product

- POST /api/v1/products

  - Headers: `Authorization: Bearer <token>`
  - Body: multipart/form-data
    - fields: `name` (required), `price` (required), `description`, `stock`, `category_id`
    - file: `image` (optional)
  - Response: { "id": <product_id> }

- GET /api/v1/products

  - Query params: `page` (int), `limit` (int), `search` (string), `category_id`, `store_id`, `min_price`, `max_price`
  - Response: { "data": [...], "pagination": { "page":, "limit":, "total": } }

- GET /api/v1/products/:id

  - Response: product object

- PUT /api/v1/stores/:id

  - Headers: `Authorization: Bearer <token>`
  - Body (JSON): { "name": string }
  - Response: 204 No Content
  - Notes: Owner-only

### 3. Category

- GET /api/v1/categories

  - Query params: none
  - Response: { "data": [ { "id": int, "name": string, "created_at": string }, ... ] }

- GET /api/v1/categories/:id

  - Response: category object

- POST /api/v1/categories

  - Headers: `Authorization: Bearer <admin-token>`
  - Body (JSON): { "name": string }
  - Response: { "id": <category_id> }
  - Notes: Admin-only — use a user with `role='admin'`.

- PUT /api/v1/categories/:id

  - Headers: `Authorization: Bearer <admin-token>`
  - Body (JSON): { "name": string }
  - Response: 204 No Content
  - Notes: Admin-only

- DELETE /api/v1/categories/:id

  - Headers: `Authorization: Bearer <admin-token>`
  - Response: 204 No Content
  - Notes: Admin-only

### 4. Transaction

- POST /api/v1/transactions

  - Headers: `Authorization: Bearer <token>`
  - Body (JSON): { "items": [ { "product_id": int, "quantity": int }, ... ] }
  - Behavior: all items must be from the same store; creates `transactions` and `product_logs`, decrements product stock atomically.
  - Response: { "id": <transaction_id> }

- GET /api/v1/transactions

  - Headers: `Authorization: Bearer <token>`
  - Query params: `page`, `limit`
  - Response: list of user's transactions (paginated)

- GET /api/v1/transactions/:id
  - Headers: `Authorization: Bearer <token>`
  - Response: { "transaction": {...}, "logs": [...] }

Notes

- Category management is admin-only; use DB to mark a user as admin (see `sql/seed.sql`).
- Pagination and filtering params follow the patterns above.
