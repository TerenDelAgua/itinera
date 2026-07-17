/**
 * SEO utilities — pure functions only (no DOM access).
 * Safe to use in both SSR and client contexts.
 */

/**
 * Mapping from BCP 47 language tag → Open Graph locale format.
 * OG requires the underscore-separated form with uppercase region.
 * https://ogp.me/#optional
 */
export type SupportedLocale = "en" | "es" | "ja" | "id";

const LOCALE_TO_OG_LOCALE: Record<SupportedLocale, string> = {
  en: "en_US",
  es: "es_ES",
  ja: "ja_JP",
  id: "id_ID",
};

export function localeToOgLocale(locale: string): string {
  if (locale in LOCALE_TO_OG_LOCALE) {
    return LOCALE_TO_OG_LOCALE[locale as SupportedLocale];
  }
  // Safe fallback: do not let an unknown locale produce an invalid og:locale tag.
  return "en_US";
}

/**
 * Mapping from BCP 47 language tag → HTML lang attribute value.
 * 2-letter codes are sufficient and unambiguous for the languages we support.
 */
export function localeToHtmlLang(locale: string): string {
  if (locale in LOCALE_TO_OG_LOCALE) {
    return locale;
  }
  return "en";
}

/**
 * Resolves an absolute URL for an asset path.
 * If the input is already absolute (starts with http:// or https://) it is returned as-is.
 * Otherwise, the page origin is prepended.
 *
 * - `origin` should be `page.url.origin` (or empty string on SSR pre-render).
 * - `path` is typically the asset path returned by SeoHead defaults (e.g. "/og-default.png").
 */
export function resolveAbsoluteUrl(origin: string, path: string): string {
  if (!path) return "";
  if (path.startsWith("http://") || path.startsWith("https://")) {
    return path;
  }
  // Strip query params for canonical URLs.
  const cleanPath = path.split("?")[0].split("#")[0];
  // Ensure exactly one slash between origin and path.
  const separator = cleanPath.startsWith("/") ? "" : "/";
  return `${origin}${separator}${cleanPath}`;
}

/**
 * Builds the full page title by appending the brand suffix once.
 * Prevents accidental double-suffixing ("Foo | Itinera | Itinera").
 */
export function buildPageTitle(title: string, brand = "Itinera"): string {
  const trimmed = title.trim();
  if (!trimmed) return brand;
  if (trimmed.endsWith(`| ${brand}`) || trimmed === brand) {
    return trimmed;
  }
  return `${trimmed} | ${brand}`;
}
