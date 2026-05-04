<script lang="ts">
  import { slide } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import { apiFetch } from "$lib/api";
  import type { Trip } from "$lib/types/Trip";
  import { t } from "$lib/i18n/store";

  let {
    onsuccess,
  }: {
    onsuccess: (trip: Trip) => void;
  } = $props();

  let name = $state("");
  let startDate = $state(new Date().toISOString().split("T")[0]);
  let endDate = $state("");

  let isSubmitting = $state(false);
  let error = $state<string | null>(null);

  function validate(): boolean {
    if (!name.trim()) {
      error = $t("trip_form.error_name_required");
      return false;
    }
    if (!startDate) {
      error = $t("trip_form.error_start_required");
      return false;
    }
    if (endDate && new Date(endDate) <= new Date(startDate)) {
      error = $t("trip_form.error_end_after_start");
      return false;
    }
    return true;
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();
    if (!validate()) return;

    isSubmitting = true;
    error = null;

    try {
      const newTrip = await apiFetch<Trip>("/trips", {
        method: "POST",
        body: JSON.stringify({
          name: name.trim(),
          start_date: startDate,
          end_date: endDate || startDate,
          base_currency: "EUR",
          default_expense_currency: "EUR",
        }),
      });

      onsuccess(newTrip);
    } catch (err) {
      error = err instanceof Error ? err.message : $t("trip_form.create_error");
    } finally {
      isSubmitting = false;
    }
  }

  function handleInput() {
    if (error) error = null;
  }
</script>

<div transition:slide={{ duration: 250, easing: cubicOut }}>
  {#if error}
    <div
      class="mb-2 bg-error-subtle border border-error-base/30 rounded-lg px-3 py-2 text-sm text-error-base font-medium flex items-center gap-2"
      transition:slide={{ duration: 150 }}
    >
      <svg class="w-4 h-4 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10" /><line x1="12" y1="8" x2="12" y2="12" /><line x1="12" y1="16" x2="12.01" y2="16" />
      </svg>
      {error}
    </div>
  {/if}

  <!-- Single container = the orange-bordered form itself -->
  <form
    onsubmit={handleSubmit}
    class="mb-4 bg-teren-surface rounded-xl border-2 border-teren-primary/30 overflow-hidden
           transition-all duration-300 focus-within:border-teren-primary/60 focus-within:shadow-md focus-within:shadow-teren-primary/10"
  >
    <!-- ROW 1: Trip name -->
    <div class="border-b border-teren-border/50">
      <input
        id="trip-name"
        data-testid="trip-name-input"
        type="text"
        bind:value={name}
        oninput={handleInput}
        onkeydown={(e) => e.key === "Enter" && !e.shiftKey && handleSubmit(e)}
        placeholder={$t("trip_form.name_placeholder")}
        class="w-full h-12 px-4 bg-transparent font-bold text-base text-teren-text-main placeholder:text-teren-text-muted/30 focus:outline-none"
      />
    </div>

    <!-- ROW 2: Start date | End date -->
    <div class="flex items-stretch divide-x divide-teren-border border-b border-teren-border/50">
      <!-- Start date -->
      <div class="flex-1 relative flex items-center group min-w-0">
        <span class="absolute left-2.5 text-teren-text-muted group-focus-within:text-teren-primary transition-colors pointer-events-none">
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5" />
          </svg>
        </span>
        <input
          id="start-date"
          type="date"
          bind:value={startDate}
          oninput={handleInput}
          onclick={(e) => e.currentTarget.showPicker()}
          aria-label={$t("trip_form.start_date")}
          class="w-full h-11 pl-8 pr-3 bg-transparent cursor-pointer text-sm text-teren-text-main focus:outline-none tabular-nums min-w-0
                 [&::-webkit-calendar-picker-indicator]:opacity-0 [&::-webkit-calendar-picker-indicator]:absolute [&::-webkit-calendar-picker-indicator]:inset-0 [&::-webkit-calendar-picker-indicator]:w-full [&::-webkit-calendar-picker-indicator]:cursor-pointer"
        />
      </div>

      <!-- End date (no icon, shared with start date icon) -->
      <div class="flex-1 relative flex items-center group min-w-0">
        <input
          id="end-date"
          type="date"
          bind:value={endDate}
          oninput={handleInput}
          onclick={(e) => e.currentTarget.showPicker()}
          aria-label={$t("trip_form.end_date")}
          class="w-full h-11 pl-3 pr-3 bg-transparent cursor-pointer text-sm text-teren-text-main focus:outline-none tabular-nums min-w-0
                 [&::-webkit-calendar-picker-indicator]:opacity-0 [&::-webkit-calendar-picker-indicator]:absolute [&::-webkit-calendar-picker-indicator]:inset-0 [&::-webkit-calendar-picker-indicator]:w-full [&::-webkit-calendar-picker-indicator]:cursor-pointer"
        />
      </div>
    </div>

    <!-- ROW 3: Currency note + submit -->
    <div class="flex items-stretch divide-x divide-teren-border">
      <div class="flex-1 flex items-center px-4 py-3 min-w-0">
        <p class="text-xs text-teren-text-muted leading-snug">
          {$t("trip_form.currency_help", { currency: "EUR" })}
        </p>
      </div>
      <div class="flex-shrink-0 flex items-center px-3 py-2">
        <button
          type="submit"
          disabled={isSubmitting}
          class="w-10 h-10 bg-teren-primary hover:bg-teren-primary-hover text-white rounded-lg shadow-sm active:scale-95 transition-all disabled:opacity-50 flex items-center justify-center"
        >
          {#if isSubmitting}
            <div class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
          {:else}
            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 4v16m8-8H4" />
            </svg>
          {/if}
        </button>
      </div>
    </div>
  </form>
</div>
