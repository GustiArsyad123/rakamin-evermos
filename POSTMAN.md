# Postman-like Route Documentation

This file lists the main routes, methods, request body and query parameters for the implemented services (Auth, Product, Transaction). Use this as a quick reference.

Base URLs when running locally via Docker Compose:

- Auth: `http://localhost:8080`
- Product: `http://localhost:8081`
- Transaction: `http://localhost:8082`

1. Auth

- POST /api/v1/auth/register

  - Body (JSON): { "name": string, "email": string, "phone": string, "password": string }
  - Response: { "token": string }

- POST /api/v1/auth/login

  - Body (JSON): { "email": string, "password": string }
  - Response: { "token": string }

- GET /api/v1/auth/users

  - Headers: `Authorization: Bearer <token>`
  - Response: { "data": [ { "id": int, "name": string, ... }, ... ] }

- GET /api/v1/auth/users/:id

  - Headers: `Authorization: Bearer <token>`
  - Response: { "id": int, "name": string, ... }

- POST /api/v1/auth/forgot-password

  - Body (JSON): { "email": string }
  - Response: { "reset_token": string } (development only — token is returned; in production this would be emailed)

- POST /api/v1/auth/reset-password

  - Body (JSON): { "token": string, "password": string }
  - Response: 204 No Content

- POST /api/v1/auth/sso/google

  - Body (JSON): { "id_token": string }
  - Response: { "token": string }
  - Notes: Accepts a Google ID token (JWT). In this development setup the server validates the token via `https://oauth2.googleapis.com/tokeninfo?id_token=<token>` and will create a new user automatically if the email does not exist. In production, verify audience (`aud`) against your Google OAuth Client ID and send the ID token to the backend securely.

2. Product

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

3. Transaction

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

4. Category

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

Notes

- Category management is admin-only; use DB to mark a user as admin (see `sql/seed.sql`).
- Pagination and filtering params follow the patterns above.
