/**
 * Tests for the slugifier used by /terms and /privacy. Spec 018
 * §8 DoD "Calidad" mandates:
 *   - Test de TOC: ids exist in HTML.
 *   - Test de anti-colisiones: two h2 with same text → distinct ids.
 *
 * We test `slugBase` (pure) and `slugifyBatch` (counter-aware) here.
 */

import { describe, expect, it } from 'vitest';
import { slugBase, slugifyBatch } from './slugify';

describe('slugBase', () => {
	it('lowercases and replaces spaces with hyphens', () => {
		expect(slugBase('Datos personales')).toBe('datos-personales');
	});

	it('strips diacritics', () => {
		expect(slugBase('Protección de datos')).toBe('proteccion-de-datos');
		expect(slugBase('Información')).toBe('informacion');
	});

	it('collapses repeated hyphens', () => {
		expect(slugBase('Foo  --  Bar')).toBe('foo-bar');
	});

	it('strips leading/trailing whitespace', () => {
		expect(slugBase('  Datos  ')).toBe('datos');
	});

	it('drops punctuation', () => {
		expect(slugBase('1. Responsable del tratamiento')).toBe(
			'1-responsable-del-tratamiento',
		);
	});

	it('preserves digits', () => {
		expect(slugBase('Art. 8 LOPDGDD')).toBe('art-8-lopdgdd');
	});
});

describe('slugifyBatch', () => {
	it('returns the base slug for unique headings', () => {
		const ids = slugifyBatch(['Datos', 'Cookies', 'Seguridad']);
		expect(ids).toEqual(['datos', 'cookies', 'seguridad']);
	});

	it('produces unique ids for repeated headings (anti-colisiones)', () => {
		// Two "Cambios" sections in different parts of the doc must
		// produce "cambios" and "cambios-2". A third would be "-3".
		const ids = slugifyBatch([
			'Introducción',
			'Cambios',
			'Datos',
			'Cambios',
			'Resumen',
			'Cambios',
		]);
		expect(ids).toEqual([
			'introduccion',
			'cambios',
			'datos',
			'cambios-2',
			'resumen',
			'cambios-3',
		]);
		// The set must be unique (no two headings share an id).
		expect(new Set(ids).size).toBe(ids.length);
	});

	it('matches the actual PRIVACY_POLICY.md heading sequence', () => {
		// Snapshot of the real headings in docs/legal/PRIVACY_POLICY.md
		// — this guards against accidental refactors that would
		// re-shape the TOC and break anchor links.
		const headings = [
			'1. Responsable del tratamiento',
			'2. Qué datos tratamos',
			'3. Finalidades y base legal',
			'4. Destinatarios y encargados del tratamiento',
			'5. Conservación de datos',
			'6. Tus derechos (Art. 15-22 RGPD)',
			'7. Seguridad',
			'8. Cookies',
			'9. Menores',
			'10. Cambios en esta política',
			'11. Contacto',
			'12. Ley aplicable y jurisdicción',
		];
		const ids = slugifyBatch(headings);
		expect(ids[0]).toBe('1-responsable-del-tratamiento');
		expect(ids[3]).toBe('4-destinatarios-y-encargados-del-tratamiento');
		expect(ids[5]).toBe('6-tus-derechos-art-15-22-rgpd');
	});

	it('preserves order — earlier duplicates get lower indices', () => {
		const ids = slugifyBatch(['A', 'A', 'A', 'B', 'A']);
		expect(ids).toEqual(['a', 'a-2', 'a-3', 'b', 'a-4']);
	});
});