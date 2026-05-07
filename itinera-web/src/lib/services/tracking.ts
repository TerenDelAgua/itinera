import { browser } from "$app/environment";

const API_BASE = import.meta.env.VITE_API_URL ?? "http://localhost:8080/api/v1";
const EVENT_ENDPOINT = `${API_BASE}/events`;
const SESSION_KEY = "itinera_session";


export const EventTypes = {
    LANDING_VIEWED: 'landing_viewed',
    DEMO_VIEWED: 'demo_viewed',
    DEMO_DEEP_FORKED: 'demo_deep_forked',
    TRIP_CREATED: 'trip_created',
    SESSION_STARTED: 'session_started'
} as const;

export type EventType = typeof EventTypes[keyof typeof EventTypes];

function getSessionId(): string {
    if (!browser) return 'ssr_unknown';

    // Simple cookie read
    const match = document.cookie.match(new RegExp('(^| )' + SESSION_KEY + '=([^;]+)'));
    if (match) return match[2];

    const newSession = 'sess_' + crypto.randomUUID();

    //30 persistency days
    const maxAge = 30 * 24 * 60 * 60;
    document.cookie = `${SESSION_KEY}=${newSession}; path=/; max-age=${maxAge}; SameSite=Lax; Secure`;

    return newSession;
}

export function trackEvent(
    type: EventType,
    metadata: Record<string, unknown> = {},
    tripId: string | null = null
): void {
    if (!browser) return;

    const payload = {
        type,
        trip_id: tripId,
        metadata: {
            ...metadata,
            'client.timestamp': Date.now(),
            'client.platform': navigator.platform
        }
    };

    const blob = new Blob([JSON.stringify(payload)], { type: 'application/json' });

    if (navigator.sendBeacon) {
        navigator.sendBeacon(EVENT_ENDPOINT, blob);
    } else {
        fetch(EVENT_ENDPOINT, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
            keepalive: true
        }).catch(() => { });
    }
}

export const Events = {
    landingView: () => trackEvent(EventTypes.LANDING_VIEWED, { source: document.referrer || 'direct' }),

    sessionStarted: (source: string) =>
        trackEvent(EventTypes.SESSION_STARTED, { landing_source: source }),

    demoViewed: (tripId: string, name: string) =>
        trackEvent(EventTypes.DEMO_VIEWED, { demo_name: name }, tripId),

    demoForked: (originalId: string, trigger: string) =>
        trackEvent(EventTypes.DEMO_DEEP_FORKED, {
            original_demo: originalId,
            trigger
        })
};