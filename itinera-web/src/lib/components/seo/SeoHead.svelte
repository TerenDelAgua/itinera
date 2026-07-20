<script lang="ts">
  import { page } from "$app/state";
  import { locale, t } from "$lib/i18n/store";
  import {
    buildPageTitle,
    localeToOgLocale,
    resolveAbsoluteUrl,
  } from "$lib/utils/seo";

  /**
   * Centralised SEO meta tags for every route.
   *
   * Render once per page, near the top of the <script> block.
   * Replaces all inline <svelte:head> meta tags previously scattered across routes.
   *
   * Defaults are intentionally safe: any route that omits props still gets
   * a valid, brand-consistent set of tags.
   */
  interface SeoProps {
    /** Page title WITHOUT brand suffix. The component appends " | Itinera". */
    title: string;
    /** Meta description. Defaults to $t("seo.default_description"). */
    description?: string;
    /** Absolute or root-relative URL. Defaults to current pathname (no query/hash). */
    canonical?: string;
    /**
     * Open Graph type. The spec mentions "place" for place detail pages, but the
     * OG spec itself does not define a `place` type — Facebook treats unknown
     * types as "website" and the type becomes informational. We keep the union
     * aligned with the OG spec and let callers choose "website" or "article".
     */
    ogType?: "website" | "article" | "product";
    /** Path or absolute URL to the OG image. Defaults to /og-default.png. */
    ogImage?: string;
    /** If true, emits noindex,nofollow. */
    noindex?: boolean;
    /** If true, adds nofollow to the robots directive. */
    nofollow?: boolean;
    /**
     * Structured data (JSON-LD). Either a single object or an @graph array.
     * Pass `null` (or just omit it) to skip JSON-LD entirely.
     */
    jsonLd?: Record<string, unknown> | Record<string, unknown>[] | null;
    /** Override OG title (rare). */
    ogTitle?: string;
    /** Override OG description (rare). */
    ogDescription?: string;
    /** Twitter card type. */
    twitterCard?: "summary" | "summary_large_image";
  }

  let {
    title,
    description,
    canonical,
    ogType = "website",
    ogImage,
    noindex = false,
    nofollow = false,
    jsonLd,
    ogTitle,
    ogDescription,
    twitterCard = "summary_large_image",
  }: SeoProps = $props();

  const DEFAULT_OG = "/og-default.png";
  const SITE_NAME = "Itinera";

  // Reactive derivations — re-evaluate on locale/page/url changes.
  let resolvedDescription = $derived(
    description ?? $t("seo.default_description"),
  );

  let resolvedCanonical = $derived.by(() => {
    if (canonical) return resolveAbsoluteUrl(page.url.origin, canonical);
    // Strip query/hash for canonical — same content should always resolve to one URL.
    return page.url.origin + page.url.pathname;
  });

  let resolvedOgImage = $derived(
    resolveAbsoluteUrl(page.url.origin, ogImage ?? DEFAULT_OG),
  );

  let fullTitle = $derived(buildPageTitle(title, SITE_NAME));

  let ogLocale = $derived(localeToOgLocale($locale));

  let robotsContent = $derived(
    noindex
      ? `noindex,${nofollow ? "nofollow" : "follow"}`
      : `index,${nofollow ? "nofollow" : "follow"}`,
  );

  let jsonLdString = $derived.by(() => {
    if (!jsonLd) return null;
    // JSON.stringify with no whitespace keeps the HTML smaller; Google parses either form.
    try {
      return JSON.stringify(jsonLd);
    } catch {
      // Never let a malformed JSON-LD break the page render.
      return null;
    }
  });
</script>

<svelte:head>
  <title>{fullTitle}</title>
  <meta name="description" content={resolvedDescription} />
  <link rel="canonical" href={resolvedCanonical} />
  <meta name="robots" content={robotsContent} />

  <!-- Open Graph -->
  <meta property="og:title" content={ogTitle ?? title} />
  <meta property="og:description" content={ogDescription ?? resolvedDescription} />
  <meta property="og:type" content={ogType} />
  <meta property="og:url" content={resolvedCanonical} />
  <meta property="og:image" content={resolvedOgImage} />
  <meta property="og:site_name" content={SITE_NAME} />
  <meta property="og:locale" content={ogLocale} />

  <!-- Twitter Cards -->
  <meta name="twitter:card" content={twitterCard} />
  <meta name="twitter:title" content={ogTitle ?? title} />
  <meta name="twitter:description" content={ogDescription ?? resolvedDescription} />
  <meta name="twitter:image" content={resolvedOgImage} />

  <!-- Structured Data -->
  {#if jsonLdString}
    {@html `<script type="application/ld+json">${jsonLdString}</script>`}
  {/if}
</svelte:head>
