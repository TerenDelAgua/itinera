<script lang="ts">
  import "../app.css";
  import { locale } from "$lib/i18n/store";
  import { resolve } from "$app/paths";
  import { localeToHtmlLang } from "$lib/utils/seo";

  import { themeStore } from "$lib/stores/theme";
  import { auth } from "$lib/stores/auth.svelte";
  import UserMenu from "$lib/components/user-menu/UserMenu.svelte";
  import Footer from "$lib/components/Footer.svelte";
  import { onMount } from "svelte";
  import { page } from "$app/state";

  let isDark = $state(false);

  let logoHref = $derived(
    page.url.pathname.startsWith("/") ? resolve("/") : resolve("/"),
  );

  // Keep <html lang> in sync with the active locale. This is the SEO-i18n
  // strategy for F1: we rely on the lang attribute (which Google respects
  // regardless of JS execution) rather than on hreflang alternates, which
  // would require i18n subpath routing (out of scope for F1).
  $effect(() => {
    if (typeof document !== "undefined") {
      document.documentElement.lang = localeToHtmlLang($locale);
    }
  });

  onMount(() => {
    themeStore.init();

    // Boot-time probe: ask the backend "do I have a session?". For
    // guests this is a 401 and is the EXPECTED outcome (auth store
    // handles it as a no-op). Centralising the call here means
    // /login and /register don't need to re-probe — they can read
    // `auth.isLoggedIn` synchronously.
    void auth.bootstrap();

    // Update local state based on DOM to sync icons
    isDark = document.documentElement.classList.contains("dark");

    // Set --header-height CSS var to the real header height so the
    // legal-page TOC and any other sticky-below-header UI can offset
    // themselves correctly. We measure after layout and re-measure
    // on resize because the header can change height with the
    // viewport (locale/theme switchers wrap on narrow screens).
    const measure = () => {
      const h = document.querySelector("header");
      if (h) {
        document.documentElement.style.setProperty(
          "--header-height",
          `${h.getBoundingClientRect().height}px`,
        );
      }
    };
    measure();
    window.addEventListener("resize", measure);

    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const handler = () => {
      const saved = localStorage.getItem("teren-theme");
      if (!saved) {
        themeStore.init();
      }
      isDark = document.documentElement.classList.contains("dark");
    };
    mq.addEventListener("change", handler);
    return () => {
      mq.removeEventListener("change", handler);
      window.removeEventListener("resize", measure);
    };
  });

  function handleToggleTheme() {
    themeStore.toggle();
    isDark = document.documentElement.classList.contains("dark");
  }

  let { children: childrenProp } = $props();
</script>

<div
  class="flex flex-col min-h-screen bg-teren-background text-teren-text-main transition-colors duration-300"
>
  <header
    class="bg-teren-card border-b border-teren-border sticky top-0 z-50 shadow-sm transition-colors duration-300"
  >
    <div class="max-w-6xl mx-auto px-6 h-16 flex justify-between items-center">
      <!-- Logo -->
      <a
        href={logoHref}
        class="flex items-center gap-3 group no-underline text-inherit"
      >
        <div
          class="w-10 h-10 bg-teren-primary rounded-xl flex items-center justify-center text-white shadow-sm group-hover:bg-teren-primary-hover transition-colors duration-200"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="22"
            height="22"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M9 3L5 6v15l4-3 6 3 4-3V3l-4 3-6-3z"></path>
            <path d="M9 3v15"></path><path d="M15 6v15"></path>
          </svg>
        </div>
        <span class="text-xl font-bold tracking-tight">Itinera</span>
      </a>

      <!-- Menu Derecha -->
      <div class="flex items-center gap-2 sm:gap-4">
        <!-- Idioma -->
        <select
          bind:value={$locale}
          class="bg-transparent text-sm font-bold text-teren-text-muted hover:text-teren-primary transition px-2 py-1 rounded hover:bg-teren-interactive-hover uppercase focus:outline-none cursor-pointer border-none"
        >
          <option value="en" class="bg-teren-surface text-teren-text-main"
            >EN</option
          >
          <option value="es" class="bg-teren-surface text-teren-text-main"
            >ES</option
          >
          <option value="ja" class="bg-teren-surface text-teren-text-main"
            >JA</option
          >
          <option value="id" class="bg-teren-surface text-teren-text-main"
            >ID</option
          >
        </select>

        <!-- Theme Toggle -->
        <button
          onclick={handleToggleTheme}
          id="theme-toggle"
          class="w-10 h-10 flex items-center justify-center rounded-lg text-teren-text-muted hover:text-teren-primary hover:bg-teren-interactive-hover transition-all duration-200 active:scale-95 shadow-sm"
          aria-label="Toggle theme"
        >
          {#if isDark}
            <!-- Sun icon -->
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              ><circle cx="12" cy="12" r="4" /><path
                d="M12 2v2m0 16v2M4.93 4.93l1.41 1.41m11.32 11.32l1.41 1.41M2 12h2m16 0h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41"
              /></svg
            >
          {:else}
            <!-- Moon icon -->
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              ><path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z" /></svg
            >
          {/if}
        </button>
        <UserMenu />
      </div>
    </div>
  </header>

  <main class="flex-1 w-full">
    {@render childrenProp?.()}
  </main>

  <Footer />
</div>
