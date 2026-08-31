/**
 * Integration test for docs.ts: ensures the TOC extracted from
 * the markdown matches the headings rendered by marked. Spec 018
 * §8 DoD "Test de TOC: para cada doc, los `id` del TOC existen en
 * el HTML renderizado" — this is the test.
 *
 * The import of `docs.ts` will inline-load the four .md files via
 * Vite's `?raw` loader. Vitest needs the same loader behaviour;
 * we configure `vite.config.ts` to enable it. See vite.config.ts
 * `assetsInclude`/`vite-plugin-svelte` config for the `?raw` block.
 */

import { describe, expect, it } from 'vitest';
import { loadDoc } from './docs';

describe('loadDoc', () => {
	describe('TOC consistency', () => {
		it('emits every toc id inside the rendered HTML (terms ES)', async () => {
			const { toc, html } = await loadDoc('terms', 'es');
			expect(toc.length).toBeGreaterThan(0);
			for (const { id } of toc) {
				// The h2 must exist in the HTML with the exact id.
				expect(html).toContain(`id="${id}"`);
			}
		});

		it('emits every toc id inside the rendered HTML (terms EN)', async () => {
			const { toc, html } = await loadDoc('terms', 'en');
			expect(toc.length).toBeGreaterThan(0);
			for (const { id } of toc) {
				expect(html).toContain(`id="${id}"`);
			}
		});

		it('emits every toc id inside the rendered HTML (privacy ES)', async () => {
			const { toc, html } = await loadDoc('privacy', 'es');
			expect(toc.length).toBeGreaterThan(0);
			for (const { id } of toc) {
				expect(html).toContain(`id="${id}"`);
			}
		});

		it('emits every toc id inside the rendered HTML (privacy EN)', async () => {
			const { toc, html } = await loadDoc('privacy', 'en');
			expect(toc.length).toBeGreaterThan(0);
			for (const { id } of toc) {
				expect(html).toContain(`id="${id}"`);
			}
		});
	});

	describe('meta propagation', () => {
		it('exposes the meta from the manifest (terms ES)', async () => {
			const { meta } = await loadDoc('terms', 'es');
			expect(meta.title).toContain('Términos');
			expect(meta.version).toMatch(/^\d+\.\d+$/);
		});

		it('exposes the meta from the manifest (privacy EN)', async () => {
			const { meta } = await loadDoc('privacy', 'en');
			expect(meta.title).toContain('Privacy Policy');
			expect(meta.version).toMatch(/^\d+\.\d+$/);
		});
	});
});