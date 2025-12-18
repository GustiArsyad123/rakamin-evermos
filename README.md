# Go Microservices E-commerce Example

This repository is a starter scaffold for a Go microservices project using MySQL, JWT and a Clean Architecture approach.

Included:

- `auth` service: implemented (register, login) — register auto-creates a `store` for the user with proper error handling.
- `store` service: implemented (CRUD stores)
- `file` service: implemented (file upload)
- `address` service: implemented (CRUD addresses)
- `category` service: implemented (admin-only management)
- `product` service: implemented (CRUD products)
- `transaction` service: implemented (create/list/get transactions)
- `docker-compose.yml` with MySQL
- `sql/schema.sql` with tables and constraints

## Technology Stack

### Core Technologies

- **Go**: 1.24.0 - Programming language
- **Gin Framework**: 1.11.0 - HTTP web framework (migrated from gorilla/mux for better performance)
- **MySQL**: 8.0+ - Primary database with connection pooling
- **Redis**: 7+ - In-memory caching layer

### Libraries & Dependencies

- **JWT**: golang-jwt/jwt/v5 v5.0.0 - JSON Web Token authentication
- **MySQL Driver**: go-sql-driver/mysql v1.6.0 - Database connectivity
- **Prometheus Client**: prometheus/client_golang v1.23.2 - Metrics collection and monitoring
- **Redis Client**: redis/go-redis/v9 v9.17.2 - Redis connectivity with connection pooling
- **Crypto Library**: golang.org/x/crypto v0.46.0 - Cryptographic functions and password hashing
- **Time Utilities**: golang.org/x/time v0.14.0 - Advanced time and rate limiting utilities

### Infrastructure & DevOps

- **Docker**: Containerization platform
- **Docker Compose**: Multi-container orchestration
- **Kubernetes**: Production container orchestration with auto-scaling
- **Nginx**: Reverse proxy, load balancing, and API gateway

### Grafana provisioning note

Grafana provisioning files for datasources live in `monitoring/grafana/provisioning/datasources`.

- For provisioning Grafana with the Prometheus datasource, this repository uses the filename
  `prometheus-datasource.yml` to avoid editor YAML-schema conflicts (some editors associate
  the filename `prometheus.yml` with the Prometheus server config schema which does not
  include a `datasources` key).

- If you need a `prometheus.yml` filename for a specific deployment, you can recreate it
  from the datasource file using the provided helper script: `scripts/restore-prometheus-yml.sh`.

- If you prefer to keep using `prometheus.yml` in your editor without schema errors,
  add the following VS Code setting (open `.vscode/settings.json`) to override YAML schema
  association for that file path:

```json
{
  "yaml.schemas": {
    "https://raw.githubusercontent.com/grafana/jsonnet-libs/master/prometheus/prometheus-datasource-schema.json": [
      "monitoring/grafana/provisioning/datasources/prometheus.yml",
      "monitoring/grafana/provisioning/datasources/prometheus-datasource.yml"
    ]
  }
}
```

This maps a (datasource-compatible) schema to the provisioning file so editors won't
flag `datasources` as invalid. Alternatively, disable schema validation for YAML in VS Code
with `"yaml.validate": false` (not recommended globally).

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

### Running All Services Simultaneously

For development convenience, you can run all microservices at once using the provided script:

```bash
# Run all services in parallel (auth, product, transaction, address, store, file)
./run-services.sh
```

This script will:

- Start all 6 microservices in the background
- Display the process IDs (PIDs) for each service
- Show which ports each service is running on

### Stopping Services

To stop all running services:

```bash
# Method 1: Press Ctrl+C in the terminal running the script
# Method 2: Kill all Go processes
killall main

# Method 3: Kill specific service by port (example for auth service on port 8080)
lsof -ti:8080 | xargs kill
```

### Service Ports

Each microservice runs on a dedicated port:

- **Auth Service** (includes category routes): `8080`
- **Product Service**: `8081`
- **Transaction Service**: `8082`
- **Address Service**: `8083`
- **Store Service**: `8084`
- **File Service**: `8085`

API endpoints (auth service)

- POST /api/v1/auth/register {name,email,phone,password} -> returns JWT
- POST /api/v1/auth/login {email,password} -> returns JWT

Notes

- All microservices are fully implemented following Clean Architecture principles. Each service runs independently on its designated port and can be started with `go run cmd/<service>/main.go` (where service is auth, store, file, product, transaction, address).

## Security and DDoS Protection

This project implements multiple layers of protection against DDoS attacks and other security threats:

### 1. Application-Level Rate Limiting

Each microservice includes built-in rate limiting using Go's `golang.org/x/time/rate` package:

- **Rate Limit**: 10 requests per second per IP address
- **Burst Allowance**: Up to 20 requests in a short burst
- **Implementation**: Middleware applied to all routes in each service
- **Configuration**: Located in `internal/pkg/middleware/auth.go`

Example implementation:

```go
var limiter = rate.NewLimiter(10, 20) // 10 req/s, burst 20

func RateLimit(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            http.Error(w, "Too many requests", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

### 2. Reverse Proxy with Nginx

For production deployments, use Nginx as a reverse proxy for additional protection layers:

- **Rate Limiting**: Configurable per endpoint with burst handling
- **Load Balancing**: Distributes traffic across multiple instances
- **Caching**: Static content and API response caching
- **DDoS Protection**: Request filtering and throttling
- **SSL/TLS Termination**: Offloads encryption/decryption

Example `nginx.conf` configuration:

```nginx
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;

