# GR Multi Aggregator

A high-performance Go backend service for the Flight Multi-Aggregator. It combines multiple flight providers, queries availability concurrently, and exposes CRUD REST API endpoints to manage flight providers, quests, bookings, and users.

> [!NOTE]
> The administrative frontend application has been decoupled from this repository and is designed to run separately (e.g., Next.js / UmiJS frontend on port `3000` or `8000`). The backend allows cross-origin requests from these frontend applications via configured CORS middleware.

---

## Technical Features & Stack

- **Framework**: **Echo** (high-performance, minimalist Go web framework).
- **ORM**: **GORM** supporting both **MySQL** and **PostgreSQL** databases.
- **Unified Port Architecture**: Starts a single unified Echo API server on port `4001` (by default) that serves both client APIs and administrative dashboard REST endpoints.
- **CORS Support**: Ready-configured middleware to securely communicate with a separated modern frontend application.
- **Auto Migration & Seeding**: Automatically migrates schemas on startup and seeds default Roles and an Administrator user (`admin` / `admin123`).
- **Utility Scripts**: Built-in Go scripts for token generation, database backups, and staging-to-local synchronization.
- **CI/CD Pipeline**: Automated GitHub Actions workflow for linting, security scanning (`govulncheck`), Docker builds, and deployment via VPN to Staging and Production servers.

---

## Directory Structure

```
├── .github
│   └── workflows
│       └── deploy.yml            # CI/CD pipeline definition
├── config
│   └── config.go                 # Configuration loader using envconfig
├── scripts
│   ├── dump_db
│   │   └── main.go               # Utility to dump database structure and data to init.sql
│   ├── generate_token
│   │   └── main.go               # Utility to generate JWT tokens for local testing
│   └── sync_to_local
│       └── main.go               # Utility to sync configurations from staging RDS to local Postgres
├── src
│   ├── dto
│   │   ├── dtos.go               # Shared request/response data transfer objects
│   │   └── search.go             # Flight search models and structures
│   ├── entity
│   │   └── entities.go           # GORM DB entity models (User, Role, Provider, Quest, Booking)
│   ├── factory
│   │   └── resolver.go           # Resolver factory for Dependency Injection & DB bootloading
│   ├── handler
│   │   ├── api_handler.go        # Flight API controllers & OAuth2 client credentials endpoint
│   │   └── dashboard_handler.go  # Dashboard CRUD REST endpoints
│   ├── middleware
│   │   └── auth_middleware.go    # Bearer JWT and Cookie Session validation middlewares
│   ├── repository
│   │   └── repositories.go       # Data access layer implementations using GORM
│   ├── service
│   │   └── services.go           # Business logic, concurrent search aggregator, and booking proxy
│   └── utils
│       └── utils.go              # BCrypt, MD5, and string helper functions
├── .dockerignore
├── .env.example                  # Example environment variables template
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── entrypoint.sh                 # Docker container entrypoint script
├── go.mod
├── go.sum
├── main.go                       # Unified server entrypoint (starts the Echo API server)
└── README.md
```

---

## Getting Started

### Prerequisites

- Go 1.21 or later
- MySQL or PostgreSQL database

### Configuration

Copy `.env.example` to `.env` and fill in your database and environment settings:

```bash
cp .env.example .env
```

Default credentials in `.env.example`:
- `API_PORT`: `4001`
- `DB_DRIVER`: `mysql`
- `DB_NAME`: `gr-flight-service`
- `JWT_SECRET`: `super_secret_jwt_key`

### Running the Project

Start the Go backend application:

```bash
go run main.go
```

On startup, the system will:
1. Load configurations from environment variables.
2. Connect to the database.
3. Automatically run GORM migrations to verify or create necessary tables (`users_ma`, `roles_ma`, `flight_quests`, `flight_bookings`, `flight_providers`).
4. Seed default roles and the initial Administrator user (`admin` / `admin123`).
5. Start the Echo server listening on the port configured by `API_PORT` (default: `4001`).

---

## Utility Scripts

The project includes administrative scripts located inside the `scripts/` directory:

### 1. Database Dumper
Dumps the `flight_quests`, `flight_bookings`, and `flight_providers` tables from your active database to an `init.sql` file.
```bash
go run scripts/dump_db/main.go
```

### 2. JWT Token Generator
Generates a mock JWT token valid for 24 hours to quickly test endpoints locally.
```bash
go run scripts/generate_token/main.go
```

### 3. Staging-to-Local Database Synchronizer
Connects to the staging RDS MySQL database and synchronizes all structures, roles, providers, quests, bookings, and users into a local PostgreSQL database.
```bash
go run scripts/sync_to_local/main.go
```

---

## API Endpoints

### 1. Client Authorization (OAuth2)
- `POST /oauth2/token`
  - Generates client tokens using client credentials flow (returns a Bearer token for api tests).

### 2. Flight APIs (Port 4001 - Requires Bearer JWT Token)
- `POST /api/v1/search`
  - Concurrently queries flight availability across all active providers, sorts results, and applies uniqueness filtering.
- `POST /api/v1/fare-detail`
- `POST /api/v1/reservation`
- `POST /api/v1/check-reservation`
- `POST /api/v1/issue-ticket`
- `POST /api/v1/cancel-reservation`
  - Proxies requests directly to the respective third-party provider indicated by the `X-Provider` header.

### 3. Dashboard CRUD REST API (Port 4001 - Requires Session Cookie `admin_auth`)
- **Authentication**:
  - `POST /api/dashboard/login` - Log in to establish an admin session.
  - `POST /api/dashboard/logout` - Clear the session.
  - `GET /api/dashboard/me` - Fetch details of the logged-in administrator.
- **Flight Providers**:
  - `GET /api/dashboard/providers` - List all providers.
  - `POST /api/dashboard/providers` - Create a provider.
  - `GET /api/dashboard/providers/:id` - Fetch details.
  - `PUT /api/dashboard/providers/:id` - Update provider.
  - `DELETE /api/dashboard/providers/:id` - Delete provider.
- **Quest Configurations**:
  - `GET /api/dashboard/quests` - List all quests.
  - `POST /api/dashboard/quests` - Create a quest configuration.
  - `GET /api/dashboard/quests/:id` - Fetch quest details.
  - `PUT /api/dashboard/quests/:id` - Update quest configuration.
  - `DELETE /api/dashboard/quests/:id` - Delete quest configuration.
- **Booking Configurations**:
  - `GET /api/dashboard/bookings` - List booking configs.
  - `POST /api/dashboard/bookings` - Create a booking config.
  - `GET /api/dashboard/bookings/:id` - Fetch booking config details.
  - `PUT /api/dashboard/bookings/:id` - Update booking config.
  - `DELETE /api/dashboard/bookings/:id` - Delete booking config.
- **User Management**:
  - `GET /api/dashboard/users` - List all users.
  - `POST /api/dashboard/users` - Create a user.
  - `GET /api/dashboard/users/:id` - Fetch user details.
  - `PUT /api/dashboard/users/:id` - Update user.
  - `DELETE /api/dashboard/users/:id` - Delete user.
- **Role Listing**:
  - `GET /api/dashboard/roles` - Get all roles in the system.
