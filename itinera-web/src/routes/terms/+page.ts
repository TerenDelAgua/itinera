// This file controls the CANONICAL /terms (no [lang] suffix).
// SvelteKit prerenders this with the locale fallback to ES via
// `+page.server.ts`. The /terms/[lang] variant lives in the
// sibling directory `terms/[lang]/+page.ts` and is the
// `?lang=` equivalent used by the LegalToggle.
export const prerender = true;