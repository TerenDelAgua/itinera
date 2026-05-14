/// <reference lib="webworker" />
// SvelteKit detecta automáticamente este archivo como Service Worker.
// Debe estar en src/service-worker.ts (no en src/lib/).

declare const self: ServiceWorkerGlobalScope;

const STATIC_CACHE = 'itinera-static-v1';
const DYNAMIC_CACHE = 'itinera-dynamic-v1';

// ============================================
// ASSETS ESTÁTICOS A PRECACHEAR (Fase 0)
// Se cachean en install y nunca se borran hasta nueva versión
// ============================================
const STATIC_ASSETS = [
    '/',
    '/offline.html',
    '/manifest.json',
    '/icon-192.png',
    '/icon-512.png',
    '/maskable-icon.png',
    '/apple-touch-icon.png',
    '/favicon.ico',
];

// ============================================
// INSTALACIÓN: Precachear assets críticos
// ============================================
self.addEventListener('install', (event: ExtendableEvent) => {
    event.waitUntil(
        caches.open(STATIC_CACHE).then((cache) => {
            console.log('[SW] Installing: caching static assets');
            // addAll falla si CUALQUIER recurso falla. Usamos add individual con catch.
            return Promise.all(
                STATIC_ASSETS.map((url) =>
                    cache.add(url).catch((err) => {
                        console.warn(`[SW] Failed to cache: ${url}`, err);
                    })
                )
            );
        }).then(() => {
            console.log('[SW] Install complete');
            self.skipWaiting();
        })
    );
});

// ============================================
// ACTIVACIÓN: Limpiar caches antiguas
// ============================================
self.addEventListener('activate', (event: ExtendableEvent) => {
    event.waitUntil(
        caches.keys().then((keys) => {
            return Promise.all(
                keys
                    .filter((key) => key !== STATIC_CACHE && key !== DYNAMIC_CACHE)
                    .map((key) => {
                        console.log(`[SW] Deleting old cache: ${key}`);
                        return caches.delete(key);
                    })
            );
        }).then(() => {
            console.log('[SW] Activate complete');
            return self.clients.claim();
        })
    );
});

// ============================================
// FETCH: Estrategia híbrida por tipo de recurso
// ============================================
self.addEventListener('fetch', (event: FetchEvent) => {
    const { request } = event;
    const url = new URL(request.url);

    // Solo interceptar GET requests
    if (request.method !== 'GET') {
        return;
    }

    // Ignorar requests de analytics/tracking (no cachear)
    if (url.pathname === '/api/events') {
        return;
    }

    // API GET: Network-First (datos frescos, fallback a cache)
    if (url.pathname.startsWith('/api/')) {
        event.respondWith(networkFirstStrategy(request));
        return;
    }

    // Assets estáticos (imágenes, CSS, JS, fuentes): Cache-First
    if (['image', 'style', 'script', 'font'].includes(request.destination)) {
        event.respondWith(cacheFirstStrategy(request));
        return;
    }

    // Navegación (HTML pages): Network-First
    if (request.mode === 'navigate' || request.destination === 'document') {
        event.respondWith(networkFirstStrategy(request));
        return;
    }

    // Default: Cache-First para todo lo demás
    event.respondWith(cacheFirstStrategy(request));
});

// ============================================
// ESTRATEGIA: Cache-First
// Para assets estáticos. Rápido, funciona offline.
// ============================================
async function cacheFirstStrategy(request: Request): Promise<Response> {
    const cached = await caches.match(request);
    if (cached) {
        // Revalidar en background (stale-while-revalidate ligero)
        fetch(request).then((response) => {
            if (response.ok) {
                caches.open(DYNAMIC_CACHE).then((cache) => cache.put(request, response));
            }
        }).catch(() => { }); // Silenciar errores de background fetch
        return cached;
    }

    try {
        const response = await fetch(request);
        if (response.ok) {
            const cache = await caches.open(DYNAMIC_CACHE);
            cache.put(request, response.clone());
        }
        return response;
    } catch {
        // Asset no cacheado y offline: devolver 503 silencioso
        // No hacer throw — rompe la app
        return new Response('Offline', { status: 503 });
    }
}

// ============================================
// ESTRATEGIA: Network-First
// Para API y navegación. Datos frescos, fallback a cache.
// ============================================
async function networkFirstStrategy(request: Request): Promise<Response> {
    try {
        const response = await fetch(request);
        if (response.ok) {
            const cache = await caches.open(DYNAMIC_CACHE);
            cache.put(request, response.clone());
        }
        return response;
    } catch {
        const cached = await caches.match(request);
        if (cached) return cached;

        // Navegación offline: mostrar offline.html
        if (request.mode === 'navigate') {
            const offline = await caches.match('/offline.html');
            if (offline) return offline;
        }

        // API offline: devolver JSON de error
        if (request.url.includes('/api/')) {
            return new Response(
                JSON.stringify({
                    error: 'offline',
                    message: 'You are offline. Some features may be limited.'
                }),
                {
                    status: 503,
                    headers: { 'Content-Type': 'application/json' }
                }
            );
        }

        // Fallback genérico
        return new Response('Offline', { status: 503 });
    }
}