/**
 * Legal-doc orchestration: locale resolution, doc loading, render,
 * TOC extraction. Reads markdown files directly from `docs/legal/`
 * via Vite `?raw` (no mirrored copy in src/lib/legal/content/,
 * see Spec 018 decision #13).
 */

import { parse as markedParse } from '$lib/markdown';
import { slugifyBatch } from './slugify';
import { LEGAL_DOCS, SUPPORTED_LOCALES, type LegalLocale } from './manifest';

// Static imports so the build fails at compile time if any of the
// four files is missing. Vite resolves these to `string` content
// thanks to the `?raw` query.
import termsES from '../../../../docs/legal/TERMS_OF_SERVICE.md?raw';
import termsEN from '../../../../docs/legal/TERMS_OF_SERVICE_EN.md?raw';
import privacyES from '../../../../docs/legal/PRIVACY_POLICY.md?raw';
import privacyEN from '../../../../docs/legal/PRIVACY_POLICY_EN.md?raw';

const DOC_BY_PATH: Record<string, string> = {
	'../../../../docs/legal/TERMS_OF_SERVICE.md': termsES,
	'../../../../docs/legal/TERMS_OF_SERVICE_EN.md': termsEN,
	'../../../../docs/legal/PRIVACY_POLICY.md': privacyES,
	'../../../../docs/legal/PRIVACY_POLICY_EN.md': privacyEN,
};

export type DocKind = 'terms' | 'privacy';

// Each `meta` is the same shape regardless of locale or doc kind,
// so we accept any of the four literal entries. Indexing with
// `[LegalLocale]` would create a structural union that TS treats
// as unrelated (because `as const` makes each entry a distinct
// literal type). We use a permissive shape: a meta has these
// fields and nothing else.
type LegalMetaShape = {
	version: string;
	updated: string;
	title: string;
	description: string;
	path: string;
};
export type LegalMeta = LegalMetaShape;
export type LegalDoc = {
	html: string;
	toc: { id: string; text: string }[];
	meta: LegalMeta;
};

/**
 * Resolves which legal locale to render given the user's cookie
 * preference and the app's active locale. Cookie takes precedence;
 * if neither is supported, falls back to `es` (the governing
 * language of all our legal docs).
 */
export function resolveLegalLocale(
	cookieLocale: string | undefined,
	appLocale: string | null | undefined,
): LegalLocale {
	if (
		cookieLocale &&
		(SUPPORTED_LOCALES as readonly string[]).includes(cookieLocale)
	) {
		return cookieLocale as LegalLocale;
	}
	if (
		appLocale &&
		(SUPPORTED_LOCALES as readonly string[]).includes(appLocale)
	) {
		return appLocale as LegalLocale;
	}
	return 'es';
}

/**
 * Loads the markdown for `doc`/`locale`, parses it to HTML, and
 * returns the TOC (h2 entries with their collision-safe ids).
 *
 * The slug counter is reset at the start so the TOC ids and the
 * ids rendered into the HTML match exactly. Tests cover both the
 * TOC consistency (TOC ids exist in HTML) and the collision case.
 */
export async function loadDoc(doc: DocKind, locale: LegalLocale): Promise<LegalDoc> {
	const meta = LEGAL_DOCS[doc][locale];
	const md = DOC_BY_PATH[meta.path];
	if (!md) {
		throw new Error(
			`legal doc not bundled. path=${meta.path} (check src/lib/legal/docs.ts DOC_BY_PATH)`,
		);
	}

	// Extract all h2 texts from the markdown (raw, pre-render) and
	// produce a deterministic id list via `slugifyBatch`. The marked
	// renderer in markdown.ts uses the SAME id list (passed via
	// `headingIds`) to populate `<h2 id="...">`, keeping TOC anchors
	// and heading `id`s in sync without relying on module-level
	// counters (which Vite + SvelteKit can give consumers separate
	// instances of).
	const toc: { id: string; text: string }[] = [];
	const h2Regex = /^## (.+)$/gm;
	const texts: string[] = [];
	let m: RegExpExecArray | null;
	while ((m = h2Regex.exec(md)) !== null) {
		texts.push(m[1].trim());
	}
	const ids = slugifyBatch(texts);
	for (let i = 0; i < texts.length; i++) {
		toc.push({ id: ids[i], text: texts[i] });
	}

	const html = markedParse(md, { headingIds: ids });
	return { html, toc, meta };
}