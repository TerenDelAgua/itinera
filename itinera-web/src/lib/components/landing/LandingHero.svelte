<script lang="ts">
  import { t } from "$lib/i18n/store";
  import PublicCounter from "./PublicCounter.svelte";
  import ValueCard from "./ValueCard.svelte";

  let { totalTrips = 0 } = $props<{ totalTrips?: number }>();

  function scrollToInspiration() {
    document
      .getElementById("inspiration-grid")
      ?.scrollIntoView({ behavior: "smooth" });
  }
</script>

<section
  class="min-h-[90vh] flex flex-col items-center justify-center pt-20 pb-12 relative overflow-hidden"
>
  <!-- Subtle Background Glow -->
  <div
    class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-teren-primary/5 rounded-full blur-[120px] pointer-events-none"
  ></div>

  <div class="max-w-4xl mx-auto text-center px-6 relative z-10">
    <div class="mb-10 animate-in fade-in slide-in-from-top-4 duration-1000">
      <PublicCounter count={totalTrips} />
    </div>

    <h1
      class="text-5xl md:text-7xl font-bold text-teren-text-main tracking-tight leading-[1.1] mb-8 animate-in fade-in slide-in-from-bottom-4 duration-700 delay-100"
    >
      {$t("landing.hero_title_plan")}<br />
      <span class="text-teren-primary"
        >{$t("landing.hero_title_frictionless")}</span
      >
    </h1>

    <p
      class="text-lg md:text-xl text-teren-text-muted max-w-2xl mx-auto mb-12 leading-relaxed animate-in fade-in slide-in-from-bottom-4 duration-700 delay-300"
    >
      {$t("landing.hero_subtitle")}
    </p>

    <div
      class="flex flex-col sm:flex-row items-center justify-center gap-4 mb-20 animate-in fade-in slide-in-from-bottom-4 duration-700 delay-500"
    >
      <a
        href="/trips"
        class="w-full sm:w-auto px-8 py-4 bg-teren-primary hover:bg-teren-primary-hover text-white font-bold rounded-xl shadow-lg shadow-teren-primary/20 transition-all duration-300 hover:-translate-y-0.5 active:scale-95 text-center"
      >
        {$t("landing.cta_primary")}
      </a>

      <button
        onclick={scrollToInspiration}
        class="w-full sm:w-auto px-8 py-4 bg-transparent border border-teren-border text-teren-text-main font-bold rounded-xl hover:bg-teren-interactive-hover transition-all duration-300 active:scale-95"
      >
        {$t("landing.cta_secondary")}
      </button>
    </div>

    <div
      class="grid grid-cols-1 md:grid-cols-3 gap-6 animate-in fade-in zoom-in-95 duration-1000 delay-700"
    >
      <ValueCard
        icon="👤"
        title={$t("landing.value_guest_title")}
        description={$t("landing.value_guest_desc")}
      />
      <ValueCard
        icon="🗺️"
        title={$t("landing.value_offline_title")}
        description={$t("landing.value_offline_desc")}
      />
      <ValueCard
        icon="🚄"
        title={$t("landing.value_japan_title")}
        description={$t("landing.value_japan_desc")}
      />
    </div>
  </div>
</section>
