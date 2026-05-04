<script lang="ts">
  import { apiFetch } from "$lib/api";
  import { slide } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import { t, locale } from "$lib/i18n/store";
  import { getCurrencySymbol } from "$lib/utils";
  import { formatDisplayDate } from "$lib/utils/date";
  import type { Place } from "$lib/types/Place";
  import { resolve } from "$app/paths";

  let { tripId, places, baseCurrency, onRefresh, onRequestDelete } = $props<{
    tripId: string;
    places: Place[];
    baseCurrency: string;
    onRefresh: () => void;
    onRequestDelete: (id: string) => void;
  }>();

  let isCreatingPlace = $state(false);
  let isSaving = $state(false);
  let newPlaceDraft = $state({
    name: "",
    start_date: "",
    end_date: "",
  });

  async function createPlace() {
    if (!newPlaceDraft.name) return;
    isSaving = true;
    try {
      const payload: { name: string; start_date?: string; end_date?: string } = {
        name: newPlaceDraft.name,
      };
      if (newPlaceDraft.start_date)
        payload.start_date = new Date(newPlaceDraft.start_date).toISOString();
      if (newPlaceDraft.end_date)
        payload.end_date = new Date(newPlaceDraft.end_date).toISOString();

      await apiFetch(`/trips/${tripId}/places`, {
        method: "POST",
        body: JSON.stringify(payload),
      });
      newPlaceDraft = { name: "", start_date: "", end_date: "" };
      isCreatingPlace = false;
      onRefresh();
    } catch (err) {
      console.error("Error creating place:", err);
    } finally {
      isSaving = false;
    }
  }

  function formatSmartDate(dateStr?: string) {
    return formatDisplayDate(
      dateStr,
      $t("common.today_short"),
      $t("common.tomorrow_short"),
      $locale,
    );
  }
</script>

