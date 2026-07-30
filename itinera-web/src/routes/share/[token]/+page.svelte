<script lang="ts">
  import { goto } from "$app/navigation";
  import { t } from "$lib/i18n/store";
  import { apiFetch } from "$lib/api";
  import SeoHead from "$lib/components/seo/SeoHead.svelte";
  import type { Trip } from "$lib/types/Trip";
  import TripCard from "$lib/components/dashboard/TripCard.svelte";

  let { data } = $props<{
    trip?: any;
    token: string;
    shareUrl: string;
    ogImage: string;
    status?: number;
    error?: string;
  }>();

  let isForking = $state(false);
  let forkError = $state<string | null>(null);

  async function useAsBase() {
    isForking = true;
    forkError = null;

    try {
      const newTrip = await apiFetch<Trip>("/share/fork", {
        method: "POST",
        body: JSON.stringify({ token: data.token }),
      });
      goto(`/trips/${newTrip.id}`);
    } catch {
      forkError = $t("share.errors.fork_failed");
      isForking = false;
    }
  }

  let daysUntilExpiry = $derived.by(() => {
    if (!data.trip?.share_expires_at) return null;
    const diff = new Date(data.trip.share_expires_at).getTime() - Date.now();
    return Math.max(0, Math.ceil(diff / (1000 * 60 * 60 * 24)));
  });
</script>

<svelte:head>
  <SeoHead
    title={data.trip ? `${data.trip.name} | Itinera` : "Itinera"}
    description={data.trip?.description || $t("share.seo_fallback")}
    canonical={data.shareUrl}
    ogImage={data.ogImage}
    ogType="article"
  />
</svelte:head>

{#if data.error}
  <div
    class="min-h-screen flex items-center justify-center px-6 bg-teren-background"
  >
    <div class="text-center max-w-md">
      <div class="text-6x1 mb-4">🔒</div>
      <h1 class="text-2x1 font-bold mb-2 text-teren-text-main">
        {$t("share.not_available_title")}
      </h1>
      <p class="mb-6 text-teren-text-muted">{$t(data.error)}</p>
      <button
        onclick={() => goto("/")}
        class="px-6 py-3 rounded-lg font-semibold text-white bg-teren-primary hover:bg-teren-primary-hover transition active:scale-95"
      >
        {$t("share.discover_cta")} ->
      </button>
    </div>
  </div>
{:else if data.trip}
  <div class="min-h-screen bg-teren-background">
    <div
      class="sticky top-0 z-50 border-b border-teren-primary px-6 py-3 flex items-center justify-between gap-4 bg-teren-primary-subtle"
    >
      <div class="flex items-center gap-2 min-w-0">
        <span class="text-lg shrink-0">👁️</span>
        <span class="text-sm font-medium text-teren-text-main truncate"
          >{$t("share.readonly_banner")}</span
        >
        {#if daysUntilExpiry !== null && daysUntilExpiry <= 3}
          <span
            class="text-xs font-medium px-2 py-0.5 rounded-full bg-error-subtle text-error-base shrink-0"
          >
            ⏰ {$t("share.expires_in", { days: daysUntilExpiry })}
          </span>
        {/if}
      </div>
      <button
        onclick={() => goto("/")}
        class="text-sm font-medium text-teren-primary hover:underline shrink-0"
      >
        {$t("share.discover_cta")}
      </button>
    </div>

    <main class="max-w-2x1 mx-auto px-6 py-6 space-y-8">
      <header>
        <h1 class="text 3x1 font-bold tracking-tight text-teren-text-main mb-2">
          {data.trip.name}
        </h1>
        {#if data.trip.description}
          <p class="text-lg text-teren-text-muted">{data.trip.description}</p>
        {/if}
        <div
          class="flex flex-wrap items-center gap-4 mt-4 text-sm text-teren-text-muted"
        >
          {#if data.trip.start_date && data.trip.end_date}
            <span
              >💰 {data.trip.expense_summary?.grand_total?.toFixed(2) || "0.00"}
              {data.trip.base_currency}</span
            >
          {/if}
        </div>
      </header>

      {#if data.trip.places?.length > 0}
        <section>
          <h2 class="text-xl font-bold text-teren-text-main mb-4">
            📍 {$t("share.places_title")}
          </h2>
          <div class="space-y-3">
            {#each data.trip.places as place (place.id)}
              <div
                class="rounded-xl border border-teren-border bg-teren-surface p-4"
              >
                <h3 class="font-semibold text-teren-text-main">{place.name}</h3>
                {#if place.notes}
                  <p class="text-sm text-teren-text-muted mt-1">
                    {place.notes}
                  </p>
                {/if}
              </div>
            {/each}
          </div>
        </section>
      {/if}

      {#if data.trip.activities?.length > 0}
        <section>
          <h2 class="text-xl font-bold text-teren-text-main mb-4">
            📅 {$t("share.activities_title")}
          </h2>
          <ul class="space-y-2">
            {#each data.trip.activities as activity (activity.id)}
              <li class="text-sm text-teren-text-main flex items-center gap-2">
                <span class="opacity-60"
                  >{activity.date}{activity.time
                    ? ` ${activity.time}`
                    : ""}</span
                >
                <span>·</span>
                <span>{activity.title}</span>
              </li>
            {/each}
          </ul>
        </section>
      {/if}

      <div class="sticky bottom-6">
        <div
          class="rounded-xl border border-teren-border bg-teren-surface p-5 shadow-lg"
        >
          <p class="font-semibold text-teren-text-main mb-3">
            {$t("share.fork_prompt")}
          </p>
          {#if forkError}
            <p class="text-sm text-error-base mb-3">{forkError}</p>
          {/if}
          <div class="flex flex-col sm:flex-row gap-3">
            <button
              onclick={useAsBase}
              disabled={isForking}
              class="flex-1 px-5 py-3 rounded-lg font-semibold text-white bg-teren-primary hover:bg-teren-primary-hover transition active:scale-95 disabled:opacity-50"
            >
              {isForking ? "⏳" : "✨"}
              {$t("share.fork_cta")}
            </button>
            <button
              onclick={() => goto("/")}
              class="px-5 py-3 rounded-lg font-medium text-teren-text-muted border border-teren-border hover:bg-teren-interactive-hover transition"
            >
              {$t("share.create_own_cta")}
            </button>
          </div>
        </div>
      </div>
    </main>
  </div>
{/if}
