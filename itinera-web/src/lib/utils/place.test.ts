import { describe, it, expect } from 'vitest';
import { isJapanPlace } from './place';

describe('isJapanPlace', () => {
  // Prioridad: country_code explícito
  it('returns true for explicit JP', () => {
    expect(isJapanPlace({ country_code: 'JP', city: 'Rome' })).toBe(true);
  });

  it('returns false for explicit IT', () => {
    expect(isJapanPlace({ country_code: 'IT', city: 'Tokyo' })).toBe(false);
  });

  it('returns false for explicit FR', () => {
    expect(isJapanPlace({ country_code: 'FR', city: 'Kyoto' })).toBe(false);
  });

  // Fallback por ciudad conocida
  it('returns true for Tokyo (fallback)', () => {
    expect(isJapanPlace({ country_code: null, city: 'Tokyo' })).toBe(true);
  });

  it('returns true for Kyoto (fallback)', () => {
    expect(isJapanPlace({ country_code: null, city: 'Kyoto' })).toBe(true);
  });

  it('returns true for Shinjuku (estación, fallback)', () => {
    expect(isJapanPlace({ country_code: null, city: 'Shinjuku' })).toBe(true);
  });

  it('returns true for Hakata (estación, fallback)', () => {
    expect(isJapanPlace({ country_code: null, city: 'Hakata' })).toBe(true);
  });

  it('returns false for Paris (fallback)', () => {
    expect(isJapanPlace({ country_code: null, city: 'Paris' })).toBe(false);
  });

  it('returns false for Rome (fallback)', () => {
    expect(isJapanPlace({ country_code: null, city: 'Rome' })).toBe(false);
  });

  // Edge cases
  it('returns false for empty city and null country_code', () => {
    expect(isJapanPlace({ country_code: null, city: '' })).toBe(false);
  });

  it('returns false for undefined city and null country_code', () => {
    expect(isJapanPlace({ country_code: null })).toBe(false);
  });

  it('handles uppercase city names', () => {
    expect(isJapanPlace({ country_code: null, city: 'TOKYO' })).toBe(true);
  });

  it('handles mixed case city names', () => {
    expect(isJapanPlace({ country_code: null, city: 'KyOtO' })).toBe(true);
  });

  it('handles city names with whitespace', () => {
    expect(isJapanPlace({ country_code: null, city: '  tokyo  ' })).toBe(true);
  });

  // Fallback por name
  it('returns true if name contains a known city (extreme fallback)', () => {
    expect(isJapanPlace({ country_code: null, city: '', name: 'Tokyo Classic Trip' })).toBe(true);
  });

  it('returns true if name exactly matches a known city', () => {
    expect(isJapanPlace({ country_code: null, name: 'Osaka' })).toBe(true);
  });

  it('does not match if name has no japan cities', () => {
    expect(isJapanPlace({ country_code: null, name: 'Madrid' })).toBe(false);
  });
});
