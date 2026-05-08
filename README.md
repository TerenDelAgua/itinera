# Itinera

> **Trip planning without friction.** Create trips, organize destinations, track expenses in multiple currencies, and build your itinerary — no account required.

[![SvelteKit](https://img.shields.io/badge/SvelteKit-5-FF3E00?logo=svelte&logoColor=white)](https://svelte.dev)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql&logoColor=white)](https://www.postgresql.org)
[![Tailwind CSS](https://img.shields.io/badge/Tailwind-v4-06B6D4?logo=tailwindcss&logoColor=white)](https://tailwindcss.com)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## Overview

Itinera is a full-stack travel planning application built around the **Guest-First** philosophy: users start planning immediately without registration, experience real value, and only commit to an account when they want to save or collaborate.

**Design language:** [TEREN Design System](docs/TEREN_DESIGN_SYSTEM.md) — outdoor-readable contrast, zero cognitive friction, single-page feel.

---

## Tech Stack

| Layer | Technology |
|---|---|
| **Frontend** | SvelteKit 5 (Runes) · Tailwind v4 · TypeScript |
| **Backend** | Go · Chi Router · pgx/v5 connection pool |
| **Database** | PostgreSQL 16 (UUID PKs, relational integrity) |
| **Auth** | Guest sessions via HttpOnly cookie · optional JWT upgrade |
| **i18n** | English & Spanish · runtime locale switching |
| **PWA** | Service Worker · Web App Manifest · offline fallback page |
| **Deploy** | Frontend → Vercel · Backend → Railway/Fly.io · DB → Supabase |

---

## Features

- **Trips** — Create, edit and delete trips with date ranges and base currency
- **Destinations** — Organize places/cities within a trip with optional dates
- **Itinerary** — Schedule activities per destination with time and notes
- **Expenses** — Log expenses in any currency with real-time conversion to the trip's base currency
- **Multi-currency** — Per-trip and per-destination default currencies with automatic inheritance
- **Offline support** — PWA with Cache-First for assets, Network-First for API calls
- **Localization** — Full EN/ES UI with locale-aware date and number formatting

---

## Project Structure

```
itinera/
├── backend/                  # Go API
│   ├── cmd/api/main.go       # Entry point, router setup
│   ├── internal/
│   │   ├── config/           # Environment configuration
│   │   ├── database/         # Repositories (trips, places, expenses, activities, auth)
│   │   ├── http/handlers/    # HTTP handlers
│   │   ├── models/           # Domain types
│   │   └── services/         # Business logic (expense conversion, exchange rates)
│   ├── migrations/           # SQL migrations (001→006, auto-applied via Docker)
│   └── docker-compose.yaml   # PostgreSQL 16 local dev container
│
└── itinera-web/              # SvelteKit frontend
    ├── src/
    │   ├── routes/           # SvelteKit file-based routing
    │   ├── lib/
    │   │   ├── components/   # Reusable Svelte components
    │   │   ├── i18n/         # en.json / es.json translation files + store
    │   │   ├── types/        # TypeScript interfaces (Trip, Place, Activity, Expense)
    │   │   └── utils/        # Date helpers, currency symbols, category maps
    │   └── service-worker.ts # PWA service worker (Cache-First + Network-First)
    └── static/
        ├── manifest.json     # PWA manifest
        └── offline.html      # Offline fallback (bilingual, no JS framework)
```

---

## Getting Started

### Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org) + [pnpm](https://pnpm.io)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)

### 1. Database

```bash
cd backend
docker compose up -d   # Starts PostgreSQL 16, auto-runs all migrations
```

### 2. Backend

```bash
# Create backend/.env (copy from example below)
cd backend
go run cmd/api/main.go  # API starts on http://localhost:8080
```

**`backend/.env`**
```env
DATABASE_URL=postgres://teren_dev:qwerty123@localhost:5432/itinera?sslmode=disable
API_PORT=8080
JWT_SECRET=change-me-in-production
ENVIRONMENT=development
```

### 3. Frontend

```bash
cd itinera-web

# Create itinera-web/.env
echo "VITE_API_URL=http://localhost:8080" > .env

pnpm install
pnpm dev   # App starts on http://localhost:5173
```

---

## API Reference

Base URL: `http://localhost:8080/api/v1`

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/health` | Health check + DB ping |
| `POST` | `/auth/guest` | Create guest session |
| `POST` | `/auth/upgrade` | Upgrade guest → registered user |
| `GET` | `/trips` | List user's trips |
| `POST` | `/trips` | Create trip |
| `GET` | `/trips/:id` | Get trip details |
| `PUT` | `/trips/:id` | Update trip |
| `DELETE` | `/trips/:id` | Delete trip |
| `GET` | `/trips/:id/places` | List destinations |
| `POST` | `/trips/:id/places` | Create destination |
| `GET` | `/trips/:id/places/:pid` | Get destination detail |
| `DELETE` | `/trips/:id/places/:pid` | Delete destination |
| `GET` | `/trips/:id/activities` | List activities |
| `POST` | `/trips/:id/activities` | Create activity |
| `DELETE` | `/trips/:id/activities/:aid` | Delete activity |
| `GET` | `/trips/:id/expenses` | List expenses |
| `POST` | `/trips/:id/expenses` | Create expense |
| `DELETE` | `/trips/:id/expenses/:eid` | Delete expense |

---

## Database Schema

Migrations are applied automatically on container start from `backend/migrations/`:

| File | Description |
|---|---|
| `001_init.sql` | Core tables: `users`, `sessions`, `trips`, `places` |
| `002_expenses.sql` | `expenses` table with category and amount |
| `003_update_places.sql` | Place date fields and ordering |
| `004_activities.sql` | `activities` table with time and notes |
| `005_multi_currency.sql` | `original_amount`, `original_currency`, `exchange_rate` |
| `006_default_expense_currency.sql` | Per-trip and per-place default currencies |

---

## Development Scripts

### Backend

```bash
go run cmd/api/main.go   # Start API server
go test ./... -v         # Run all tests
docker compose up -d     # Start DB
docker compose down -v   # Stop DB and remove volumes
```

### Frontend

```bash
pnpm dev          # Dev server (http://localhost:5173)
pnpm build        # Production build (includes service worker)
pnpm preview      # Preview production build
pnpm check        # TypeScript + Svelte type check
```

---

## Deployment

### Frontend (Vercel)

1. Connect the `itinera-web/` directory to a Vercel project
2. Set build command: `pnpm build`
3. Set output directory: `.svelte-kit/output` (or use SvelteKit adapter auto-detection)
4. Add environment variable: `VITE_API_URL=https://your-api-url.com`

### Backend (Railway / Fly.io)

1. Deploy the `backend/` directory
2. Set environment variables:
   ```
   DATABASE_URL=postgresql://user:password@host:5432/dbname?sslmode=require
   JWT_SECRET=your-production-secret
   ENVIRONMENT=production
   API_PORT=8080
   ```

### Database (Supabase)

1. Create a new Supabase project
2. Run migrations manually in the SQL editor (in order: 001 → 006)
3. Copy the connection string from **Project Settings → Database → URI**

---

## Roadmap

| Phase | Status | Scope |
|---|---|---|
| **Phase 1** — MVP Web | 🟢 In Progress | Trips, places, activities, expenses, guest auth, i18n, PWA |
| **Phase 2** — Collaboration | 🔜 Planned | Trip sharing, roles, budget tracking, PDF export |
| **Phase 3** — Mobile (Flutter) | 🔜 Planned | iOS/Android, offline-first, GPS tagging, push notifications |
| **Phase 4** — Intelligence | 🔜 Planned | AI suggestions, smart budgeting, community templates |

See [ITINERA_FUNCTIONAL_SCOPE.md](docs/ITINERA_FUNCTIONAL_SCOPE.md) for full details.

---

## Design System

Itinera is built on the **TEREN Design System** — a design language optimized for outdoor readability, minimal cognitive load, and single-page interaction patterns.

Key principles:
- **Unified Widgets** over modals — inline forms that don't interrupt flow
- **Guest-First** — zero friction before any registration gate
- **Warm neutrals** (`#F5F4F1` background, `#FF8C42` primary accent)

See [TEREN_DESIGN_SYSTEM.md](docs/TEREN_DESIGN_SYSTEM.md) for tokens, components, guidelines and styles.

---

## Contributing

This is a private project under active development. Internal contributors should follow the patterns documented in [AGENTS.md](AGENTS.md).

---

*Built with intention. Designed for flow. Owned by TEREN.* 🧡
