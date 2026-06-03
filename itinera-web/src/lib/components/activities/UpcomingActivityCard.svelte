<script lang="ts">
  import type { Activity } from "$lib/types/Activity";
  import { fly } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import { getRelativeDateLabel } from "$lib/utils";
  import ActivityQuickAdd from "./ActivityQuickAdd.svelte";
  import { t, locale } from "$lib/i18n/store";

  let {
    activities,
    tripId,
    tripStart,
    tripEnd,
    defaultDate,
    placeId,
    city,
    places,
    onRefresh,
    onOpenDrawer,
  } = $props<{
    activities: Activity[];
    tripId: string;
    tripStart?: string;
    tripEnd?: string;
    defaultDate?: string;
    placeId?: string;
    city?: string;
    places?: any[];
    onRefresh: () => void;
    onOpenDrawer: () => void;
  }>();

  import ClimateBadge from "$lib/components/japan-context/ClimateBadge.svelte";
  import { getClimate, shouldShowClimate, type ClimateDisplay } from "$lib/services/japanContext";
  import { isJapanPlace } from "$lib/utils/place";

  let showQuickAdd = $state(false);
  let upcomingClimate = $state<ClimateDisplay | null>(null);

  let upcoming = $derived.by(() => {
    const today = new Date().toISOString().split("T")[0];
    return activities
      .filter((a: Activity) => a.date >= today)
      .sort((a: Activity, b: Activity) => {
        if (a.date !== b.date) return a.date.localeCompare(b.date);
        if (a.time && b.time) return a.time.localeCompare(b.time);
        return a.time ? -1 : 1; // All-day al final
      })[0];
  });

  $effect(() => {
    let resolvedCity = city;
    let matchedPlace = null;
    if (upcoming && !resolvedCity && places) {
      matchedPlace = places.find((p: any) => p.start_date && p.end_date && upcoming.date >= p.start_date.split('T')[0] && upcoming.date <= p.end_date.split('T')[0]);
      if (matchedPlace) {
        resolvedCity = matchedPlace.city;
      }
    }
    const checkPlace = matchedPlace || (placeId && places?.find((p: any) => p.id === placeId)) || { city: resolvedCity };
    if (upcoming && resolvedCity && isJapanPlace(checkPlace) && shouldShowClimate(upcoming.date, tripStart, tripEnd)) {
      getClimate(resolvedCity, upcoming.date).then(c => {
        upcomingClimate = c;
      });
    } else {
      upcomingClimate = null;
    }
  });
</script>

<div
  class="bg-teren-card p-5 rounded-xl border border-teren-border shadow-sm flex flex-col gap-4"
  transition:fly={{ y: -10, duration: 250, easing: cubicOut }}
>
  <!-- Header: Título + ADD (Top-Right) -->
  <div
    class="flex justify-between items-center pb-2 border-b border-teren-border/50"
  >
    <h3 class="text-lg font-semibold text-teren-text-main">
      {$t("itinerary.title")}
    </h3>
    <button
      data-testid="add-activity-btn"
      onclick={() => (showQuickAdd = !showQuickAdd)}
      class="text-sm font-medium text-teren-primary hover:text-teren-primary-hover transition px-3 py-1.5 rounded-lg bg-teren-primary-subtle active:scale-95 flex-shrink-0"
    >
      + {$t("itinerary.add_activity")}
    </button>
  </div>

  <!-- Inline Form (Progressive Disclosure) -->
  {#if showQuickAdd}
    <div
      class="mb-2"
      transition:fly={{ y: -10, duration: 200, easing: cubicOut }}
    >
      <ActivityQuickAdd
        {tripId}
        {tripStart}
        {tripEnd}
        {defaultDate}
        {placeId}
        onSuccess={() => {
          onRefresh();
        }}
      />
    </div>
  {/if}

  <!-- Contenido: Actividad Próxima -->
  {#if upcoming}
    <div class="flex items-start gap-4">
      <div
        class="flex flex-col items-center justify-center w-14 h-14 rounded-lg bg-teren-primary-subtle border border-teren-primary/20 text-teren-primary shrink-0"
      >
        <span class="text-[10px] font-bold uppercase tracking-wide">{$t("itinerary.next_activity")}</span>
        <span class="text-xs font-medium"
          >{getRelativeDateLabel(upcoming.date, $locale)}</span
        >
      </div>
      <div class="flex-1 min-w-0 pt-1">
        <div class="flex items-center gap-2 mb-1">
          <h4 class="text-sm font-semibold text-teren-text-main truncate">
            {upcoming.title}
          </h4>
          {#if upcoming.time}
            <span
              class="text-xs px-2 py-0.5 rounded-full bg-teren-background border border-teren-border text-teren-text-muted tabular-nums ml-auto shrink-0"
            >
              {upcoming.time}
            </span>
          {/if}
        </div>
        {#if upcoming.notes}
          <p class="text-xs text-teren-text-muted italic line-clamp-1 mt-0.5">
            {upcoming.notes}
          </p>
        {/if}
      </div>
    </div>

    <!-- View All (Bottom-Left) & Climate (Bottom-Right) -->
    <div class="flex justify-between items-end mt-1">
      <button
        onclick={onOpenDrawer}
        class="text-sm text-teren-text-muted hover:text-teren-primary transition-colors self-start flex items-center gap-1 group"
      >
        {$t("itinerary.view_all")}
        <svg
          class="w-4 h-4 group-hover:translate-x-1 transition-transform"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M17 8l4 4m0 0l-4 4m4-4H3"
          />
        </svg>
      </button>

      {#if upcomingClimate}
        <ClimateBadge climate={upcomingClimate} />
      {/if}
    </div>
  {:else}
    <!-- Empty State -->
    <div class="text-center py-8 px-4 flex flex-col items-center justify-center border border-dashed border-teren-border rounded-xl bg-teren-background/30">
      <span class="text-3xl mb-3 filter select-none">📋</span>
      <p class="text-sm font-semibold text-teren-text-main">
        {$t("itinerary.no_activities")}
      </p>
      <p class="text-xs text-teren-text-muted mt-1 max-w-sm">
        {$t("itinerary.empty_state_subtitle")}
      </p>
      <button
        onclick={() => (showQuickAdd = true)}
        class="mt-4 inline-flex items-center gap-1.5 px-4 py-2 bg-teren-primary hover:bg-teren-primary-hover text-white text-xs font-semibold rounded-lg shadow-sm shadow-teren-primary/10 active:scale-95 transition-all cursor-pointer"
      >
        + {$t("itinerary.plan_first_activity")}
      </button>
    </div>

    <!-- View All (Bottom-Left) -->
    <button
      onclick={onOpenDrawer}
      class="text-sm text-teren-text-muted hover:text-teren-primary transition-colors self-start flex items-center gap-1 group"
    >
      {$t("itinerary.view_all")}
      <svg
        class="w-4 h-4 group-hover:translate-x-1 transition-transform"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M17 8l4 4m0 0l-4 4m4-4H3"
        />
      </svg>
    </button>
  {/if}
</div>
