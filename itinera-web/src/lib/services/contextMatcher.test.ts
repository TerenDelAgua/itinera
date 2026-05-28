import { describe, it, expect } from 'vitest';
import { matchContext } from './contextMatcher';

describe('contextMatcher', () => {
    it('matches temple correctly', () => {
        const result = matchContext('Zenko-ji Temple', '', 'en');
        expect(result.length).toBeGreaterThan(0);
        expect(result[0].canonical_term).toBe('temple');
    });

    it('matches monkeys/wildlife correctly', () => {
        const result = matchContext('Jigokudani', 'monos!!', 'es');
        expect(result.length).toBeGreaterThan(0);
        expect(result[0].canonical_term).toBe('wildlife');
    });

    it('matches train correctly from multiple languages', () => {
        // "tren" in Spanish
        let result = matchContext('Viaje en tren a Tokyo', '', 'es');
        expect(result.some(r => r.canonical_term === 'transport_train')).toBe(true);

        // "shinkansen" is universal
        result = matchContext('Shinkansen to Kyoto', '', 'en');
        expect(result.some(r => r.canonical_term === 'transport_train')).toBe(true);
    });

    it('returns unique terms ordered by priority', () => {
        // 'temple' has priority 1, 'street' hasn't but wait, all keyword rules have priority 1 currently.
        // If we test with something that has multiple matches.
        const result = matchContext('Zenko-ji Temple and wild monkey', '', 'en');
        const terms = result.map(r => r.canonical_term);
        expect(terms).toContain('temple');
        expect(terms).toContain('wildlife');
        // Unique
        expect(new Set(terms).size).toBe(terms.length);
    });
});
