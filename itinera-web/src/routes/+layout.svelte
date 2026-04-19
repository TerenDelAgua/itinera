<script lang="ts">
  import "../app.css";
  import { t, locale } from "$lib/i18n/store";

  function toggleLang() {
    locale.update((lang) => (lang === "en" ? "es" : "en"));
  }

  let { children: childrenProp } = $props();
</script>

<div
  class="flex flex-col min-h-screen bg-teren-background text-teren-text-main"
>
  <header class="bg-white border-b border-gray-100 sticky top-0 z-50 shadow-sm">
    <div class="max-w-6xl mx-auto px-6 h-16 flex justify-between items-center">
      <!-- Logo / Marca -->
      <div class="flex items-center gap-3 cursor-pointer group">
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
            <path d="M9 3v15"></path>
            <path d="M15 6v15"></path>
          </svg>
        </div>
        <span class="text-xl font-bold tracking-tight">Itinera</span>
      </div>

      <!-- User Menu -->

      <div class="flex items-center gap-6">
        <div class="flex items-center gap-4">
          <!-- Language Toggle (TEREN Style) -->
          <button
            onclick={toggleLang}
            class="text-sm font-medium text-teren-text-muted hover:text-teren-primary transition px-2 py-1 rounded hover:bg-gray-50 uppercase"
          >
            {$locale}
          </button>
        </div>
        <div
          class="hidden sm:flex items-center gap-2 text-sm text-teren-text-muted font-medium"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="opacity-70"
          >
            <path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"></path>
            <circle cx="12" cy="7" r="4"></circle>
          </svg>
          <span>teren_91@hotmail.com</span>
        </div>

        <button
          class="text-sm font-semibold text-teren-text-main hover:text-teren-primary transition-all duration-200 flex items-center gap-2 hover:-translate-y-0.5 active:translate-y-0 active:scale-95"
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
          >
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
            <polyline points="16 17 21 12 16 7"></polyline>
            <line x1="21" y1="12" x2="9" y2="12"></line>
          </svg>
          <span class="hidden sm:inline">Salir</span>
        </button>
      </div>
    </div>
  </header>

  <main class="mx-auto flex-1 w-full max-w-6xl px-6 py-12">
    {@render childrenProp?.()}
  </main>

  <footer
    class="py-8 text-center text-sm font-medium text-teren-text-muted opacity-70"
  >
    Built with TEREN Design System
  </footer>
</div>
