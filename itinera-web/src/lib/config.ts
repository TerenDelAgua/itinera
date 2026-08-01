/**
 * Itinera runtime configuration.
 *
 * Centralised place for environment-driven constants so components can import
 * from a stable path (`$lib/config`) without having to know whether a value
 * came from `import.meta.env`, a build flag, or a fallback.
 *
 * Values read from `import.meta.env` are typed as `string | undefined` by
 * Vite, hence the `??` fallbacks. The defaults assume:
 *   - local dev runs the web app at http://localhost:5173
 *   - production reads VITE_PUBLIC_APP_URL (defined in the deploy pipeline)
 */

const PUBLIC_APP_URL =
	(import.meta.env.VITE_PUBLIC_APP_URL as string | undefined) ??
	(import.meta.env.VITE_APP_URL as string | undefined) ??
	"http://localhost:5173";

/** Origin used to build absolute URLs (e.g. share links). No trailing slash. */
export const APP_URL = PUBLIC_APP_URL.replace(/\/+$/, "");