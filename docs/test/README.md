# Itinera Testing Suite

Esta documentación detalla la arquitectura de pruebas de Itinera, los tipos de tests disponibles y cómo ejecutarlos para asegurar la estabilidad del sistema.

## 🏗️ Estructura de Testing

Itinera sigue una estrategia de pirámide de tests dividida en tres capas:

1.  **Backend (Go):** Pruebas de integración de base de datos y validación de lógica de API.
2.  **Frontend (Svelte 5):** Pruebas unitarias de utilidades y reactividad de componentes (Runes).
3.  **E2E (Playwright):** Pruebas de flujos completos de usuario en navegadores reales.

---

## 🚀 Cómo ejecutar los tests

### 1. Backend (Go)
Asegúrate de tener la base de datos de desarrollo levantada (`docker compose up -d`).

```bash
cd backend
go test ./... -v
```

### 2. Frontend (Vitest)
Pruebas de lógica pura y renderizado de componentes.

```bash
cd itinera-web
pnpm test          # Modo interactivo (watch)
pnpm test --run    # Ejecución única
```

### 3. E2E (Playwright)
Validación de flujos completos. Requiere que tanto el backend como el frontend estén disponibles (Playwright intentará levantar el servidor de desarrollo si no está activo).

```bash
cd itinera-web
pnpm test:e2e
```

---

## 📋 Inventario de Tests

### Backend
*   **Integración DB:** Localizado en `backend/internal/database/*_test.go`. Verifica agregaciones SQL y relaciones.
*   **API Handlers:** Localizado en `backend/internal/http/handlers/*_test.go`. Valida entradas JSON y códigos de error.
*   **Middleware:** Localizado en `backend/internal/http/middleware/*_test.go`. Asegura la gestión de sesiones y el sistema de "Fork-on-Write" para demos.
*   **Event Tracking:** Localizado en `backend/internal/database/events_test.go` (integración) y `backend/internal/http/handlers/events_test.go` (validación).

### Frontend
*   **Utilidades:** `src/lib/utils.test.ts`. Formateo de fechas, monedas y emojis.
*   **Servicios:** `src/lib/services/tracking.test.ts`. Validación de payloads de eventos y fallback de telemetría.
*   **Componentes:** `src/lib/components/*.test.ts`. Reactividad de Svelte 5 y lógica de `localStorage`.

### E2E
*   **Flujos Críticos:** `tests/itinerary.spec.ts`. Creación de viajes, persistencia de sesión y filtrado de actividades locales vs globales.
*   **Multidivisa:** `tests/currency.spec.ts`. Herencia de divisas y cálculos.
---

## 🛠️ Herramientas utilizadas
- **Go Test:** Herramienta nativa de Go con la librería `testify`.
- **Vitest:** Test runner ultra-rápido para Vite/Svelte.
- **Svelte Testing Library:** Utilidades para testear componentes Svelte sin depender de implementación interna.
- **Playwright:** Automatización de navegador para pruebas de extremo a extremo.