<section>
  <div class="flex justify-between items-center mb-4">
    <h2 class="text-lg font-semibold text-teren-text-main tracking-tight">
      {$t("detail.destinations")}
    </h2>
    <button
      onclick={() => (isCreatingPlace = !isCreatingPlace)}
      data-testid="add-place-button"
      class="text-sm font-medium text-teren-primary hover:text-teren-primary-hover transition px-3 py-1.5 rounded-lg hover:bg-teren-primary-subtle active:scale-95"
    >
      + {$t("common.add")}
    </button>
  </div>

  {#if isCreatingPlace}
    <div transition:slide={{ duration: 250, easing: cubicOut }}>
      <!-- Unified widget — same pattern as CreateTripForm -->
      <form
        onsubmit={(e) => { e.preventDefault(); createPlace(); }}
        class="mb-4 bg-teren-surface rounded-xl border-2 border-teren-primary/30 overflow-hidden
               transition-all duration-300 focus-within:border-teren-primary/60 focus-within:shadow-md focus-within:shadow-teren-primary/10"
      >
        <!-- ROW 1: Place name -->
        <div class="border-b border-teren-border/50">
          <input
            type="text"
            bind:value={newPlaceDraft.name}
            placeholder={$t("place_form.name_placeholder")}
            onkeydown={(e) => e.key === "Enter" && createPlace()}
            class="w-full h-12 px-4 bg-transparent font-bold text-base text-teren-text-main placeholder:text-teren-text-muted/30 focus:outline-none"
          />
        </div>

        <!-- ROW 2: Start date | End date -->
        <div class="flex items-stretch divide-x divide-teren-border border-b border-teren-border/50">
          <!-- Start date (with calendar icon) -->
          <div class="flex-1 relative flex items-center group min-w-0">
            <span class="absolute left-2.5 text-teren-text-muted group-focus-within:text-teren-primary transition-colors pointer-events-none">
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5" />
              </svg>
            </span>
            <input
              id="new-place-start-date"
              type="date"
              bind:value={newPlaceDraft.start_date}
              onclick={(e) => e.currentTarget.showPicker()}
              onkeydown={(e) => e.key === "Enter" && createPlace()}
              aria-label={$t("place_form.start_date")}
              class="w-full h-11 pl-8 pr-3 bg-transparent cursor-pointer text-sm text-teren-text-main focus:outline-none tabular-nums min-w-0
                     [&::-webkit-calendar-picker-indicator]:opacity-0 [&::-webkit-calendar-picker-indicator]:absolute [&::-webkit-calendar-picker-indicator]:inset-0 [&::-webkit-calendar-picker-indicator]:w-full [&::-webkit-calendar-picker-indicator]:cursor-pointer"
            />
          </div>

          <!-- End date (no icon) -->
          <div class="flex-1 relative flex items-center group min-w-0">
            <input
              id="new-place-end-date"
              type="date"
              bind:value={newPlaceDraft.end_date}
              onclick={(e) => e.currentTarget.showPicker()}
              onkeydown={(e) => e.key === "Enter" && createPlace()}
              aria-label={$t("place_form.end_date")}
              class="w-full h-11 pl-3 pr-3 bg-transparent cursor-pointer text-sm text-teren-text-main focus:outline-none tabular-nums min-w-0
                     [&::-webkit-calendar-picker-indicator]:opacity-0 [&::-webkit-calendar-picker-indicator]:absolute [&::-webkit-calendar-picker-indicator]:inset-0 [&::-webkit-calendar-picker-indicator]:w-full [&::-webkit-calendar-picker-indicator]:cursor-pointer"
            />
          </div>
        </div>

        <!-- ROW 3: Submit -->
        <div class="flex justify-end px-3 py-2">
          <button
            type="submit"
            disabled={isSaving || !newPlaceDraft.name}
            class="w-10 h-10 bg-teren-primary hover:bg-teren-primary-hover text-white rounded-lg shadow-sm active:scale-95 transition-all disabled:opacity-50 flex items-center justify-center"
          >
            {#if isSaving}
              <div class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
            {:else}
              <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 4v16m8-8H4" />
              </svg>
            {/if}
          </button>
        </div>
      </form>
    </div>
  {/if}

  {#if !places || places.length === 0}
    <div
      class="text-center py-12 bg-teren-surface rounded-xl border border-teren-border border-dashed"
    >
      <p class="text-teren-text-muted text-sm">
        {$t("detail.destinations_empty")}
      </p>
      <p class="text-xs text-teren-text-muted mt-1">
        {$t("detail.destinations_empty_sub")}
      </p>
    </div>
  {:else}
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      {#each places as place (place.id)}
        <a
          href={resolve(`/trips/${tripId}/places/${place.id}`)}
          class="group block bg-teren-surface p-5 rounded-xl border border-teren-border hover:border-teren-primary/30 hover:shadow-md transition-all cursor-pointer relative"
        >
          <div class="flex flex-col gap-1">
            <div class="flex items-center gap-2 pr-8">
              <h3
                class="text-lg font-semibold text-teren-text-main group-hover:text-teren-primary-hover transition-colors"
              >
                {place.name}
              </h3>
              {#if place.default_expense_currency}
                <span
                  class="inline-flex items-center px-2 py-0.5 rounded-full bg-teren-primary-subtle text-teren-primary text-[11px] font-bold border border-teren-primary/15"
                >
                  {place.default_expense_currency}
                </span>
              {/if}
            </div>

            <div class="flex justify-between items-start gap-2">
              <div
                class="text-sm text-teren-text-muted flex items-start gap-1.5 min-w-0"
              >
                <svg
                  class="w-4 h-4 flex-shrink-0 opacity-60 mt-0.5"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                  />
                </svg>
                <span class="leading-snug"
                  >{formatSmartDate(place.start_date)} — {place.end_date
                    ? formatSmartDate(place.end_date)
                    : "No end"}</span
                >
              </div>

              {#if place.total_expenses !== undefined}
                <div class="text-right pl-2 flex-shrink-0 whitespace-nowrap">
                  <span
                    class="font-bold {place.total_expenses > 0
                      ? 'text-teren-primary'
                      : 'text-[11px] text-teren-text-muted opacity-50'}"
                  >
                    {place.total_expenses.toFixed(2)}
                    {getCurrencySymbol(baseCurrency)}
                  </span>
                </div>
              {/if}
            </div>
          </div>

          <button
            onclick={(e) => {
              e.preventDefault();
              e.stopPropagation();
              onRequestDelete(place.id);
            }}
            class="absolute top-3 right-3 opacity-100 sm:opacity-0 sm:group-hover:opacity-100 text-error-base/70 hover:text-error-base p-2 rounded-lg hover:bg-error-subtle transition active:scale-95"
            aria-label="Delete place"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="h-5 w-5"
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
        </a>
      {/each}
    </div>
  {/if}
</section>
