<script lang="ts">
  import type { PageData } from "./$types";
  import LegalToggle from "$lib/components/LegalToggle.svelte";

  let { data }: { data: PageData } = $props();

  // Build-time base URL for canonical/hreflang. Spec 018 §6
  // recommends the production domain for SEO; in dev we fall
  // back to localhost so `pnpm dev` still renders valid meta.
  const BASE_URL = "https://itinera.teren.dev";

  const canonical = `${BASE_URL}/terms`;
  const hreflangEs = `${BASE_URL}/terms/es`;
  const hreflangEn = `${BASE_URL}/terms/en`;
</script>

<svelte:head>
  <title>{data.doc.meta.title} · Itinera</title>
  <meta name="description" content={data.doc.meta.description} />

  <link rel="canonical" href={canonical} />
  <link rel="alternate" hreflang="es" href={hreflangEs} />
  <link rel="alternate" hreflang="en" href={hreflangEn} />
  <link rel="alternate" hreflang="x-default" href={canonical} />

  <meta name="robots" content="index, follow" />
  <meta property="og:type" content="article" />
  <meta property="og:title" content={data.doc.meta.title} />
  <meta property="og:description" content={data.doc.meta.description} />
  <meta name="twitter:card" content="summary" />
</svelte:head>

<!-- Layout's <main> wraps us. Inner container to control prose width. -->
<div class="max-w-[70ch] mx-auto px-4 sm:px-6 py-12 lg:py-16">
  <header class="mb-10">
    <!-- The h1 is provided by the rendered markdown's first line. -->
    <p class="text-sm text-teren-text-muted mb-2">
      <span
        class="inline-block px-2 py-0.5 bg-teren-interactive-hover rounded text-xs font-medium uppercase tracking-wide"
      >
        {data.localeShown === "es" ? "Documento Legal" : "Legal Document"}
      </span>
    </p>
    <div
      class="text-sm text-teren-text-muted mt-4 flex flex-wrap items-center gap-x-4 gap-y-2"
    >
      <span>
        {data.localeShown === "es" ? "Última actualización" : "Last updated"}:
        <time datetime={data.doc.meta.updated}>{data.doc.meta.updated}</time>
        ·
        {data.localeShown === "es" ? "Versión" : "Version"}
        {data.doc.meta.version}
      </span>
      <LegalToggle currentLocale={data.localeShown} />
    </div>
  </header>

  <div class="lg:grid lg:grid-cols-[1fr_220px] lg:gap-12">
    <article class="legal-prose">
      {@html data.doc.html}

      <details class="mt-16 border-t border-teren-border pt-8">
        <summary
          class="cursor-pointer text-sm text-teren-text-muted font-medium hover:text-teren-text-main transition"
        >
          {data.localeShown === "es"
            ? "Historial de versiones"
            : "Version history"}
        </summary>
      </details>
    </article>

    {#if data.doc.toc.length > 0}
      <aside class="hidden lg:block">
        <nav
          class="sticky lg:top-[calc(var(--header-height,64px)+16px)] text-sm text-teren-text-muted"
          aria-label={data.localeShown === "es"
            ? "Índice"
            : "Table of contents"}
        >
          <p class="font-semibold text-teren-text-main mb-3">
            {data.localeShown === "es" ? "En esta página" : "On this page"}
          </p>
          <ul class="space-y-2">
            {#each data.doc.toc as item (item.id)}
              <li>
                <a
                  href="#{item.id}"
                  class="hover:text-teren-primary transition-colors"
                >
                  {item.text}
                </a>
              </li>
            {/each}
          </ul>
        </nav>
      </aside>
    {/if}
  </div>
</div>
