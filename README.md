# Barbearia API

A RESTful API for a barbershop management system, built as a personal study project to explore Go's ecosystem in the context of a real-world application.

## About

This project was developed to practice building production-grade backend services with Go, covering authentication flows, database migrations, caching, email delivery, rate limiting, and automated testing.

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.26 |
| Web Framework | [Gin](https://github.com/gin-gonic/gin) |
| ORM | [GORM](https://gorm.io) with PostgreSQL driver |
| Database | PostgreSQL (hosted on [Neon](https://neon.tech)) |
| Migrations | [Goose](https://github.com/pressly/goose) — embedded and auto-applied on startup |
| Cache / Session | Redis via [go-redis](https://github.com/redis/go-redis) |
| Authentication | JWT ([golang-jwt/jwt](https://github.com/golang-jwt/jwt)) with access + refresh token strategy |
| Email | [Resend](https://resend.com) SDK for transactional emails (OTP, password reset) |
| Rate Limiting | [ulule/limiter](https://github.com/ulule/limiter) backed by Redis |
| CORS | gin-contrib/cors |
| Hot Reload | [Air](https://github.com/air-verse/air) |
| CI | GitHub Actions — `go vet` + `go test -race` on every push/PR |

## API Endpoints

### Auth

| Method | Route | Description |
|---|---|---|
| POST | `/auth/register` | Create a new account |
| POST | `/auth/login` | Authenticate and receive tokens |
| GET | `/auth/verify` | Verify email via OTP link |
| POST | `/auth/refresh` | Rotate access token using refresh token |
| POST | `/auth/forgot-password` | Send password reset email |
| POST | `/auth/reset-password` | Reset password using OTP token |
| POST | `/auth/logout` | Invalidate the current session |

### Users (admin only)

| Method | Route | Description |
|---|---|---|
| GET | `/users` | List all users |
| GET | `/users/:id` | Get a specific user |
| POST | `/users` | Create a user |
| PATCH | `/users/:id` | Update a user |
| DELETE | `/users/:id` | Delete a user |

## Project Structure

```
.
├── app/
│   ├── auth/
│   │   ├── actions/    # business logic: register, login, refresh, OTP…
│   │   └── handlers/   # HTTP handlers for auth routes
│   ├── users/
│   │   ├── actions/    # business logic: create, list, find, update, delete
│   │   └── handlers/   # HTTP handlers for user routes
│   ├── middleware/     # auth guard, rate limiter
│   └── notifications/
│       └── mail/       # Resend email client
├── cache/              # Redis connection
├── database/
│   ├── migrations/     # SQL migrations (goose, embedded via go:embed)
│   └── database.go     # GORM connection + auto-migrate on startup
├── models/             # shared data models (User, OTP, Auth)
├── routes/             # route registration
└── main.go
```

## Getting Started

**Prerequisites:** Go, PostgreSQL, Redis, a [Resend](https://resend.com) account.

```powershell
# Clone and install dependencies
git clone https://github.com/claudsondouglas/barbearia-api
cd barbearia-api
go mod download

# Configure environment
cp .env-example .env
# fill in the values in .env

# Run with hot reload
air

# Or without hot reload
go run .
```

Migrations are applied automatically on startup — no manual step required.

## Running Tests

```powershell
go test -race ./...
```

Tests use [miniredis](https://github.com/alicebob/miniredis) to mock Redis and [go-sqlmock](https://github.com/DATA-DOG/go-sqlmock) for database interactions, so no external services are needed to run the test suite.

## Environment Variables

See [.env-example](.env-example) for all required variables.
