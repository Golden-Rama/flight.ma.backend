# GR Multi Aggregator

A high-performance Go rewrite of the Multi Aggregator for Flight Service. It combines multiple flight providers, queries availability concurrently, and exposes a beautiful, lightweight Single Page Application (SPA) dashboard to manage flight providers, quests, bookings, and users.

This project implements the same high standards and clean architecture as `flight.service`.

## Technical Features & Stack

- **Framework**: **Echo** (high-performance web framework).
- **ORM**: **GORM** supporting both **MySQL** and **Postgres**.
- **Clean Architecture**:
  - `config/` - Environment configuration loaders.
  - `src/entity/` - Database entities mapping GORM models.
  - `src/repository/` - GORM repository pattern implementations.
  - `src/service/` - Business logic including concurrent search aggregation.
  - `src/handler/` - HTTP request handlers (API and Web).
  - `src/middleware/` - Custom middlewares (API Bearer JWT and Web Cookie Session validation).
  - `src/factory/` - Dependency Injection resolver.
  - `src/utils/` - Shared encryption and string utils.
- **Dual Server Ports**: Starts separate API (`:4001`) and Web Dashboard (`:4000`) servers concurrently in a single process.
- **Lightweight Premium Frontend**:
  - Served at `/dashboard`.
  - Built using **Vanilla HTML5**, **Vanilla CSS**, and **Vanilla JS (fetch)**.
  - No massive frameworks; fast, responsive, and styled with modern gradients, glassmorphism card layouts, and hover transitions.
  - Uses Lucide Icons for high-fidelity rendering.
- **Auto Migration & Seeding**: Automatically migrates database schemas on boot and seeds a default Administrator (`admin` / `admin123`).

---

## Directory Structure

```
├── config
│   └── config.go          # Config loader using envconfig
├── src
│   ├── dto
│   │   ├── dtos.go        # Shared request/response data transfer objects
│   │   └── search.go      # Flight search results representation model
│   ├── entity
│   │   └── entities.go    # GORM DB models (User, Role, Provider, Quest, Booking)
│   ├── factory
│   │   └── resolver.go    # Resolver factory (DI & DB bootloader)
│   ├── handler
│   │   ├── api_handler.go        # Flight API controllers
│   │   └── dashboard_handler.go  # Dashboard CRUD rest endpoints
│   ├── middleware
│   │   └── auth_middleware.go    # API JWT & Dashboard session middlewares
│   ├── repository
│   │   └── repositories.go       # Data access layer using GORM
│   ├── service
│   │   └── services.go           # Concurrent search aggregator & booking proxy
│   └── utils
│       └── utils.go              # BCrypt, MD5, and string helper functions
├── static
│   ├── index.html         # SPA HTML skeleton structure
│   ├── style.css          # Premium glassmorphism dark mode stylesheet
│   └── app.js             # Front-end router, state, and API fetch calls
├── .env.example
├── go.mod
├── go.sum
├── main.go                # Unified entrypoint (starts API & Dashboard Echo instances)
└── README.md
```

---

## Getting Started

### Prerequisites

- Go 1.21 or later
- MySQL or PostgreSQL database

### Configuration

Copy `.env.example` to `.env` and fill in your database configuration details:

```bash
cp .env.example .env
```

Default credentials in `.env.example`:
- `DASHBOARD_PORT`: `4000`
- `API_PORT`: `4001`
- `DB_DRIVER`: `mysql`

### Running the Project

Start the application:

```bash
go run main.go
```

The system will:
1. Connect to the database.
2. Auto-run GORM migrations to create tables (`users_ma`, `roles_ma`, `flight_quests`, `flight_bookings`, `flight_providers`).
3. Seed the default roles and user `admin` with password `admin123`.
4. Start the API Server on `http://localhost:4001`.
5. Start the Web Dashboard on `http://localhost:4000`.

---

## API Endpoints

### Flight Search Availability (API Port 4001)
- `POST /api/v1/search`
  - Concurrently queries search availability from all active providers.
  - Sorts flights by fare and applies `DepartTime` uniqueness filtering.

### Booking & Reservation Passthrough (API Port 4001)
- `POST /api/v1/fare-detail`
- `POST /api/v1/reservation`
- `POST /api/v1/check-reservation`
- `POST /api/v1/issue-ticket`
- `POST /api/v1/cancel-reservation`
  - Proxies payloads directly to the provider based on the `X-Provider` header.

### Web Dashboard & CRUD (Dashboard Port 4000)
- `GET /dashboard` - Interactive SPA Web Interface.
- `/api/dashboard/login` & `/api/dashboard/logout` - Session endpoints.
- `/api/dashboard/providers` - CRUD for Flight Providers.
- `/api/dashboard/quests` - CRUD for Quest configurations.
- `/api/dashboard/bookings` - CRUD for Booking configurations.
- `/api/dashboard/users` - CRUD for System Users.
