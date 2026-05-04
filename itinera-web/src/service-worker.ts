/// <reference lib="webworker" />
// SvelteKit detecta automáticamente este archivo como Service Worker.
// Debe estar en src/service-worker.ts (no en src/lib/).

declare const self: ServiceWorkerGlobalScope;

const STATIC_CACHE = 'itinera-static-v1';
const DYNAMIC_CACHE = 'itinera-dynamic-v1';

const STATIC_ASSETS = [
    '/',
    '/offline.html',
];

// Instalación: Cache de assets estáticos
self.addEventListener('install', (event: ExtendableEvent) => {
    event.waitUntil(
        caches.open(STATIC_CACHE).then((cache) => {
            console.log('[SW] Caching static assets');
            return cache.addAll(STATIC_ASSETS);
        }).catch(err => console.error('[SW] Cache install error:', err))
    );
    self.skipWaiting();
});

// Activación: Limpieza de caches antiguos
self.addEventListener('activate', (event: ExtendableEvent) => {
    event.waitUntil(
        caches.keys().then((keys) => {
            return Promise.all(
                keys
                    .filter((key) => key !== STATIC_CACHE && key !== DYNAMIC_CACHE)
                    .map((key) => caches.delete(key))
            );
        }).then(() => self.clients.claim())
    );
});

// Fetch: Estrategia híbrida
self.addEventListener('fetch', (event: FetchEvent) => {
    const { request } = event;
    const url = new URL(request.url);

    // API (GET): Network-First
    if (url.pathname.startsWith('/api/') && request.method === 'GET') {
        event.respondWith(networkFirstStrategy(request));
        return;
    }

    // Assets estáticos: Cache-First
    if (request.method === 'GET') {
        event.respondWith(cacheFirstStrategy(request));
        return;
    }
});

// Estrategia Cache-First (para assets)
async function cacheFirstStrategy(request: Request): Promise<Response> {
    const cached = await caches.match(request);
    if (cached) return cached;

    try {
        const response = await fetch(request);
        if (response.ok) {
            const cache = await caches.open(DYNAMIC_CACHE);
            cache.put(request, response.clone());
        }
        return response;
    } catch {
        // Fallback a offline.html para navegación
        if (request.mode === 'navigate') {
            // caches.match() es async — hay que awaitar antes del ?? 
            const offline = await caches.match('/offline.html');
            return offline ?? new Response('Offline', { status: 503 });
        }
        throw new Error('Fetch failed and no cache available');
    }
}

// Estrategia Network-First (para API)
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

        // Respuesta de fallback para API
        return new Response(JSON.stringify({
            error: 'offline',
            message: 'You are offline. Some features may be limited.'
        }), {
            status: 503,
            headers: { 'Content-Type': 'application/json' }
        });
    }
}
