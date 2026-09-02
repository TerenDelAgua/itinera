// Sibling of `terms/+page.ts`. This variant is what the
// LegalToggle navigates to (`/terms/es`, `/terms/en`). The path
// segment IS the locale — cleaner URLs than `?lang=` and matches
// the canonical/hreflang URLs from the meta tags.
export const prerender = true;
// Restrict the dynamic segment to our supported legal locales;
// SvelteKit would otherwise crawl any string and 404.
export const entries = () => [{ lang: 'es' }, { lang: 'en' }];