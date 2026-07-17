<script lang="ts">
  import "../app.css";
  import { locale } from "$lib/i18n/store";
  import { resolve } from "$app/paths";
  import { localeToHtmlLang } from "$lib/utils/seo";

  import { themeStore } from "$lib/stores/theme";
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

    // Update local state based on DOM to sync icons
    isDark = document.documentElement.classList.contains("dark");

    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const handler = () => {
      const saved = localStorage.getItem("teren-theme");
      if (!saved) {
        themeStore.init();
      }
      isDark = document.documentElement.classList.contains("dark");
    };
    mq.addEventListener("change", handler);
    return () => mq.removeEventListener("change", handler);
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
        <!-- 
        <div
          class="hidden md:flex items-center gap-2 text-teren-text-muted border-l border-teren-border pl-4 ml-2"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="opacity-70"
            ><path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"></path><circle
              cx="12"
              cy="7"
              r="4"
            ></circle></svg
          >
          <span class="text-sm font-medium">teren_91@hotmail.com</span>
        </div> -->

        <button
          class="text-sm font-semibold text-teren-text-main hover:text-teren-primary transition-all duration-200 flex items-center gap-2 hover:-translate-y-0.5 active:translate-y-0 active:scale-95 ml-2"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            ><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path><polyline
              points="16 17 21 12 16 7"
            ></polyline><line x1="21" y1="12" x2="9" y2="12"></line></svg
          >
          <span class="hidden sm:inline">Salir</span>
        </button>
      </div>
    </div>
  </header>

  <main class="mx-auto flex-1 w-full max-w-6xl px-6 py-12">
    {@render childrenProp?.()}
  </main>
</div>
