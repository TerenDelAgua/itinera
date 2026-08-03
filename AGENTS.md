# ITINERA - Agent Instructions

## Project Overview

**Type:** Full-stack monorepo (Go backend + SvelteKit frontend)
**Stack:** Go + Chi Router + PostgreSQL 16 / SvelteKit 5 + Tailwind v4 + TypeScript

## Key Commands

### Frontend (`itinera-web/`)
```bash
pnpm dev          # Start dev server (http://localhost:5173)
pnpm check        # TypeScript + Svelte typecheck
pnpm build        # Production build
pnpm preview      # Preview production build
```

### Backend (`backend/`)
```bash
go run cmd/api/main.go   # Start API server (http://localhost:8080)
go test ./... -v         # Run all tests
```

### Database
```bash
cd backend && docker compose up -d    # Start PostgreSQL (auto-runs migrations)
docker compose down -v                 # Stop and remove volumes
```

## Required Setup

1. **Database:** `cd backend && docker compose up -d` (migrations auto-applied from `./migrations/`)
2. **Frontend env:** Create `itinera-web/.env` with `VITE_API_URL=http://localhost:8080`
3. **Local backend env:** Create `backend/.env` with the development template
   (see `backend/.env.example` if present, or copy the variables listed in
   `docs/ITINERA_FUNCTIONAL_SCOPE.md` § Configuration)
4. **Startup order:** Backend first → then `pnpm dev`

### Production env vars (Railway)

These are validated by `internal/config/config.go::Validate`. A missing
or weak one is a fatal startup error in production:

| Variable | Min length | Notes |
|---|---|---|
| `DATABASE_URL` | — | Postgres connection string |
| `JWT_SECRET` | 32 | Must not equal the default `dev-secret-change-me` |
| `PUBLIC_ORIGIN` | — | Used by `BuildShareURL`; e.g. `https://goitinera.app` |
| `ITINERA_INTERNAL_TOKEN` | 32 | Gates `/api/analytics/*`; generate with `openssl rand -base64 32` |

You can copy this directly to Railway once per project — values are stable
unless you rotate them intentionally.

```powershell
# PowerShell snippet: generate a 40-char Internal Token
$token = [Convert]::ToBase64String((New-Object Security.Cryptography.RNGCryptoServiceProvider).GetBytes(32)).TrimEnd('=').Substring(0, 40)
Set-Clipboard -Value $token; Write-Host "Token: $token"
```

If the deploy fails with `❌ Config validation failed: ...`, the missing
variable is named in the log line. Cross-reference with the table above.

## Architecture Notes

- **Svelte 5:** Uses runes (`$state`, `$derived`, `$effect`). Do NOT use legacy `$:` reactive syntax
- **Guest-First Auth:** Sessions via HttpOnly cookie, optional JWT upgrade (endpoint exists, no UI yet)
- **Multi-currency:** Expenses store `original_amount` + `original_currency` + converted `amount` + `exchange_rate`
- **Internal Sessions:** `internal_test.go` covers the dev cookie/header helper; the flag is set on the request side, not from the source trip (spec 015 §4.4)
- **DB Migrations:** Applied automatically on container start via `docker-entrypoint-initdb.d`. Every migration must be idempotent (see user rules) and use a guard schema at the top

## Directory Ownership

| Directory | Owner |
|-----------|-------|
| `backend/` | Go API + PostgreSQL |
| `itinera-web/` | SvelteKit frontend |
| `docs/` | Project documentation |
| `backend/migrations/` | SQL schema (apply order: 001→NNN, currently up to 016) |

## Common Pitfalls

- **DB not running:** Always check `docker ps` before reporting "connection refused"
- **Backend deploy fails with `Config validation failed`:** Missing production env var (see table above). The process exits, Railway restarts in a crash loop, the gateway returns 502 — which the browser then shows as a CORS error
- **Svelte 4 patterns:** Don't use `$:` or `export let` - use runes instead
- **Missing .env:** Frontend will crash without `VITE_API_URL`
- **Wrong `PUBLIC_ORIGIN`:** Share links come out as `localhost:5173` in production. Set the production origin in Railway before pushing the share feature
- **Wrong `VITE_PUBLIC_APP_URL` on Vercel:** Same share-link symptom, but on the frontend side. Set it on Vercel too
- **Binary files:** Don't commit `.exe` files in `backend/`

## Reference Docs

- `docs/ITINERA_FUNCTIONAL_SCOPE.md` - Feature roadmap and phases
- `docs/TEREN_DESIGN_SYSTEM.md` - UI tokens and component guidelines