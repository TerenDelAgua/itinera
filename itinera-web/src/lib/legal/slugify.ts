/**
 * Slugifier for legal-doc headings. 0 dependencies.
 *
 * Produces deterministic ids from heading text. Two h2 headings
 * with the same text get distinct ids: `cambios`, `cambios-2`,
 * `cambios-3`, etc. The disambiguation index is computed by walking
 * the list left-to-right and counting prior occurrences of the same
 * base slug.
 *
 * The base slug is computed by `slugBase` — same algorithm for both
 * the TOC extractor (which uses `slugifyBatch` over the full list
 * of h2 texts) and the marked renderer (which uses `slugBase`
 * directly on each heading's text). As long as both compute from
 * the SAME raw heading text, both produce identical base slugs.
 */

/**
 * Compute the base slug from heading text: lowercase, strip
 * diacritics, drop punctuation, collapse whitespace to `-`,
 * collapse repeated `-`.
 */
export function slugBase(text: string): string {
	return text
		.toLowerCase()
		.trim()
		.normalize('NFD')
		.replace(/[\u0300-\u036f]/g, '') // strip diacritics
		.replace(/[^\w\s-]/g, '')
		.replace(/\s+/g, '-')
		.replace(/-+/g, '-');
}

/**
 * Batched: produce ids for the full sequence of headings. The
 * collision index is computed by walking the list left-to-right
 * and counting prior occurrences of the same base. This is the
 * single source of truth — every consumer should derive their ids
 * from this, not from a counter that might desync.
 */
export function slugifyBatch(texts: string[]): string[] {
	const seen = new Map<string, number>();
	return texts.map((t) => {
		const base = slugBase(t);
		const count = seen.get(base) ?? 0;
		seen.set(base, count + 1);
		return count === 0 ? base : `${base}-${count + 1}`;
	});
}