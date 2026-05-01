<script lang="ts">
  import { apiFetch } from "$lib/api";
  import { slide } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import { t, locale } from "$lib/i18n/store";
  import { getCurrencySymbol } from "$lib/utils";
  import { formatDisplayDate } from "$lib/utils/date";
  import type { Place } from "$lib/types/Place";

  let {
    tripId,
    places,
    baseCurrency,
    onRefresh,
    onRequestDelete,
  } = $props<{
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
      const payload: any = {
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
    return formatDisplayDate(dateStr, $t, $locale);
  }
</script>

<section>
  <div class="flex justify-between items-center mb-4">
    <h2 class="text-lg font-semibold text-teren-text-main tracking-tight">
      {$t("detail.destinations")}
    </h2>
    <button
      onclick={() => (isCreatingPlace = !isCreatingPlace)}
      class="text-sm font-medium text-teren-primary hover:text-teren-primary-hover transition px-3 py-1.5 rounded-lg hover:bg-teren-primary-subtle active:scale-95"
    >
      + {$t("common.add")}
    </button>
  </div>

  {#if isCreatingPlace}
    <div
      class="mb-4 p-4 bg-teren-background border-2 border-teren-primary/30 rounded-xl space-y-3"
      transition:slide={{ duration: 250, easing: cubicOut }}
    >
      <input
        type="text"
        bind:value={newPlaceDraft.name}
        placeholder={$t("place_form.name_placeholder")}
        class="w-full px-3 py-2 text-sm font-bold bg-white border border-teren-border rounded-lg focus:ring-2 focus:ring-teren-primary/30 outline-none"
        onkeydown={(e) => e.key === "Enter" && createPlace()}
      />
      <div class="flex flex-wrap gap-3 items-end">
        <div class="w-[calc(50%-0.375rem)] sm:flex-1 min-w-0">
          <label
            for="new-place-start-date"
            class="block text-xs text-teren-text-muted mb-1 ml-1 font-medium"
            >{$t("place_form.start_date")}</label
          >
          <input
            id="new-place-start-date"
            type="date"
            bind:value={newPlaceDraft.start_date}
            class="w-full min-w-0 px-2 sm:px-3 py-2 text-sm bg-white border border-teren-border rounded-lg focus:ring-2 focus:ring-teren-primary/30 outline-none"
            onkeydown={(e) => e.key === "Enter" && createPlace()}
          />
        </div>
        <div class="w-[calc(50%-0.375rem)] sm:flex-1 min-w-0">
          <label
            for="new-place-end-date"
            class="block text-xs text-teren-text-muted mb-1 ml-1 font-medium"
            >{$t("place_form.end_date")}</label
          >
          <input
            id="new-place-end-date"
            type="date"
            bind:value={newPlaceDraft.end_date}
            class="w-full min-w-0 px-2 sm:px-3 py-2 text-sm bg-white border border-teren-border rounded-lg focus:ring-2 focus:ring-teren-primary/30 outline-none"
            onkeydown={(e) => e.key === "Enter" && createPlace()}
          />
        </div>

        <div class="w-full h-0 sm:hidden"></div>

        <div class="w-full sm:w-auto flex justify-end">
          <button
            onclick={createPlace}
            disabled={isSaving || !newPlaceDraft.name}
            class="flex-shrink-0 bg-teren-primary hover:bg-teren-primary-hover text-white h-[38px] px-8 sm:px-4 rounded-lg font-bold shadow-sm active:scale-95 transition-all disabled:opacity-50 disabled:active:scale-100 flex items-center justify-center"
          >
            {#if isSaving}
              <div
                class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"
              ></div>
            {:else}
              <svg
                class="w-5 h-5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2.5"
                  d="M12 4v16m8-8H4"
                />
              </svg>
            {/if}
          </button>
        </div>
      </div>
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
          href="/trips/{tripId}/places/{place.id}"
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

            <div class="flex justify-between items-center">
              <div
                class="text-sm text-teren-text-muted flex items-center gap-1.5"
              >
                <svg
                  class="w-4 h-4 opacity-60"
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
                <span
                  >{formatSmartDate(place.start_date)} — {place.end_date
                    ? formatSmartDate(place.end_date)
                    : "No end"}</span
                >
              </div>

              {#if place.total_expenses !== undefined}
                <div class="text-right pl-2">
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
            class="absolute top-3 right-3 opacity-0 group-hover:opacity-100 text-red-400 hover:text-red-600 p-2 rounded-lg hover:bg-red-50 transition active:scale-95"
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
