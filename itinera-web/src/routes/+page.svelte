<script lang="ts">
  import { onMount } from "svelte";
  import { t } from "$lib/i18n/store";
  import LandingHero from "$lib/components/landing/LandingHero.svelte";
  import DemoCard from "$lib/components/landing/DemoCard.svelte";
  import LandingFooter from "$lib/components/landing/LandingFooter.svelte";
  import { resolve } from "$app/paths";
  import { Events } from "$lib/services/tracking";

  let { data } = $props();

  onMount(() => {
    Events.landingView();
  });
</script>

<svelte:head>
  <title
    >Itinera — {$t("landing.hero_title_plan")}
    {$t("landing.hero_title_frictionless")}</title
  >
  <meta name="description" content={$t("landing.hero_subtitle")} />
</svelte:head>

<div class="landing-page bg-teren-background transition-colors duration-300">
  <!-- Hero Section -->
  <LandingHero totalTrips={data.totalTrips} />

  <!-- Inspiration Section -->
  {#if data.demos && data.demos.length > 0}
    <section
      id="inspiration-grid"
      class="py-24 px-6 bg-teren-surface transition-colors duration-300"
    >
      <div class="max-w-6xl mx-auto">
        <div class="text-center mb-16">
          <h2
            class="text-3xl md:text-4xl font-bold text-teren-text-main tracking-tight mb-4"
          >
            {$t("landing.inspiration_title")}
          </h2>
          <p class="text-teren-text-muted text-lg">
            {$t("landing.inspiration_subtitle")}
          </p>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {#each data.demos as demo (demo.id)}
            <DemoCard {...demo} />
          {/each}
        </div>
      </div>
    </section>
  {/if}

  <!-- How It Works Section -->
  <section
    class="py-24 px-6 bg-teren-background transition-colors duration-300"
  >
    <div class="max-w-4xl mx-auto text-center">
      <div class="text-center mb-16">
        <h2 class="text-3xl md:text-4xl font-bold text-teren-text-main tracking-tight mb-4">
          {$t("landing.how_title")}
        </h2>
        <p class="text-teren-text-muted text-lg">
          {$t("landing.how_subtitle")}
        </p>
      </div>

      <div class="space-y-12 max-w-2xl mx-auto">
        <!-- Step 1 -->
        <div class="flex gap-6 items-start group">
          <div class="flex-shrink-0 w-12 h-12 rounded-2xl bg-teren-primary-subtle text-teren-primary flex items-center justify-center font-bold text-xl border border-teren-primary/10 shadow-sm transition-transform group-hover:scale-110 duration-300">
            1
          </div>
          <div class="space-y-2 pt-1 text-left">
            <h3 class="text-xl font-bold text-teren-text-main">
              {$t("landing.how_step1_title")}
            </h3>
            <p class="text-teren-text-muted leading-relaxed">
              {$t("landing.how_step1_desc")}
            </p>
          </div>
        </div>

        <!-- Step 2 -->
        <div class="flex gap-6 items-start group">
          <div class="flex-shrink-0 w-12 h-12 rounded-2xl bg-teren-primary-subtle text-teren-primary flex items-center justify-center font-bold text-xl border border-teren-primary/10 shadow-sm transition-transform group-hover:scale-110 duration-300">
            2
          </div>
          <div class="space-y-2 pt-1 text-left">
            <h3 class="text-xl font-bold text-teren-text-main">
              {$t("landing.how_step2_title")}
            </h3>
            <p class="text-teren-text-muted leading-relaxed">
              {$t("landing.how_step2_desc")}
            </p>
          </div>
        </div>

        <!-- Step 3 -->
        <div class="flex gap-6 items-start group">
          <div class="flex-shrink-0 w-12 h-12 rounded-2xl bg-teren-primary-subtle text-teren-primary flex items-center justify-center font-bold text-xl border border-teren-primary/10 shadow-sm transition-transform group-hover:scale-110 duration-300">
            3
          </div>
          <div class="space-y-2 pt-1 text-left">
            <h3 class="text-xl font-bold text-teren-text-main">
              {$t("landing.how_step3_title")}
            </h3>
            <p class="text-teren-text-muted leading-relaxed">
              {$t("landing.how_step3_desc")}
            </p>
          </div>
        </div>
      </div>

      <div class="mt-20">
        <a
          href={resolve("/trips")}
          class="inline-flex px-10 py-4 bg-teren-primary hover:bg-teren-primary-hover text-white font-bold rounded-xl shadow-lg shadow-teren-primary/20 transition-all duration-300 hover:-translate-y-0.5 active:scale-95"
        >
          {$t("landing.cta_primary")}
        </a>
      </div>
    </div>
  </section>

  <!-- Footer Section -->
  <LandingFooter />
</div>

