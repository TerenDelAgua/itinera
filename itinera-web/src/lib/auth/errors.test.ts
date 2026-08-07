import { describe, it, expect } from 'vitest';
import { resolveErrorMessage, toErrorCode } from './errors';

describe('resolveErrorMessage', () => {
    it('returns the requested locale translation for a known code', () => {
        expect(resolveErrorMessage('EMAIL_ALREADY_EXISTS', 'es'))
            .toBe('Este email ya está en uso. ¿Iniciamos sesión?');
        expect(resolveErrorMessage('RATE_LIMITED', 'en'))
            .toBe('Too many attempts. Please wait a moment.');
        expect(resolveErrorMessage('PASSWORD_TOO_WEAK', 'ja'))
            .toContain('8文字以上');
        expect(resolveErrorMessage('NETWORK_ERROR', 'id'))
            .toContain('Tidak ada koneksi');
    });

    it('falls back to English when the requested locale has no translation', () => {
        // 'fr' is not in the registry. The function should still find a
        // string by falling back to the `en` entry.
        expect(resolveErrorMessage('RATE_LIMITED', 'fr')).toContain('Too many');
    });

    it('returns the supplied fallback when the code is unknown', () => {
        expect(resolveErrorMessage('NOT_A_REAL_CODE', 'es', 'Mi fallback'))
            .toBe('Mi fallback');
    });

    it('returns the generic fallback when both code and fallback are unknown', () => {
        // UNKNOWN_ERROR exists in the registry, so the function returns
        // the localised version for the requested locale, not English.
        expect(resolveErrorMessage('NOT_A_REAL_CODE', 'es')).toBe(
            'Algo ha ido mal. Inténtalo de nuevo.'
        );
    });
});

describe('toErrorCode', () => {
    it('returns the code itself for known entries', () => {
        expect(toErrorCode('EMAIL_ALREADY_EXISTS')).toBe('EMAIL_ALREADY_EXISTS');
        expect(toErrorCode('RATE_LIMITED')).toBe('RATE_LIMITED');
    });

    it('returns UNKNOWN_ERROR for codes not in the registry', () => {
        expect(toErrorCode('SOMETHING_NEW')).toBe('UNKNOWN_ERROR');
        expect(toErrorCode('')).toBe('UNKNOWN_ERROR');
    });
});