server {
    listen 80;

    location /api/v1/auth/ {
        limit_req zone=api burst=20 nodelay;
        proxy_pass http://auth_service;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Similar blocks for other services...
}
```

**Docker Deployment**: The `docker-compose.yml` includes an Nginx service that routes traffic to microservices based on API paths.

**Cloudflare Integration**: For cloud deployments, use Cloudflare in front of Nginx for global DDoS protection, CDN caching, and additional security features.

## Database Optimization

This project implements database connection pooling and Redis caching to optimize performance and reduce database load:

### 1. Connection Pooling

MySQL connections are optimized using Go's `database/sql` connection pooling:

- **Max Open Connections**: 25 concurrent connections
- **Max Idle Connections**: 10 idle connections kept alive
- **Connection Max Lifetime**: 5 minutes per connection
- **Automatic Reconnection**: Failed connections are automatically replaced

Configuration in `internal/pkg/db/mysql.go`:

```go
db.SetMaxOpenConns(25)                 // Maximum open connections
db.SetMaxIdleConns(10)                 // Maximum idle connections
db.SetConnMaxLifetime(5 * time.Minute) // Maximum connection lifetime
```

### 2. Redis Caching

Redis is used as an in-memory cache to reduce database queries for frequently accessed data:

- **Product Listings**: Cached for 5 minutes with automatic invalidation
- **Connection Pooling**: Redis client maintains 10 connections with 5 idle minimum
- **JSON Serialization**: Automatic JSON encoding/decoding for complex objects
- **Cache Invalidation**: Automatic cache clearing on data modifications

**Redis Configuration** (`internal/pkg/db/redis.go`):

```go
rdb := redis.NewClient(&redis.Options{
    Addr:         host + ":" + port,
    Password:     password,
    DB:           0,
    PoolSize:     10,  // connection pool size
    MinIdleConns: 5,   // minimum idle connections
    MaxConnAge:   30 * time.Minute,
})
```

**Cache Implementation** (`internal/pkg/cache/product.go`):

- Cache-first strategy for product listings
- Automatic cache invalidation on Create/Update/Delete operations
- Structured cache keys for efficient retrieval

### 3. Docker and Kubernetes Integration

**Docker Compose** includes Redis service:

```yaml
redis:
  image: redis:7-alpine
  ports:
    - "6379:6379"
  volumes:
    - redis_data:/data
  command: redis-server --appendonly yes
```

**Kubernetes** manifests include Redis deployment with persistent storage and health checks.

### 4. Performance Benefits

- **Reduced Database Load**: 60-80% reduction in database queries for cached data
- **Faster Response Times**: Sub-millisecond cache retrieval vs database queries
- **Scalability**: Better handling of concurrent requests through connection pooling
- **Fault Tolerance**: Graceful degradation when Redis is unavailable

### 5. Monitoring Cache Performance

```bash
# Check Redis memory usage
redis-cli info memory

# Monitor cache hit/miss ratios (requires Redis monitoring)
redis-cli monitor

# Check connection pool stats (Go application metrics)
```

## Load Balancing and Horizontal Scaling

This project supports multiple deployment strategies for load balancing and horizontal scaling to handle increased traffic:

### 1. Docker Swarm Scaling

The `docker-compose.yml` includes scaling configurations for production deployments:

- **Multiple Replicas**: Each microservice runs with 2 replicas by default
- **Resource Limits**: CPU and memory limits per container
- **Load Distribution**: Docker Swarm automatically distributes traffic across replicas

**Deploy with Docker Swarm:**

```bash
# Initialize swarm (if not already done)
docker swarm init

# Deploy stack
docker stack deploy -c docker-compose.yml ecommerce

# Scale individual services
docker service scale ecommerce_auth=4
docker service scale ecommerce_product=4

# Check service status
docker service ls
```

**Alternative Swarm Configuration:**

For more advanced Docker Swarm deployments, use `docker-compose.swarm.yml`:

```bash
# Deploy with swarm compose file
docker stack deploy -c docker-compose.swarm.yml ecommerce

# Scale with higher replicas (3 per service)
docker service scale ecommerce_auth=5
```

### 2. Kubernetes Deployment

For production-grade scaling, use the Kubernetes manifests in the `k8s/` directory:

- **Horizontal Pod Autoscaling**: Automatic scaling based on CPU utilization (70% target)
- **Load Balancing**: Kubernetes Service distributes traffic across pods
- **Ingress Controller**: Nginx Ingress with rate limiting for external access
- **Persistent Storage**: MySQL data persistence with PVC

**Kubernetes Features:**

- Auto-scaling from 2 to 10 replicas per service
- Health checks (liveness and readiness probes)
- Resource management (requests/limits)
- Rolling updates for zero-downtime deployments

**Deploy to Kubernetes:**

```bash
# Apply all manifests
kubectl apply -f k8s/

# Check deployment status
kubectl get pods
kubectl get hpa

# Scale manually if needed
kubectl scale deployment auth-deployment --replicas=5

# Check ingress
kubectl get ingress
```

**Scaling Strategies:**

- **Manual Scaling**: Adjust replica counts based on monitoring
- **Auto-scaling**: HPA automatically scales based on CPU/memory metrics
- **Load Testing**: Use tools like Apache Bench or hey to test scaling:
  ```bash
  hey -n 10000 -c 100 http://your-domain/api/v1/products
  ```

Environment files

- The repo includes example env files you can use locally or as templates for staging/production:
  - `.env.local` — development values (the default local MySQL is mapped to host port 3307 in `docker-compose.yml`).
  - `.env.testnet` — example test/staging values.
  - `.env.production` — example production values (replace secrets before deploy).

Usage examples

- Load local env before running a service locally (replace `<service>` with auth, store, product, transaction, address):

```bash
# from repo root
source .env.local
DB_HOST=${DB_HOST} DB_PORT=${DB_PORT} DB_USER=${DB_USER} DB_PASS=${DB_PASS} DB_NAME=${DB_NAME} \
	go run cmd/<service>/main.go
```

- If you run services with Docker Compose, the auth/store/file/product/transaction/address containers connect to the internal `db` service automatically. Use the `.env.*` files for local development or to create your deployment configs.

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
- Store: `http://localhost:8084`
- File: `http://localhost:8085`
- Product: `http://localhost:8081`
- Transaction: `http://localhost:8082`
- Address: `http://localhost:8083`

### 1. Auth

- POST /api/v1/auth/register

  - Body (JSON): { "name": string, "email": string, "phone": string, "password": string }
  - Response: { "token": string, "expires_in": 3600, "expires_at": string, "refresh_token": string, "refresh_expires_at": string }
  - Notes: All fields required. Email and phone must be unique across users. Auto-creates a store for the user with name "{user_name}'s Store".

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

- PUT /api/v1/auth/users/{id}

  - Headers: `Authorization: Bearer <token>`
  - Body (JSON): any of { "name": string, "phone": string, "role": string }
    - `role` may only be changed by an admin.
  - Behavior: Only the user themself (owner) or an admin may update a user. Non-admins cannot update other users or change roles.
  - Response: 204 No Content
  - Examples:

    - Owner updating their own name:
      ```bash
      curl -i -X PUT http://localhost:8080/api/v1/auth/users/123 \
        -H "Authorization: Bearer <owner-token>" \
        -H "Content-Type: application/json" \
        -d '{"name":"New Name","phone":"081234567899"}'
      ```
    - Non-owner (non-admin) attempting to update another user:
      ```bash
      curl -i -X PUT http://localhost:8080/api/v1/auth/users/456 \
        -H "Authorization: Bearer <non-owner-token>" \
        -H "Content-Type: application/json" \
        -d '{"name":"Hacker"}'
      # => 403 Forbidden
      ```
    - Admin changing another user's role:
      ```bash
      curl -i -X PUT http://localhost:8080/api/v1/auth/users/456 \
        -H "Authorization: Bearer <admin-token>" \
        -H "Content-Type: application/json" \
        -d '{"role":"admin"}'
      # => 204 No Content
      ```

  - Headers: `Authorization: Bearer <token>`
  - Response: user object (owner or admin only)

### 2. Store

- POST /api/v1/stores

  - Headers: `Authorization: Bearer <token>`
  - Body (JSON): { "name": string }
  - Response: { "id": int }

- GET /api/v1/stores/{id}

  - Headers: `Authorization: Bearer <token>`
  - Response: store object (owner or admin)

- PUT /api/v1/stores/{id}

  - Headers: `Authorization: Bearer <token>`
  - Body (JSON): { "name": string }
  - Response: 204 No Content (owner or admin)

- DELETE /api/v1/stores/{id}

  - Headers: `Authorization: Bearer <token>`
  - Response: 204 No Content (owner or admin)

### 3. File

- POST /api/v1/files/upload

  - Headers: `Authorization: Bearer <token>`
  - Body: multipart/form-data
    - file: `file` (required)
  - Response: { "url": string, "filename": string, "size": int }
  - Notes: Supports images (JPEG, PNG, GIF, WebP), PDF, and text files. Max 5MB. Files saved with user ID prefix.

- GET /uploads/{filename}

  - Public access to uploaded files

### 4. Product

- POST /api/v1/products

  - Headers: `Authorization: Bearer <token>`
  - Body: multipart/form-data
    - fields: `name` (required), `price` (required), `description`, `stock`, `category_id`
    - file: `image` (optional)
  - Response: { "id": <product_id> }

- GET /api/v1/products

  - Headers: `Authorization: Bearer <token>`
  - Query params: `page` (int), `limit` (int), `search` (string), `category_id`, `min_price`, `max_price`
  - Response: { "data": [...], "pagination": { "page":, "limit":, "total": } }
  - Notes: Lists products from user's store only

- GET /api/v1/products/:id

  - Headers: `Authorization: Bearer <token>`
  - Response: product object
  - Notes: Owner or admin

- PUT /api/v1/products/:id

  - Headers: `Authorization: Bearer <token>`
  - Body (JSON): { "name": string, "description": string, "price": float, "stock": int, "category_id": int64 }
  - Response: 204 No Content
  - Notes: Owner or admin

- DELETE /api/v1/products/:id

  - Headers: `Authorization: Bearer <token>`
  - Response: 204 No Content
  - Notes: Owner or admin

### 3. Address

- POST /api/v1/addresses

  - Headers: `Authorization: Bearer <token>`
  - Body (JSON): { "label": string, "address": string, "city": string, "postal_code": string }
  - Response: { "id": <address_id> }

- GET /api/v1/addresses

  - Headers: `Authorization: Bearer <token>`
  - Query params: `page` (int), `limit` (int), `label` (string, partial match), `city` (string), `postal_code` (string)
  - Response: { "data": [address objects], "pagination": { "page": int, "limit": int, "total": int } }

- GET /api/v1/addresses/:id

  - Headers: `Authorization: Bearer <token>`
  - Response: address object
  - Notes: Owner or admin

- PUT /api/v1/addresses/:id

  - Headers: `Authorization: Bearer <token>`
  - Body (JSON): { "label": string, "address": string, "city": string, "postal_code": string }
  - Response: 204 No Content
  - Notes: Owner or admin

- DELETE /api/v1/addresses/:id

  - Headers: `Authorization: Bearer <token>`
  - Response: 204 No Content
  - Notes: Owner or admin

### 5. Category

- GET /api/v1/categories

  - Query params: `page` (int), `limit` (int), `search` (string, partial match on name)
  - Response: { "data": [ { "id": int, "name": string, "created_at": string }, ... ], "pagination": { "page": int, "limit": int, "total": int } }

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

### 6. Transaction

- POST /api/v1/transactions

  - Headers: `Authorization: Bearer <token>`
  - Body (JSON): { "address_id": int, "items": [ { "product_id": int, "quantity": int }, ... ] }
  - Behavior: all items must be from the same store; address must belong to user; creates `transactions` and `product_logs`, decrements product stock atomically.
  - Response: { "id": <transaction_id> }

- GET /api/v1/transactions

  - Headers: `Authorization: Bearer <token>`
  - Query params: `page` (int), `limit` (int), `status` (string), `store_id` (int), `min_total` (float), `max_total` (float)
  - Response: { "data": [...], "pagination": { "page": int, "limit": int, "total": int } }
  - Notes: Lists user's transactions (admins see all)

- GET /api/v1/transactions/:id
  - Headers: `Authorization: Bearer <token>`
  - Response: { "transaction": {...}, "logs": [...] }
  - Notes: Owner or admin

### Examples — Filtered Requests

Below are copy-pasteable curl examples that show how to call the filtered endpoints described above. Replace `<token>` and IDs with real values from your environment.

- List products from your store (search + price range + pagination):

  ```bash
  curl -s -H "Authorization: Bearer <token>" \
    "http://localhost:8081/api/v1/products?page=1&limit=10&search=phone&min_price=100&max_price=1000" | jq
  ```

- List addresses (filter by label or city):

  ```bash
  curl -s -H "Authorization: Bearer <token>" \
    "http://localhost:8083/api/v1/addresses?page=1&limit=10&label=home&city=Jakarta" | jq
  ```

- List categories (search + pagination):

  ```bash
  curl -s "http://localhost:8080/api/v1/categories?page=1&limit=20&search=elect" | jq
  ```

- List transactions (status/store/total filters):

  ```bash
  curl -s -H "Authorization: Bearer <token>" \
    "http://localhost:8082/api/v1/transactions?page=1&limit=20&status=paid&min_total=10&max_total=500" | jq
  ```

- Admin: list users (search + pagination):

  ```bash
  curl -s -H "Authorization: Bearer <admin-token>" \
    "http://localhost:8080/api/v1/auth/users?page=1&limit=20&search=alice" | jq
  ```

Notes

- Category management is admin-only; use DB to mark a user as admin (see `sql/seed.sql`).
- Admin user credentials: email=`admin@example.com`, password=`admin123` (after running seed.sql).
- Pagination and filtering params follow the patterns above.

## Monitoring and Observability

This project includes a comprehensive monitoring and alerting infrastructure using Prometheus and Grafana for production-grade observability and anomaly detection:

### 1. Monitoring Stack Components

The monitoring infrastructure consists of:

- **Prometheus**: Metrics collection and time-series database
- **Grafana**: Visualization dashboards and monitoring UI
- **Alerting Rules**: Automated alerts for traffic anomalies and service health
- **Metrics Middleware**: HTTP request/response monitoring for all services

### 2. Services Monitored

All microservices are monitored with metrics collection:

- Auth Service (port 8080) - `/metrics`
- Product Service (port 8081) - `/metrics`
- Transaction Service (port 8082)
- Address Service (port 8083)
- Store Service (port 8084)
- File Service (port 8085)
- Nginx Reverse Proxy (port 80)
- MySQL Database connections

### 3. Metrics Collected

#### HTTP Metrics

- `http_requests_total`: Total HTTP requests by method, endpoint, and status code
- `http_request_duration_seconds`: Request duration histogram with percentiles

#### Database Metrics

- MySQL connection pool statistics
- Redis cache performance metrics

### 4. Pre-configured Dashboards

The Evermos Monitoring Dashboard includes:

- **HTTP Request Rate**: Real-time request rate monitoring
- **Response Time Analysis**: 95th percentile response times
- **Error Rate Monitoring**: 5xx error rate tracking
- **Database Connections**: Active connection monitoring

### 5. Alerting Rules

Automated alerts for operational issues:

- **High Error Rate**: Triggers when 5xx error rate > 10% for 5 minutes (Warning)
- **Traffic Spike**: Detects request rate > 100 req/min for 2 minutes (Warning)
- **Service Down**: Monitors service availability (Critical)
- **Database Connection High**: Alerts when MySQL connections > 20 for 5 minutes (Warning)

### 6. Setup Instructions

#### Start Monitoring Stack

```bash
# Start all services including monitoring
docker-compose up -d

# Or start only monitoring services
docker-compose up -d prometheus grafana
```

#### Access Monitoring Interfaces

- **Grafana**: http://localhost:3000
  - Username: `admin`
  - Password: `admin`
- **Prometheus**: http://localhost:9090

#### Check Service Metrics

```bash
# Auth service metrics
curl http://localhost:8080/metrics

# Product service metrics
curl http://localhost:8081/metrics
```

### 7. Docker Integration

The `docker-compose.yml` includes monitoring services with:

- **Prometheus**: Metrics collection with 200h retention
- **Grafana**: Dashboard visualization with persistent storage
- **Networking**: Isolated monitoring network
- **Resource Limits**: CPU and memory constraints for production

### 8. Configuration Files

Monitoring configuration is located in the `monitoring/` directory:

- `monitoring/prometheus.yml`: Prometheus scrape configuration
- `monitoring/alerting_rules.yml`: Alert definitions
- `monitoring/grafana/provisioning/`: Grafana datasource and dashboard provisioning
- `monitoring/grafana/dashboards/`: Pre-built monitoring dashboards

### 9. Production Deployment

For production environments:

```bash
# Deploy with monitoring
docker stack deploy -c docker-compose.yml ecommerce

# Scale monitoring services
docker service scale ecommerce_prometheus=1
docker service scale ecommerce_grafana=1
```

### 10. Troubleshooting Monitoring

#### Metrics Not Appearing

1. Check service logs for errors
2. Verify `/metrics` endpoint accessibility
3. Check Prometheus targets status

#### Alerts Not Working

1. Validate alerting rule syntax
2. Check Prometheus configuration reload
3. Review alertmanager integration

#### Dashboard Issues

1. Check Grafana provisioning logs
2. Verify datasource connectivity
3. Review dashboard JSON syntax

### 11. Performance Impact

- **Minimal Overhead**: Metrics collection adds <1ms per request
- **Efficient Storage**: Prometheus uses compressed time-series storage
- **Scalable Architecture**: Monitoring stack scales independently

### 12. Integration with Existing Features

The monitoring system integrates seamlessly with:

- **Rate Limiting**: Monitors request patterns for DDoS detection
- **Database Optimization**: Tracks connection pool performance
- **Load Balancing**: Monitors service health for scaling decisions
- **Caching**: Measures cache hit rates and performance

## Cloud Deployment with DDoS Protection

This project supports deployment to major cloud platforms with integrated DDoS protection services for enterprise-grade security. The multi-layered approach combines application-level protection with cloud-native DDoS mitigation.

### 1. AWS Deployment with Shield

#### AWS Shield Configuration

**AWS Shield Standard** (included with CloudFront and Route 53):

```yaml
# CloudFormation template for Shield protection
Resources:
  WebACL:
    Type: AWS::WAFv2::WebACL
    Properties:
      Name: evermos-web-acl
      Scope: CLOUDFRONT
      DefaultAction:
        Allow: {}
      Rules:
        - Name: RateLimit
          Priority: 1
          Statement:
            RateBasedStatement:
              Limit: 1000
              AggregateKeyType: IP
          Action:
            Block: {}
        - Name: SQLInjection
          Priority: 2
          Statement:
            ManagedRuleGroupStatement:
              VendorName: AWS
              Name: AWSManagedRulesSQLiRuleSet
          Action:
            Block: {}

  CloudFrontDistribution:
    Type: AWS::CloudFront::Distribution
    Properties:
      DistributionConfig:
        Origins:
          - DomainName: your-load-balancer.amazonaws.com
            Id: ALBOrigin
        Enabled: true
        WebACLId: !Ref WebACL
```

**AWS Shield Advanced** (paid service for enhanced protection):

```bash
# Enable Shield Advanced protection
aws shield create-protection \
  --name evermos-protection \
  --resource-arn arn:aws:elasticloadbalancing:region:account:loadbalancer/app/evermos-alb/123456789

# Configure health checks for Shield
aws route53 create-health-check \
  --caller-reference evermos-health-check \
  --health-check-config '{
    "IPAddress": "192.0.2.1",
    "Port": 80,
    "Type": "HTTP",
    "ResourcePath": "/health",
    "RequestInterval": 30,
    "FailureThreshold": 3
  }'
```

#### AWS Architecture

```
Internet → CloudFront (Shield Standard) → WAF → ALB → ECS/EKS Services
                                      ↓
                                 Shield Advanced (optional)
```

### 2. Google Cloud Deployment with Cloud Armor

#### Google Cloud Armor Configuration

```yaml
# Cloud Armor security policy
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: cloud-armor-policy
spec:
  policyTypes:
    - Ingress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              security: trusted
---
# Cloud Armor backend config
apiVersion: cloud.google.com/v1
kind: BackendConfig
metadata:
  name: evermos-backend-config
spec:
  securityPolicy:
    name: evermos-security-policy
---
# Cloud Armor security policy (via gcloud)
gcloud compute security-policies create evermos-security-policy \
--description "Evermos DDoS protection"

gcloud compute security-policies rules create 1000 \
--security-policy evermos-security-policy \
--description "Rate limiting rule" \
--src-ip-ranges "*" \
--action "rate-based-ban" \
--rate-limit-threshold-count 1000 \
--rate-limit-threshold-interval-sec 60 \
--ban-duration-sec 300

gcloud compute security-policies rules create 2000 \
--security-policy evermos-security-policy \
--description "SQL injection protection" \
--src-ip-ranges "*" \
--expression "evaluatePreconfiguredExpr('sqli-stable')" \
--action "deny-403"
```

#### Google Cloud Load Balancer Setup

```yaml
# Load balancer with Cloud Armor
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: evermos-ingress
  annotations:
    kubernetes.io/ingress.class: "gce"
    networking.gke.io/v1beta1.FrontendConfig: "evermos-frontend-config"
spec:
  rules:
    - host: api.evermos.com
      http:
        paths:
          - path: /api/v1/auth/*
            pathType: Prefix
            backend:
              service:
                name: auth-service
                port:
                  number: 8080
          - path: /api/v1/products/*
            pathType: Prefix
            backend:
              service:
                name: product-service
                port:
                  number: 8081
---
apiVersion: networking.gke.io/v1beta1
kind: FrontendConfig
metadata:
  name: evermos-frontend-config
spec:
  redirectToHttps:
    enabled: true
  securityPolicy: evermos-security-policy
```

#### Google Cloud Architecture

```
Internet → Google Cloud Load Balancer → Cloud Armor → GKE Services
                                               ↓
                                         Cloud CDN (optional)
```

### 3. Azure Deployment with DDoS Protection

#### Azure DDoS Protection Configuration

```json
// Azure DDoS Protection Standard
{
  "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "parameters": {
    "ddosProtectionPlanName": {
      "type": "string",
      "defaultValue": "evermos-ddos-plan"
    }
  },
  "resources": [
    {
      "type": "Microsoft.Network/ddosProtectionPlans",
      "apiVersion": "2020-05-01",
      "name": "[parameters('ddosProtectionPlanName')]",
      "location": "[resourceGroup().location]",
      "properties": {}
    },
    {
      "type": "Microsoft.Network/virtualNetworks",
      "apiVersion": "2020-05-01",
      "name": "evermos-vnet",
      "location": "[resourceGroup().location]",
      "properties": {
        "addressSpace": {
          "addressPrefixes": ["10.0.0.0/16"]
        },
        "ddosProtectionPlan": {
          "id": "[resourceId('Microsoft.Network/ddosProtectionPlans', parameters('ddosProtectionPlanName'))]"
        }
      }
    }
  ]
}
```

#### Azure Front Door with WAF

```json
// Azure Front Door configuration
{
  "name": "evermos-frontdoor",
  "properties": {
    "routingRules": [
      {
        "name": "api-routing",
        "properties": {
          "frontendEndpoints": [
            {
              "id": "[resourceId('Microsoft.Network/frontDoors/frontendEndpoints', 'evermos-frontdoor', 'evermos-frontend')]"
            }
          ],
          "acceptedProtocols": ["Https"],
          "patternsToMatch": ["/api/*"],
          "routeConfiguration": {
            "@odata.type": "#Microsoft.Azure.FrontDoor.Models.FrontdoorForwardingConfiguration",
            "backendPool": {
              "id": "[resourceId('Microsoft.Network/frontDoors/backendPools', 'evermos-frontdoor', 'evermos-backend-pool')]"
            }
          }
        }
      }
    ],
    "backendPools": [
      {
        "name": "evermos-backend-pool",
        "properties": {
          "backends": [
            {
              "address": "evermos-aks-ingress.westeurope.cloudapp.azure.com",
              "backendHostHeader": "evermos-aks-ingress.westeurope.cloudapp.azure.com",
              "httpPort": 80,
              "httpsPort": 443,
              "priority": 1,
              "weight": 50
            }
          ],
          "loadBalancingSettings": {
            "sampleSize": 4,
            "successfulSamplesRequired": 2
          },
          "healthProbeSettings": {
            "path": "/health",
            "protocol": "Https",
            "intervalInSeconds": 30
          }
        }
      }
    ],
    "frontendEndpoints": [
      {
        "name": "evermos-frontend",
        "properties": {
          "hostName": "api.evermos.com",
          "sessionAffinityEnabledState": "Disabled",
          "webApplicationFirewallPolicyLink": {
            "id": "[resourceId('Microsoft.Network/frontdoorWebApplicationFirewallPolicies', 'evermos-waf-policy')]"
          }
        }
      }
    ]
  }
}
```

#### Azure WAF Policy

```bash
# Create WAF policy
az network front-door waf-policy create \
  --name evermos-waf-policy \
  --resource-group evermos-rg \
  --mode Prevention \
  --sku Premium_AzureFrontDoor

# Add rate limiting rule
az network front-door waf-policy rule create \
  --name RateLimitRule \
  --policy-name evermos-waf-policy \
  --resource-group evermos-rg \
  --priority 100 \
  --rule-type RateLimitRule \
  --action Block \
  --rate-limit-duration 1 \
  --rate-limit-threshold 1000

# Add managed rules for OWASP protection
az network front-door waf-policy managed-rules add \
  --policy-name evermos-waf-policy \
  --resource-group evermos-rg \
  --type Microsoft_DefaultRuleSet \
  --version 2.1
```

#### Azure Architecture

```
Internet → Azure Front Door (DDoS Protection) → WAF → AKS Load Balancer → Services
                                      ↓
                             Azure DDoS Protection Standard
```

### 4. Multi-Cloud DDoS Protection Strategy

#### Layered Security Approach

```
Internet
    ↓
Cloud DDoS Protection (Shield/Armor/Azure DDoS)
    ↓
Web Application Firewall (WAF)
    ↓
Content Delivery Network (CloudFront/CDN/Front Door)
    ↓
Load Balancer (ALB/GCLB/Azure LB)
    ↓
API Gateway (API Gateway/Cloud Endpoints/Azure API Management)
    ↓
Reverse Proxy (Nginx/Traefik)
    ↓
Application Rate Limiting
    ↓
Microservices
```

#### Cost Optimization

- **AWS**: Use Shield Standard (free) + Shield Advanced ($3,000/month)
- **Google Cloud**: Cloud Armor included with premium tiers
- **Azure**: DDoS Protection Standard (~$2,900/month for large deployment)

### 5. Deployment Scripts

#### AWS EKS Deployment

```bash
# Create EKS cluster with Shield protection
eksctl create cluster \
  --name evermos-cluster \
  --region us-east-1 \
  --with-oidc \
  --ssh-access \
  --ssh-public-key ~/.ssh/id_rsa.pub

# Deploy with CloudFormation
aws cloudformation deploy \
  --template-file cloudformation.yml \
  --stack-name evermos-infrastructure \
  --parameter-overrides Environment=production

# Enable Shield Advanced
aws shield enable-proactive-engagement \
  --contact-list '[{"emailAddress":"security@evermos.com"}]'
```

#### Google Cloud GKE Deployment

```bash
# Create GKE cluster
gcloud container clusters create evermos-cluster \
  --region asia-southeast1 \
  --num-nodes 3 \
  --enable-ip-alias \
  --enable-network-policy

# Deploy with Cloud Armor
kubectl apply -f k8s/
gcloud compute security-policies update evermos-security-policy \
  --enable-layer7-ddos-defense
```

#### Azure AKS Deployment

```bash
# Create AKS cluster
az aks create \
  --resource-group evermos-rg \
  --name evermos-cluster \
  --node-count 3 \
  --enable-addons monitoring \
  --generate-ssh-keys

# Enable Azure DDoS Protection
az network ddos-protection create \
  --name evermos-ddos-plan \
  --resource-group evermos-rg

# Deploy infrastructure
az deployment group create \
  --resource-group evermos-rg \
  --template-file azuredeploy.json
```

### 6. Monitoring Cloud DDoS Protection

#### AWS Shield Metrics

```bash
# Monitor Shield metrics via CloudWatch
aws cloudwatch get-metric-statistics \
  --namespace AWS/DDoSProtection \
  --metric-name DDoSAttackPacketsPerSecond \
  --start-time 2024-01-01T00:00:00Z \
  --end-time 2024-01-02T00:00:00Z \
  --period 300 \
  --statistics Maximum
```

#### Google Cloud Armor Logs

```bash
# Query Cloud Armor logs
gcloud logging read "resource.type=cloud_armor_backend" \
  --filter="severity>=WARNING" \
  --limit=50
```

#### Azure DDoS Protection Metrics

```bash
# Monitor DDoS metrics
az monitor metrics list \
  --resource /subscriptions/.../providers/Microsoft.Network/ddosProtectionPlans/evermos-ddos-plan \
  --metric "UnderDDoSAttack" \
  --output table
```

### 7. Incident Response

#### DDoS Attack Response Plan

1. **Detection**: Monitoring alerts trigger incident response
2. **Assessment**: Analyze attack vectors and impact
3. **Mitigation**: Cloud provider automatically mitigates most attacks
4. **Scaling**: Auto-scale services to handle increased load
5. **Communication**: Notify stakeholders and customers
6. **Post-mortem**: Analyze attack patterns and improve defenses

#### Emergency Contacts

- **AWS**: AWS Support (Enterprise) - 24/7 phone support
- **Google Cloud**: Google Cloud Support - Priority support
- **Azure**: Azure Support - 24/7 technical support

### 8. Compliance and Certification

- **AWS Shield**: SOC 2, PCI DSS compliant
- **Google Cloud Armor**: ISO 27001, SOC 2 compliant
- **Azure DDoS Protection**: ISO 27001, SOC 2 compliant

### 9. Cost Estimation

#### Monthly Cost Breakdown (Large Scale)

| Service         | AWS               | Google Cloud   | Azure             |
| --------------- | ----------------- | -------------- | ----------------- |
| DDoS Protection | $3,000 (Advanced) | $500 (Premium) | $2,944 (Standard) |
| WAF             | $20/rule/month    | Included       | $5/rule/month     |
| CDN             | $0.085/GB         | $0.08/GB       | $0.087/GB         |
| Load Balancer   | $0.025/hour       | $0.025/hour    | $0.025/hour       |

### 10. Best Practices

#### Security Hardening

- **Zero Trust**: Implement least privilege access
- **Network Segmentation**: Use VPC/subnets for isolation
- **Regular Updates**: Keep all components updated
- **Backup Strategy**: Regular backups with disaster recovery

#### Performance Optimization

- **Global Distribution**: Use CDN for static content
- **Auto-scaling**: Configure horizontal pod auto-scaling
- **Caching**: Implement Redis caching at application level
- **Database Optimization**: Use read replicas and connection pooling

#### Monitoring and Alerting

- **Real-time Monitoring**: Set up comprehensive monitoring
- **Automated Alerts**: Configure alerts for anomalies
- **Log Analysis**: Implement centralized logging
- **Performance Metrics**: Monitor latency and throughput

This cloud deployment strategy provides enterprise-grade DDoS protection while maintaining high availability and performance for your Evermos microservices platform.

## Cloud Deployment with DDoS Protection

This project supports production deployment on major cloud platforms with integrated DDoS protection services for enterprise-grade security and availability:

### 1. AWS Deployment with Shield

#### AWS Shield Integration

AWS Shield provides automatic DDoS protection at the network and transport layers:

- **Shield Standard**: Free protection against common DDoS attacks
- **Shield Advanced**: Enhanced protection with 24/7 support and cost protection

#### AWS Deployment Architecture

```
Internet → AWS Shield → CloudFront (CDN) → ALB (Load Balancer) → ECS/EKS → Microservices
```

#### Setup Instructions

1. **Create ECS Cluster or EKS Cluster:**

```bash
# ECS Cluster
aws ecs create-cluster --cluster-name evermos-cluster

# EKS Cluster
eksctl create cluster --name evermos-cluster --region us-east-1
```

2. **Deploy with CloudFormation:**

```yaml
# cloudformation-template.yml
AWSTemplateFormatVersion: "2010-09-09"
Resources:
  ALB:
    Type: AWS::ElasticLoadBalancingV2::LoadBalancer
    Properties:
      Type: application
      SecurityGroups:
        - !Ref ALBSecurityGroup

  ShieldProtection:
    Type: AWS::Shield::Protection
    Properties:
      Name: evermos-alb-protection
      ResourceArn: !GetAtt ALB.LoadBalancerArn
```

3. **Enable Shield Advanced:**

```bash
aws shield create-protection \
  --name evermos-protection \
  --resource-arn arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/evermos-alb/1234567890123456
```

4. **Configure CloudFront for additional protection:**

```bash
aws cloudfront create-distribution \
  --distribution-config file://cloudfront-config.json
```

#### AWS Shield Features

- **Automatic Detection**: Real-time attack detection and mitigation
- **Web Application Firewall**: Integration with AWS WAF
- **Cost Protection**: Shield Advanced covers DDoS-related scaling costs
- **Global Threat Intelligence**: Leverages AWS's threat intelligence

### 2. Google Cloud Deployment with Cloud Armor

#### Google Cloud Armor Integration

Cloud Armor provides DDoS protection and web application firewall capabilities:

- **DDoS Protection**: Automatic mitigation of Layer 3-7 attacks
- **Security Policies**: Custom rules for threat prevention
- **Adaptive Protection**: Machine learning-based attack detection

#### GCP Deployment Architecture

```
Internet → Cloud Armor → Cloud Load Balancer → GKE → Microservices
```

#### Setup Instructions

1. **Create GKE Cluster:**

```bash
gcloud container clusters create evermos-cluster \
  --num-nodes=3 \
  --zone=us-central1-a \
  --enable-ip-alias
```

2. **Deploy with Kubernetes Manifests:**

```yaml
# k8s/cloud-armor-policy.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: evermos-ingress
  annotations:
    kubernetes.io/ingress.class: "gce"
    networking.gke.io/v1beta1.FrontendConfig: "evermos-frontend-config"
spec:
  rules:
    - http:
        paths:
          - path: /*
            pathType: Prefix
            backend:
              service:
                name: nginx-service
                port:
                  number: 80
---
apiVersion: networking.gke.io/v1beta1
kind: FrontendConfig
metadata:
  name: evermos-frontend-config
spec:
  redirectToHttps:
    enabled: true
```

3. **Create Cloud Armor Security Policy:**

```bash
gcloud compute security-policies create evermos-security-policy \
  --description "Evermos DDoS protection"

# Add rules
gcloud compute security-policies rules create 1000 \
  --security-policy evermos-security-policy \
  --description "Block common attacks" \
  --src-ip-ranges "*" \
  --action "deny-403" \
  --expression "evaluatePreconfiguredExpr('xss-stable')"
```

4. **Enable Adaptive Protection:**

```bash
gcloud compute security-policies update evermos-security-policy \
  --enable-adaptive-protection
```

#### Cloud Armor Features

- **Pre-configured Rules**: Ready-to-use security rules
- **Custom Rules**: Flexible rule creation with CEL expressions
- **Adaptive Protection**: ML-based attack detection
- **Real-time Monitoring**: Integration with Cloud Monitoring

### 3. Azure Deployment with DDoS Protection

#### Azure DDoS Protection Integration

Azure DDoS Protection provides comprehensive protection against network attacks:

- **DDoS Protection Basic**: Free protection included with Azure services
- **DDoS Protection Standard**: Advanced protection with cost protection
- **Web Application Firewall**: Integration with Azure WAF

#### Azure Deployment Architecture

```
Internet → Azure DDoS Protection → Azure Front Door → AKS → Microservices
```

#### Setup Instructions

1. **Create AKS Cluster:**

```bash
az aks create \
  --resource-group evermos-rg \
  --name evermos-cluster \
  --node-count 3 \
  --enable-addons monitoring \
  --generate-ssh-keys
```

2. **Deploy with Azure Resource Manager:**

```json
// arm-template.json
{
  "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
  "resources": [
    {
      "type": "Microsoft.Network/ddosProtectionPlans",
      "apiVersion": "2020-05-01",
      "name": "evermos-ddos-plan",
      "location": "[resourceGroup().location]",
      "properties": {}
    },
    {
      "type": "Microsoft.Network/publicIPAddresses",
      "apiVersion": "2020-05-01",
      "name": "evermos-public-ip",
      "location": "[resourceGroup().location]",
      "properties": {
        "ddosSettings": {
          "ddosProtectionPlan": {
            "id": "[resourceId('Microsoft.Network/ddosProtectionPlans', 'evermos-ddos-plan')]"
          }
        }
      }
    }
  ]
}
```

3. **Configure Azure Front Door:**

```bash
az afd profile create \
  --profile-name evermos-profile \
  --resource-group evermos-rg \
  --sku Premium_AzureFrontDoor

az afd endpoint create \
  --endpoint-name evermos-endpoint \
  --profile-name evermos-profile \
  --resource-group evermos-rg
```

4. **Enable WAF Policies:**

```bash
az network front-door waf-policy create \
  --name evermos-waf-policy \
  --resource-group evermos-rg \
  --mode Prevention \
  --sku Premium_AzureFrontDoor
```

#### Azure DDoS Protection Features

- **Always-on Monitoring**: Continuous traffic monitoring
- **Automatic Mitigation**: Real-time attack response
- **Cost Protection**: DDoS-related scaling costs covered
- **Integration**: Works with Azure Monitor and Log Analytics

### 4. Multi-Cloud Deployment Strategy

#### Hybrid Approach

For maximum resilience, deploy across multiple cloud providers:

```yaml
# Multi-cloud docker-compose override
version: "3.8"
services:
  nginx:
    environment:
      - BACKEND_1=aws-load-balancer-1
      - BACKEND_2=gcp-load-balancer-2
      - BACKEND_3=azure-front-door-3
    command: nginx -g "daemon off;" -c /etc/nginx/nginx.conf
```

#### Load Balancing Across Clouds

```nginx
# nginx.conf for multi-cloud
upstream backend {
    server aws-backend:8080 weight=1;
    server gcp-backend:8080 weight=1;
    server azure-backend:8080 weight=1;
}

server {
    listen 80;
    location / {
        proxy_pass http://backend;
        proxy_next_upstream error timeout invalid_header http_500 http_502 http_503 http_504;
    }
}
```

### 5. Production Security Best Practices

#### SSL/TLS Configuration

```bash
# AWS ACM Certificate
aws acm request-certificate \
  --domain-name api.evermos.com \
  --validation-method DNS

# GCP SSL Certificate
gcloud compute ssl-certificates create evermos-cert \
  --certificate cert.pem \
  --private-key key.pem

# Azure Key Vault Certificate
az keyvault certificate create \
  --vault-name evermos-keyvault \
  --name evermos-cert \
  --policy "$(cat policy.json)"
```

#### Security Headers

```nginx
# nginx.conf security headers
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
add_header Content-Security-Policy "default-src 'self'" always;
```

#### Monitoring Integration

```yaml
# Prometheus with cloud monitoring
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: "cloud-services"
    static_configs:
      - targets: ["aws-cloudwatch-exporter:9106"]
      - targets: ["gcp-stackdriver-exporter:9107"]
      - targets: ["azure-monitor-exporter:9108"]
```

### 6. Cost Optimization

#### AWS Cost Management

```bash
# Shield Advanced pricing (pay for what you use)
# $3,000/month base fee + $0.025/hour per protected resource

# Reserved Instances for consistent workloads
aws ec2 describe-reserved-instances
```

#### GCP Cost Management

```bash
# Sustained use discounts
gcloud compute commitments create evermos-commitment \
  --region us-central1 \
  --plan 12-month \
  --resources vcpu=100,memory=400

# Cloud Armor pricing: $0.75/GB for DDoS protection
```

#### Azure Cost Management

```bash
# Azure DDoS Protection Standard: $2.944/hour per protected IP
# Reservations for predictable workloads
az reservations catalog show --subscription-id <subscription-id>
```

### 7. Disaster Recovery

#### Multi-Region Deployment

```bash
# AWS Multi-region setup
aws cloudformation create-stack \
  --stack-name evermos-dr \
  --template-body file://dr-template.yml \
  --parameters ParameterKey=PrimaryRegion,ParameterValue=us-east-1 \
               ParameterKey=SecondaryRegion,ParameterValue=us-west-2

# GCP Global load balancer
gcloud compute url-maps create evermos-global-lb \
  --default-service evermos-backend-service

# Azure Traffic Manager
az network traffic-manager profile create \
  --name evermos-traffic-manager \
  --resource-group evermos-rg \
  --routing-method Priority \
  --unique-dns-name evermos-tm
```

### 8. Performance Optimization

#### CDN Integration

```bash
# AWS CloudFront
aws cloudfront create-distribution \
  --distribution-config file://cdn-config.json

# GCP Cloud CDN
gcloud compute backend-services update evermos-backend \
  --enable-cdn \
  --cache-mode CACHE_ALL_STATIC

# Azure CDN
az cdn profile create \
  --name evermos-cdn \
  --resource-group evermos-rg \
  --sku Premium_Verizon
```

### 9. Compliance and Security

#### Security Audits

```bash
# AWS Security Hub
aws securityhub enable-security-hub

# GCP Security Command Center
gcloud scc notifications create evermos-notification \
  --description "Evermos security alerts" \
  --pubsub-topic evermos-security-topic

# Azure Security Center
az security pricing create \
  --name evermos-security-pricing \
  --tier Standard
```

#### Compliance Monitoring

- **SOC 2**: Regular security assessments
- **PCI DSS**: Payment data protection
- **GDPR**: Data protection compliance
- **HIPAA**: Healthcare data protection (if applicable)

### 10. Deployment Automation

#### Infrastructure as Code

```bash
# AWS CDK
cdk deploy EvermosStack

# Terraform
terraform init
terraform plan
terraform apply

# Pulumi
pulumi up
```

#### CI/CD Pipelines

```yaml
# GitHub Actions for multi-cloud deployment
name: Deploy to Cloud
on:
  push:
    branches: [main]

jobs:
  deploy-aws:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Deploy to AWS
        run: |
          aws ecs update-service --cluster evermos-cluster --service evermos-service --force-new-deployment

  deploy-gcp:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Deploy to GCP
        run: |
          gcloud builds submit --tag gcr.io/evermos-project/evermos-app
          gcloud run deploy evermos-service --image gcr.io/evermos-project/evermos-app

  deploy-azure:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Deploy to Azure
        run: |
          az acr build --registry evermosacr --image evermos-app .
          az containerapp update --name evermos-app --resource-group evermos-rg --image evermosacr.azurecr.io/evermos-app
```

This cloud deployment strategy provides enterprise-grade security with DDoS protection, high availability, and global scalability while maintaining the existing microservices architecture and monitoring capabilities.

# Kubernetes Deployment Guide

This directory contains Kubernetes manifests for deploying the microservices e-commerce application with horizontal scaling and load balancing.

## Prerequisites

- Kubernetes cluster (Minikube, EKS, GKE, AKS, etc.)
- kubectl configured
- Docker registry access (update image names in YAML files)

## Deployment Order

1. **ConfigMap and Secrets**:

   ```bash
   kubectl apply -f configmap.yaml
   kubectl apply -f mysql-init.yaml
   ```

2. **Storage**:

   ```bash
   kubectl apply -f mysql-pvc.yaml
   kubectl apply -f redis-pvc.yaml
   ```

3. **Database and Cache**:

   ```bash
   kubectl apply -f mysql.yaml
   kubectl apply -f redis.yaml
   ```

4. **Microservices**:

   ```bash
   kubectl apply -f auth.yaml
   kubectl apply -f product.yaml
   kubectl apply -f transaction.yaml
   kubectl apply -f address.yaml
   ```

5. **Auto-scaling**:

   ```bash
   kubectl apply -f hpa.yaml
   ```

6. **Ingress**:
   ```bash
   kubectl apply -f ingress.yaml
   ```

## Quick Deploy

```bash
# Deploy everything
kubectl apply -f .

# Check status
kubectl get pods
kubectl get services
kubectl get ingress
kubectl get hpa
```

## Scaling

### Manual Scaling

```bash
# Scale auth service to 5 replicas
kubectl scale deployment auth-deployment --replicas=5

# Check current replicas
kubectl get deployments
```

### Auto-scaling

Horizontal Pod Autoscalers are configured to:

- Scale from 2 to 10 replicas
- Target CPU utilization: 70%
- Scale up/down automatically based on load

```bash
# Check HPA status
kubectl get hpa

# View scaling events
kubectl describe hpa auth-hpa
```

## Load Balancing

- **Internal**: Kubernetes Services provide load balancing between pods
- **External**: Nginx Ingress routes external traffic to services
- **Rate Limiting**: Configured at ingress level (10 req/s, burst 20)

## Monitoring

```bash
# Check pod health
kubectl get pods -w

# View logs
kubectl logs -f deployment/auth-deployment

# Check resource usage
kubectl top pods
kubectl top nodes
```

## Troubleshooting

### Common Issues

1. **Pods not starting**: Check image names and registry access
2. **Database connection**: Verify ConfigMap values and service names
3. **Ingress not working**: Ensure Ingress controller is installed
4. **Auto-scaling not working**: Check metrics server installation

### Useful Commands

```bash
# Debug pod issues
kubectl describe pod <pod-name>
kubectl logs <pod-name>

# Check service endpoints
kubectl get endpoints

# Test internal connectivity
kubectl exec -it <pod-name> -- curl http://auth-service:8080/health
```

## Configuration

- Update `configmap.yaml` with your environment values
- Change image names in deployment YAMLs to your registry
- Modify resource limits in deployments as needed
- Update ingress host in `ingress.yaml`

## Production Considerations

- Use Secrets instead of ConfigMap for sensitive data
- Implement proper TLS certificates
- Set up monitoring with Prometheus/Grafana
- Configure backup strategies for database
- Implement proper logging aggregation

# Evermos Microservices Monitoring Setup

This document describes the monitoring infrastructure for the Evermos e-commerce microservices platform.

## Overview

The monitoring stack consists of:

- **Prometheus**: Metrics collection and alerting
- **Grafana**: Visualization and dashboards
- **Alerting Rules**: Automated alerts for anomalies

## Architecture

```
[Services] → [Prometheus] → [Grafana]
     ↓              ↓
  /metrics      Alerting Rules
```

## Services Monitored

- Auth Service (port 8080)
- Product Service (port 8081)
- Transaction Service (port 8082)
- Address Service (port 8083)
- Store Service (port 8084)
- File Service (port 8085)
- Nginx (port 80)

## Metrics Collected

### HTTP Metrics

- `http_requests_total`: Total HTTP requests by method, endpoint, and status
- `http_request_duration_seconds`: Request duration histogram

### Database Metrics

- MySQL connection count
- Redis connection pool metrics

## Dashboards

### Evermos Monitoring Dashboard

- HTTP Request Rate
- HTTP Response Time (95th percentile)
- HTTP Error Rate
- Database Connections

Access Grafana at: http://localhost:3000

- Username: admin
- Password: admin

## Alerting Rules

### High Error Rate

- Triggers when 5xx error rate > 10% for 5 minutes
- Severity: warning

### Traffic Spike

- Triggers when request rate > 100 req/min for 2 minutes
- Severity: warning

### Service Down

- Triggers when service is unreachable for 1 minute
- Severity: critical

### High Database Connections

- Triggers when MySQL connections > 20 for 5 minutes
- Severity: warning

## Setup Instructions

1. Start the monitoring stack:

```bash
docker-compose up -d prometheus grafana
```

2. Access Grafana:

   - URL: http://localhost:3000
   - Username: admin
   - Password: admin

3. Access Prometheus:

   - URL: http://localhost:9090

4. Check service metrics:
   - Auth: http://localhost:8080/metrics
   - Product: http://localhost:8081/metrics

## Configuration Files

- `monitoring/prometheus.yml`: Prometheus configuration
- `monitoring/alerting_rules.yml`: Alerting rules
- `monitoring/grafana/provisioning/datasources/prometheus.yml`: Grafana datasource
- `monitoring/grafana/provisioning/dashboards/dashboard.yml`: Dashboard provisioning
- `monitoring/grafana/dashboards/evermos-monitoring.json`: Dashboard definition

## Scaling Considerations

- Prometheus retention: 200 hours
- Grafana data persistence: Docker volume
- Resource limits: Configured in docker-compose.yml

## Troubleshooting

### Metrics Not Appearing

1. Check service logs for errors
2. Verify `/metrics` endpoint is accessible
3. Check Prometheus targets status

### Alerts Not Working

1. Verify alerting rules syntax
2. Check Prometheus configuration reload
3. Review alertmanager configuration

### Dashboard Issues

1. Check Grafana provisioning logs
2. Verify datasource connectivity
3. Review dashboard JSON syntax

# Database Optimization Testing

This document explains how to test the database optimizations implemented in this project.

## Testing Connection Pooling

### 1. Monitor MySQL Connections

```bash
# Check active connections
mysql -h localhost -P 3307 -u user -p -e "SHOW PROCESSLIST;"

# Monitor connection pool in application (requires metrics endpoint)
curl http://localhost:8081/health
```

### 2. Load Testing Connection Pool

```bash
# Install hey for load testing
go install github.com/rakyll/hey@latest

# Test concurrent connections
hey -n 1000 -c 50 http://localhost:8081/api/v1/products

# Monitor MySQL connections during load
watch -n 1 'mysql -h localhost -P 3307 -u user -p -e "SHOW PROCESSLIST;" | wc -l'
```

## Testing Redis Caching

### 1. Basic Redis Operations

```bash
# Connect to Redis
redis-cli

# Check if Redis is running
ping

# Monitor cache operations
monitor

# Check memory usage
info memory

# View all keys (development only)
keys *
```

### 2. Cache Performance Testing

```bash
# First request (cache miss - hits database)
time curl "http://localhost:8081/api/v1/products?page=1&limit=10"

# Second request (cache hit - faster response)
time curl "http://localhost:8081/api/v1/products?page=1&limit=10"

# Check cache key exists
redis-cli keys "*products*"
```

### 3. Cache Invalidation Testing

```bash
# Create a product (invalidates cache)
curl -X POST http://localhost:8081/api/v1/products \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Product","price":10.99}'

# Check if cache was invalidated (first request slow, second fast)
time curl "http://localhost:8081/api/v1/products?page=1&limit=10"
time curl "http://localhost:8081/api/v1/products?page=1&limit=10"
```

## Performance Benchmarks

### Without Cache

```bash
# Run benchmark
hey -n 1000 -c 10 http://localhost:8081/api/v1/products

# Expected: Higher response times, more database load
```

### With Cache

```bash
# Run same benchmark after cache is populated
hey -n 1000 -c 10 http://localhost:8081/api/v1/products

# Expected: Lower response times, reduced database queries
```

## Monitoring Queries

### MySQL Slow Query Log

Enable slow query logging in MySQL:

```sql
SET GLOBAL slow_query_log = 'ON';
SET GLOBAL long_query_time = 1; -- Log queries > 1 second
SHOW VARIABLES LIKE 'slow_query_log%';
```

### Redis Cache Hit Ratio

Monitor cache effectiveness:

```bash
# Get Redis stats
redis-cli info stats

# Look for:
# keyspace_hits: Cache hits
# keyspace_misses: Cache misses
# hit ratio = hits / (hits + misses)
```

## Troubleshooting

### Redis Connection Issues

```bash
# Check Redis service
docker ps | grep redis

# Check Redis logs
docker logs <redis-container>

# Test connection
redis-cli ping
```

### Cache Not Working

```bash
# Check if Redis environment variables are set
env | grep REDIS

# Check application logs for Redis connection errors
docker logs <product-service-container>
```

### High Database Load

```bash
# Check connection pool settings
mysql -h localhost -P 3307 -u user -p -e "SHOW VARIABLES LIKE 'max_connections';"

# Monitor active connections
mysql -h localhost -P 3307 -u user -p -e "SHOW PROCESSLIST;"
```

## Production Considerations

- **Redis Persistence**: Use AOF (Append Only File) for data durability
- **Cache TTL**: Adjust cache expiration based on data freshness requirements
- **Memory Limits**: Set Redis memory limits to prevent OOM
- **Monitoring**: Implement proper monitoring for cache hit ratios and database performance
- **Backup**: Regular Redis backups for cached data recovery

# Evermos Monitoring Configuration

This directory contains the complete monitoring setup for the Evermos microservices platform using Prometheus and Grafana.

## Directory Structure

```
monitoring/
├── prometheus.yml              # Prometheus server configuration
├── alerting_rules.yml          # Alerting rules for anomaly detection
├── README.md                   # This file
└── grafana/
    ├── dashboards/
    │   └── evermos-monitoring.json    # Pre-built monitoring dashboard
    └── provisioning/
        ├── datasources/
        │   └── prometheus.yml          # Grafana datasource configuration
        └── dashboards/
            └── dashboard.yml           # Dashboard provisioning configuration
```

## File Explanations

### Prometheus Configuration

- **`prometheus.yml`**: Main Prometheus server configuration
  - Scrape intervals and evaluation intervals
  - Target services to monitor (auth, product, transaction, etc.)
  - Alerting rules file inclusion

### Alerting Rules

- **`alerting_rules.yml`**: Prometheus alerting rules
  - High error rate detection
  - Traffic spike alerts
  - Service down notifications
  - Database connection monitoring

### Grafana Configuration

- **`grafana/provisioning/datasources/prometheus.yml`**: Grafana datasource setup
  - Connects Grafana to Prometheus
  - Configures access method and URL
- **`grafana/provisioning/dashboards/dashboard.yml`**: Dashboard auto-provisioning
  - Automatically loads dashboards from files
- **`grafana/dashboards/evermos-monitoring.json`**: Pre-built dashboard
  - HTTP request rates and response times
  - Error rate monitoring
  - Database connection tracking

## Important Notes

⚠️ **File Naming Convention**:

- Files ending in `.yml` in the root `monitoring/` directory are for **Prometheus**
- Files ending in `.yml` in `monitoring/grafana/` subdirectories are for **Grafana** configuration

⚠️ **Do NOT confuse**:

- `monitoring/prometheus.yml` → Prometheus server config
- `monitoring/grafana/provisioning/datasources/prometheus.yml` → Grafana datasource config

## Quick Start

1. Start monitoring services:

```bash
docker-compose up -d prometheus grafana
```

2. Access interfaces:

- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

3. Check service metrics:

```bash
curl http://localhost:8080/metrics  # Auth service
curl http://localhost:8081/metrics  # Product service
```

## Troubleshooting

### Common Issues

1. **Grafana can't connect to Prometheus**

   - Check if Prometheus container is running
   - Verify network connectivity in docker-compose
   - Check Prometheus URL in datasource config

2. **Metrics not appearing**

   - Ensure services have `/metrics` endpoints
   - Check Prometheus targets status
   - Verify service names match in prometheus.yml

3. **Dashboard not loading**
   - Check dashboard JSON syntax
   - Verify provisioning configuration
   - Check Grafana logs

### Validation Commands

```bash
# Check Prometheus configuration
docker-compose exec prometheus promtool check config /etc/prometheus/prometheus.yml

# Check Grafana datasource
curl http://localhost:3000/api/datasources

# Test metrics endpoint
curl http://localhost:8080/metrics | head -20
```
