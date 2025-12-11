# Go Microservices E-commerce Example

This repository is a starter scaffold for a Go microservices project using MySQL, JWT, and a Clean Architecture approach.

## What's Inside

- **Services:** An `auth` service is implemented, with skeletons for `product` and `transaction`.
- **Database:** `docker-compose.yml` configured to run a MySQL 8 database.
- **Schema:** A complete database schema is provided in `sql/schema.sql`.
- **API Collection:** A Postman collection (`Rakamin Evermos Virtual Internship.postman_collection.json`) and a markdown reference (`POSTMAN.md`) are included for API testing.

## Getting Started

1.  **Prerequisites:** Make sure you have Docker and Go (1.24+) installed.

2.  **Clone the repository:**

    ```bash
    git clone <repository-url>
    cd <repository-directory>
    ```

3.  **Create Environment File:**
    Copy the example environment file. This file contains all the necessary configuration for the database and services.
    ```bash
    cp .env.example .env
    ```

## Running the Application

You have two primary ways to run the services for development.

### Option 1: Run Everything with Docker Compose (Recommended)

This is the simplest and most consistent method. A single command will build and run the MySQL database and all Go microservices in their respective containers.

```bash
docker compose up --build -d
```

The services will be available at the following local addresses:

- **Auth Service:** `http://localhost:8080`
- **Product Service:** `http://localhost:8081`
- **Transaction Service:** `http://localhost:8082`

### Option 2: Run a Single Service Locally (for Debugging)

If you need to run a single service on your host machine for easier debugging, follow these steps.

1.  **Start the database container:**
    ```bash
    docker compose up -d db
    ```
2.  **Temporarily modify your `.env` file:** For your local Go application to find the database container, you must change the `DB_HOST`.
    - Change `DB_HOST=db` to `DB_HOST=127.0.0.1`.
3.  **Run the service:** From the project's **root directory**, run the following command (example for the `auth` service):
    ```bash
    export $(grep -v '^#' .env | xargs) && go run cmd/auth/main.go
    ```
4.  **Revert the change:** When you are finished, remember to change `DB_HOST` back to `db` in your `.env` file so that the full Docker Compose setup works correctly.
