<script lang="ts">
  import { fly, fade, slide } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import { tick } from "svelte";
  import type { Activity } from "$lib/types/Activity";
  import { getRelativeDateLabel } from "$lib/utils";
  import ActivityQuickAdd from "./ActivityQuickAdd.svelte";
  import { SvelteMap, SvelteSet, SvelteDate } from "svelte/reactivity";
  import ConfirmModal from "$lib/components/ConfirmModal.svelte";
  import { activityApi } from "$lib/api/activity";
  import { t } from "$lib/i18n/store";

  let { isOpen, tripId, tripStart, tripEnd, activities = [], defaultDate, onRefresh, onClose } = $props<{
    isOpen: boolean;
    tripId: string;
    tripStart?: string;
    tripEnd?: string;
    activities: Activity[];
    defaultDate?: string;
    onRefresh: () => void;
    onClose: () => void;
  }>();

  let showQuickAdd = $state(false);
  let collapsedSections = $state<Set<string>>(new SvelteSet());
  let todaySection = $state<Element | null>(null);

  let activityToDelete = $state<Activity | null>(null);
  let editingActivityId = $state<string | null>(null);

  async function handleDeleteConfirm() {
    if (!activityToDelete) return;
    try {
      await activityApi.delete(tripId, activityToDelete.id);
      onRefresh();
    } catch (e) {
      console.error("Failed to delete activity", e);
    } finally {
      activityToDelete = null;
    }
  }

  let grouped = $derived.by(() => {
    const map = new SvelteMap<string, Activity[]>();
    const sorted = [...(activities || [])].sort((a, b) => {
      if (a.date !== b.date) return a.date.localeCompare(b.date);
      if (a.time && b.time) return a.time.localeCompare(b.time);
      return a.time ? -1 : 1;
    });
    for (const act of sorted) {
      if (!map.has(act.date)) map.set(act.date, []);
      map.get(act.date)!.push(act);
    }
    return map;
  });

  function isCollapsed(date: string): boolean {
    const today = new Date().toISOString().split("T")[0];
    const tomorrow = new SvelteDate();
    tomorrow.setDate(tomorrow.getDate() + 1);
    const tomorrowStr = tomorrow.toISOString().split("T")[0];

    if (date === today || date === tomorrowStr) return false;
    return collapsedSections.has(date);
  }

  function toggleSection(date: string) {
    if (collapsedSections.has(date)) {
      collapsedSections.delete(date);
    } else {
      collapsedSections.add(date);
    }
  }

  $effect(() => {
    if (isOpen) {
      // 1. Por defecto, colapsar actividades pasadas
      const todayStr = new Date().toISOString().split("T")[0];
      const pastDates = new Set(
        activities.filter((a) => a.date < todayStr).map((a) => a.date),
      );
      collapsedSections = new SvelteSet(pastDates);

      // 2. Scroll a "Today" o la primera sección futura disponible
      tick().then(() => {
        if (todaySection) {
          todaySection.scrollIntoView({ behavior: "smooth", block: "start" });
        } else {
          // Si no hay hoy, buscar la primera sección futura
          const firstFuture = document.querySelector(
            'section[data-future="true"]',
          );
          if (firstFuture) {
            firstFuture.scrollIntoView({ behavior: "smooth", block: "start" });
          }
        }
      });
    }
  });
</script>

