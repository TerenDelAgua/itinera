/**
 * Wrapper around `marked@18.0.9` with options and renderer pinned.
 *
 * Spec 018 §3.5: do NOT use `marked` directly anywhere else. Any
 * markdown→HTML in the app goes through this wrapper so a future
 * version bump is detectable by the snapshot tests in §8.
 *
 * Heading ids are pre-computed by docs.ts and passed in via the
 * `headingIds` option. The renderer matches each heading to its
 * pre-computed id by walking the `headingIds` queue in order. This
 * keeps TOC anchors and `<h2 id="...">` in sync without relying on
 * module-level counters.
 */

import { Marked, type MarkedOptions, type RendererObject } from 'marked';

export interface LegalParseOptions extends MarkedOptions {
	headingIds?: string[];
}

const marked = new Marked({
	gfm: true,
	breaks: false,
	pedantic: false,
});

/**
 * Captures the `headingIds` queue for the current `parse()` call.
 * Reset on every parse — multiple sequential parses in tests must
 * not see stale state from a prior one.
 */
let headingQueue: string[] = [];

const renderer: RendererObject = {
	heading({ tokens, depth }) {
		const raw = tokens
			.map((t) => ('text' in t ? (t as { text: string }).text : ''))
			.join('')
			.trim();
		if (depth === 2) {
			// Pop the next id off the queue. docs.ts has already
			// computed collision-aware ids in the right order.
			const id = headingQueue.shift() ?? raw.toLowerCase().replace(/\s+/g, '-');
			return `<h2 id="${id}">${raw}</h2>\n`;
		}
		return `<h${depth}>${raw}</h${depth}>\n`;
	},
};

marked.use({ renderer });

/**
 * Wraps `marked.parse` with our pinned options. Accepts an extra
 * `headingIds` array (one entry per h2 in document order) so the
 * renderer can use the pre-computed ids instead of re-deriving
 * them (which would race the TOC extractor).
 */
export const parse = (
	src: string,
	options?: LegalParseOptions,
): string => {
	headingQueue = options?.headingIds ? [...options.headingIds] : [];
	return marked.parse(src, options) as string;
};