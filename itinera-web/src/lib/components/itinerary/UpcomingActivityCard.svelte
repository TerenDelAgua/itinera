<script lang="ts">
  import type { Activity } from "$lib/types/Activity";
  import { fly } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import { getRelativeDateLabel } from "$lib/utils";
  import ActivityQuickAdd from "./ActivityQuickAdd.svelte";
  import { t } from "$lib/i18n/store";

  let { activities, tripId, tripStart, tripEnd, defaultDate, onRefresh, onOpenDrawer } = $props<{
    activities: Activity[];
    tripId: string;
    tripStart?: string;
    tripEnd?: string;
    defaultDate?: string;
    onRefresh: () => void;
    onOpenDrawer: () => void;
  }>();

  let showQuickAdd = $state(false);

  let upcoming = $derived.by(() => {
    const today = new Date().toISOString().split("T")[0];
    return activities
      .filter((a) => a.date >= today)
      .sort((a, b) => {
        if (a.date !== b.date) return a.date.localeCompare(b.date);
        if (a.time && b.time) return a.time.localeCompare(b.time);
        return a.time ? -1 : 1; // All-day al final
      })[0];
  });
</script>

<div
  class="bg-white p-5 rounded-xl border border-teren-border shadow-sm flex flex-col gap-4"
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
      onclick={() => (showQuickAdd = !showQuickAdd)}
      class="text-sm font-medium text-teren-primary hover:text-teren-primary-hover transition px-3 py-1.5 rounded-lg bg-teren-primary-subtle active:scale-95 flex-shrink-0"
    >
      + {$t("itinerary.add")}
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
        onSuccess={() => {
          showQuickAdd = false;
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
        <span class="text-[10px] font-bold uppercase tracking-wide">Next</span>
        <span class="text-xs font-medium"
          >{getRelativeDateLabel(upcoming.date)}</span
        >
      </div>
      <div class="flex-1 min-w-0 pt-1">
        <div class="flex items-center gap-2 mb-1">
          <h4 class="text-sm font-semibold text-teren-text-main truncate">
            {upcoming.title}
          </h4>
          {#if upcoming.time}
            <span
              class="text-xs px-2 py-0.5 rounded-full bg-teren-background border border-teren-border text-teren-text-muted tabular-nums"
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

    <!-- View All (Bottom-Left) -->
    <button
      onclick={onOpenDrawer}
      class="text-sm text-teren-text-muted hover:text-teren-primary transition-colors self-start flex items-center gap-1 group"
    >
      View All
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
  {:else}
    <!-- Empty State -->
    <div class="text-center py-6">
      <p class="text-sm text-teren-text-muted">No upcoming activities.</p>
    </div>

    <!-- View All (Bottom-Left) -->
    <button
      onclick={onOpenDrawer}
      class="text-sm text-teren-text-muted hover:text-teren-primary transition-colors self-start flex items-center gap-1 group"
    >
      View All
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