{#if isOpen}
  <div
    class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/40 backdrop-blur-sm"
    transition:fade={{ duration: 200 }}
  >
    <div
      class="bg-teren-surface w-full sm:max-w-lg h-[85vh] sm:h-[80vh] rounded-t-2xl sm:rounded-2xl flex flex-col overflow-hidden shadow-2xl"
      transition:fly={{ y: 30, duration: 250, easing: cubicOut }}
    >
      <!-- Header -->
      <div
        class="px-6 py-4 border-b border-teren-border flex justify-between items-center bg-white shrink-0"
      >
        <h2 class="text-lg font-bold text-teren-text-main">Itinerary</h2>
        <button
          onclick={onClose}
          class="text-teren-text-muted hover:text-teren-text-main p-2 rounded-lg hover:bg-teren-background transition"
        >
          <svg
            class="w-5 h-5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            ><path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M6 18L18 6M6 6l12 12"
            /></svg
          >
        </button>
      </div>

      <!-- Scrollable List -->
      <div class="flex-1 overflow-y-auto px-6 py-4 space-y-6">
        {#if activities.length === 0}
          <div class="text-center py-12 text-teren-text-muted">
            <p class="text-sm">Your itinerary is empty.</p>
            <p class="text-xs mt-1 opacity-60">
              Add activities to organize your trip.
            </p>
          </div>
        {:else}
          {#snippet sectionContent(date, acts)}
            <!-- Sticky Header Colapsable -->
            <button
              onclick={() => toggleSection(date)}
              aria-expanded={!isCollapsed(date)}
              class="w-full sticky top-0 bg-teren-surface/95 backdrop-blur-sm py-3 border-b border-teren-border flex justify-between items-center group cursor-pointer focus:outline-none focus:ring-2 focus:ring-teren-primary/30 rounded-lg px-2 -mx-2"
            >
              <span
                class="text-sm font-bold text-teren-text-main tracking-wide"
              >
                {getRelativeDateLabel(date)}
              </span>

              <div class="flex items-center gap-2">
                {#if isCollapsed(date)}
                  <span
                    class="text-xs text-teren-text-muted group-hover:text-teren-primary transition-colors"
                  >
                    {acts.length} activit{acts.length === 1 ? "y" : "ies"}
                  </span>
                {/if}
                <svg
                  class="w-4 h-4 text-teren-text-muted transition-transform duration-200 {isCollapsed(
                    date,
                  )
                    ? '-rotate-90'
                    : 'rotate-0'}"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M19 9l-7 7-7-7"
                  />
                </svg>
              </div>
            </button>

            <!-- Contenido Colapsable -->
            {#if !isCollapsed(date)}
              <div
                class="divide-y divide-teren-border/50 mt-2"
                transition:slide={{ duration: 200, easing: cubicOut }}
              >
                {#each acts as activity (activity.id)}
                  {#if editingActivityId === activity.id}
                    <div class="py-3" transition:slide>
                      <ActivityQuickAdd
                        {tripId}
                        {tripStart}
                        {tripEnd}
                        {activity}
                        onCancel={() => (editingActivityId = null)}
                        onSuccess={() => {
                          editingActivityId = null;
                          onRefresh();
                        }}
                      />
                    </div>
                  {:else}
                    <div
                      onclick={() => (editingActivityId = activity.id)}
                      class="flex items-start gap-4 py-3 hover:bg-teren-background/50 transition-colors cursor-pointer group"
                    >
                      <div class="w-14 text-right shrink-0 pt-0.5">
                        {#if activity.time}
                          <span
                            class="text-xs font-mono font-medium text-teren-text-muted tabular-nums"
                            >{activity.time}</span
                          >
                        {:else}
                          <span
                            class="text-[10px] text-teren-text-muted/50 uppercase tracking-wider"
                            >All day</span
                          >
                        {/if}
                      </div>

                      <div class="flex-1 min-w-0">
                        <h4
                          class="text-sm font-semibold text-teren-text-main leading-snug truncate"
                        >
                          {activity.title}
                        </h4>
                        {#if activity.notes}
                          <p
                            class="text-xs text-teren-text-muted mt-0.5 italic line-clamp-2"
                          >
                            {activity.notes}
                          </p>
                        {/if}
                        {#if activity.place_id}
                          <span
                            class="inline-block mt-1.5 text-[10px] font-medium text-teren-primary bg-teren-primary-subtle px-2 py-0.5 rounded-full"
                          >
                            Local Activity
                          </span>
                        {/if}
                      </div>

                      <!-- Delete Action -->
                      <button
                        onclick={(e) => {
                          e.stopPropagation();
                          activityToDelete = activity;
                        }}
                        class="p-2 text-error-base/70 hover:text-error-base hover:bg-error-subtle rounded-lg transition-colors opacity-100 sm:opacity-0 sm:group-hover:opacity-100"
                        aria-label="Delete activity"
                      >
                        <svg
                          class="w-4 h-4"
                          fill="none"
                          viewBox="0 0 24 24"
                          stroke="currentColor"
                        >
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                          />
                        </svg>
                      </button>
                    </div>
                  {/if}
                {/each}
              </div>
            {/if}
          {/snippet}

          {#each grouped as [date, acts] (date)}
            {@const isToday = getRelativeDateLabel(date) === "Today"}
            {@const isFuture =
              !isToday && date > new Date().toISOString().split("T")[0]}
            {#if isToday}
              <section bind:this={todaySection} data-today="true">
                {@render sectionContent(date, acts)}
              </section>
            {:else}
              <section data-future={isFuture}>
                {@render sectionContent(date, acts)}
              </section>
            {/if}
          {/each}
        {/if}
      </div>

      <!-- Footer: Quick Add -->
      <div class="px-6 py-4 border-t border-teren-border bg-white shrink-0">
        <button
          onclick={() => (showQuickAdd = !showQuickAdd)}
          class="w-full text-center text-sm font-medium text-teren-primary hover:text-teren-primary-hover py-2 transition-colors"
        >
          {showQuickAdd ? "− Hide form" : "+ Add activity"}
        </button>
        {#if showQuickAdd}
          <div
            class="mt-3"
            transition:fly={{ y: 10, duration: 200, easing: cubicOut }}
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
      </div>
    </div>
  </div>
  <ConfirmModal
    isOpen={activityToDelete !== null}
    title={$t("confirm.delete_activity_title") || "Delete Activity"}
    message={$t("confirm.delete_activity_message") ||
      "Are you sure you want to delete this activity?"}
    confirmText={$t("common.delete")}
    cancelText={$t("common.cancel")}
    isDestructive={true}
    onConfirm={handleDeleteConfirm}
    onCancel={() => (activityToDelete = null)}
  />
{/if}
