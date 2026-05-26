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

| Method | Route | Auth | Description |
|---|---|---|---|
| POST | `/auth/register` | — | Create a new account |
| POST | `/auth/login` | — | Authenticate and receive tokens |
| GET | `/auth/verify` | — | Verify email via OTP link |
| POST | `/auth/refresh` | — | Rotate access token using refresh token |
| POST | `/auth/forgot-password` | — | Send password reset email |
| POST | `/auth/reset-password` | — | Reset password using OTP token |
| POST | `/auth/logout` | — | Invalidate the current session |

### Users (admin only)

| Method | Route | Auth | Description |
|---|---|---|---|
| GET | `/users` | admin | List all users |
| GET | `/users/:id` | admin | Get a specific user |
| POST | `/users` | admin | Create a user |
| PATCH | `/users/:id` | admin | Update a user |
| DELETE | `/users/:id` | admin | Delete a user |

### Organizations

| Method | Route | Auth | Description |
|---|---|---|---|
| POST | `/organizations` | required | Create an organization |
| GET | `/organizations` | required | List organizations |
| GET | `/organizations/:slug` | — | Get organization by slug |
| PATCH | `/organizations/:slug` | required | Update an organization |
| DELETE | `/organizations/:slug` | required | Delete an organization |
| GET | `/my/organizations` | required | List authenticated user's organizations |
| GET | `/organizations/:slug/business-hours` | — | Get organization's business hours |
| GET | `/organizations/:slug/availability` | — | Get available slots |
| GET | `/organizations/:slug/exceptions` | required | List schedule exceptions |
| POST | `/organizations/:slug/exceptions` | required | Create a schedule exception |
| DELETE | `/organizations/:slug/exceptions/:id` | required | Delete a schedule exception |

### Members

| Method | Route | Auth | Description |
|---|---|---|---|
| GET | `/organizations/:slug/members` | — | List organization members |
| POST | `/organizations/:slug/members` | required | Add a member to the organization |
| DELETE | `/organizations/:slug/members/:user_id` | required | Remove a member |
| GET | `/organizations/:slug/members/:user_id/business-hours` | required | Get member's business hours |
| PATCH | `/organizations/:slug/members/:user_id/business-hours` | required | Update all member's business hours |
| PATCH | `/organizations/:slug/members/:user_id/business-hours/:day` | required | Update a single day's business hours |
| GET | `/organizations/:slug/members/:user_id/exceptions` | required | List member's schedule exceptions |
| POST | `/organizations/:slug/members/:user_id/exceptions` | required | Create a member schedule exception |
| DELETE | `/organizations/:slug/members/:user_id/exceptions/:id` | required | Delete a member schedule exception |

### Customers

| Method | Route | Auth | Description |
|---|---|---|---|
| GET | `/organizations/:slug/customers` | required | List customers |
| POST | `/organizations/:slug/customers` | required | Create a customer |
| GET | `/organizations/:slug/customers/:id` | required | Get a customer |
| PATCH | `/organizations/:slug/customers/:id` | required | Update a customer |
| DELETE | `/organizations/:slug/customers/:id` | required | Delete a customer |

### Services

| Method | Route | Auth | Description |
|---|---|---|---|
| GET | `/organizations/:slug/services` | — | List services |
| GET | `/organizations/:slug/services/:id` | — | Get a service |
| POST | `/organizations/:slug/services` | required | Create a service |
| PATCH | `/organizations/:slug/services/:id` | required | Update a service |
| DELETE | `/organizations/:slug/services/:id` | required | Delete a service |

### Schedules

| Method | Route | Auth | Description |
|---|---|---|---|
| POST | `/organizations/:slug/schedules` | required | Create a schedule |
| GET | `/organizations/:slug/schedules` | required | List organization schedules |
| GET | `/organizations/:slug/schedules/:id` | required | Get a schedule |
| GET | `/my/schedules` | required | List authenticated user's schedules |
| PATCH | `/organizations/:slug/schedules/:id/confirm` | required | Confirm a schedule |
| PATCH | `/organizations/:slug/schedules/:id/cancel` | required | Cancel a schedule |
| PATCH | `/organizations/:slug/schedules/:id/complete` | required | Mark schedule as completed |
| PATCH | `/organizations/:slug/schedules/:id/no-show` | required | Mark as no-show |
| PATCH | `/organizations/:slug/schedules/:id/reschedule` | required | Reschedule an appointment |
| GET | `/organizations/:slug/schedules/:id/reschedule-history` | required | Get reschedule history |

## Project Structure

```
.
├── app/
│   ├── auth/
│   │   ├── actions/          # register, login, refresh, verify, OTP, logout
│   │   └── handlers/         # HTTP handlers for auth routes
│   ├── customers/
│   │   ├── actions/          # create, list, find, update, delete, auto-create
│   │   └── handlers/
│   ├── members/
│   │   ├── actions/          # add, remove, list
│   │   └── handlers/
│   ├── organizations/
│   │   ├── actions/          # org CRUD, business hours, availability, exceptions
│   │   └── handlers/
│   ├── schedules/
│   │   ├── actions/          # create, list, find, transitions, reschedule
│   │   └── handlers/
│   ├── services/
│   │   ├── actions/          # create, list, find, update, delete
│   │   └── handlers/
│   ├── users/
│   │   ├── actions/          # create, list, find, update, delete
│   │   └── handlers/
│   ├── middleware/            # auth guard, admin guard, rate limiter
│   ├── notifications/
│   │   └── mail/             # Resend email client
│   └── handler.go            # shared base handler
├── cache/                    # Redis connection
├── database/
│   ├── migrations/           # SQL migrations (goose, embedded via go:embed)
│   └── database.go           # GORM connection + auto-migrate on startup
├── models/                   # shared data models (User, OTP, Organization, Member, Customer, Service, Schedule…)
├── routes/                   # route registration
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
