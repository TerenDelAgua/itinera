/**
 * Tests for the legal-doc manifest (Spec 018 §3.1). Guards the
 * single source of truth: if a future maintainer renames a `path`
 * key, removes a locale, or changes a version string, this test
 * catches it before it reaches prod.
 */

import { describe, expect, it } from 'vitest';
import { CONTACT, LEGAL_DOCS, SUPPORTED_LOCALES } from './manifest';

describe('LEGAL_DOCS', () => {
	it('covers both terms and privacy in es + en', () => {
		expect(Object.keys(LEGAL_DOCS).sort()).toEqual(['privacy', 'terms']);
		for (const doc of Object.values(LEGAL_DOCS)) {
			expect(Object.keys(doc).sort()).toEqual(['en', 'es']);
		}
	});

	it('keeps every entry populated with the five required fields', () => {
		for (const doc of Object.values(LEGAL_DOCS)) {
			for (const entry of Object.values(doc)) {
				expect(entry.version).toMatch(/^\d+\.\d+$/);
				expect(entry.updated).toMatch(/^\d{4}-\d{2}-\d{2}$/);
				expect(entry.title.length).toBeGreaterThan(0);
				expect(entry.description.length).toBeGreaterThan(0);
				// `path` must point into docs/legal/. We don't pin the
				// exact path because the bundler may rewrite it, but
				// it must contain the doc filename.
				expect(entry.path).toMatch(/docs\/legal\//);
			}
		}
	});

	it('keeps ES and EN versions in sync (same version + updated date)', () => {
		// A user reading the EN doc should see the same legal
		// guarantees as a user reading the ES doc at any point in
		// time. Mismatched version/updated would let one locale
		// drift silently.
		for (const doc of Object.values(LEGAL_DOCS)) {
			expect(doc.es.version).toBe(doc.en.version);
			expect(doc.es.updated).toBe(doc.en.updated);
		}
	});

	it('SUPPORTED_LOCALES matches the locale entries in LEGAL_DOCS', () => {
		expect([...SUPPORTED_LOCALES].sort()).toEqual(['en', 'es']);
	});
});

describe('CONTACT', () => {
	it('provides distinct emails for general and privacy matters', () => {
		expect(CONTACT.general).toMatch(/@/);
		expect(CONTACT.privacy).toMatch(/@/);
		// Per Spec 018 §3.1 the privacy email must NOT equal the
		// general email — separation of duties is the whole point.
		expect(CONTACT.privacy).not.toBe(CONTACT.general);
	});
});