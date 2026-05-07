import { describe, it, expect, vi, beforeEach } from 'vitest';
import { Events, EventTypes, trackEvent } from './tracking';

// Mock browser environment
vi.mock('$app/environment', () => ({
    browser: true
}));

describe('Tracking Service', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        // Mock global fetch and navigator.sendBeacon
        global.fetch = vi.fn().mockImplementation(() => Promise.resolve({ ok: true }));
        // @ts-ignore
        navigator.sendBeacon = vi.fn().mockReturnValue(true);
        
        // Mock document.cookie
        let cookieStore = '';
        Object.defineProperty(document, 'cookie', {
            get: () => cookieStore,
            set: (v) => { cookieStore = v; },
            configurable: true
        });
    });

    it('should send a landing_viewed event', () => {
        Events.landingView();
        
        expect(navigator.sendBeacon).toHaveBeenCalled();
        const [url, blob] = (navigator.sendBeacon as any).mock.calls[0];
        expect(url).toContain('/events');
        
        // Verify payload
        // @ts-ignore
        blob.text().then((text: string) => {
            const payload = JSON.parse(text);
            expect(payload.type).toBe(EventTypes.LANDING_VIEWED);
            expect(payload.metadata.source).toBeDefined();
        });
    });

    it('should use fetch if sendBeacon is not available', () => {
        // @ts-ignore
        navigator.sendBeacon = undefined;
        
        Events.landingView();
        
        expect(global.fetch).toHaveBeenCalled();
        const [url, options] = (global.fetch as any).mock.calls[0];
        expect(url).toContain('/events');
        expect(options.method).toBe('POST');
    });

    it('should track demo_viewed with correct metadata', () => {
        const tripId = 'trip-123';
        const demoName = 'Japan Trip';
        
        Events.demoViewed(tripId, demoName);
        
        expect(navigator.sendBeacon).toHaveBeenCalled();
        const [, blob] = (navigator.sendBeacon as any).mock.calls[0];
        
        // @ts-ignore
        blob.text().then((text: string) => {
            const payload = JSON.parse(text);
            expect(payload.type).toBe(EventTypes.DEMO_VIEWED);
            expect(payload.trip_id).toBe(tripId);
            expect(payload.metadata.demo_name).toBe(demoName);
        });
    });
});
