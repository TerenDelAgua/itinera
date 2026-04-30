<script lang="ts">
  import { t } from "$lib/i18n/store";
  import { apiFetch } from "$lib/api";
  import type { Trip } from "$lib/types/Trip";
  import { fade } from "svelte/transition";
  import CurrencySelector from "$lib/components/currency/CurrencySelector.svelte";


  let {
    onsuccess,
    oncancel,
  }: {
    onsuccess: (trip: Trip) => void;
    oncancel: () => void;
  } = $props();

  let name = $state("");
  let startDate = $state("");
  let endDate = $state("");
  let baseCurrency = $state("EUR");
  let defaultExpenseCurrency = $state("EUR");
  
  let isSubmitting = $state(false);
  let error = $state<string | null>(null);

 $effect(() => {
    if (!startDate) {
      const today = new Date().toISOString().split('T')[0];
      const tomorrow = new Date();
      tomorrow.setDate(tomorrow.getDate() + 1);
      startDate = today;
      endDate = tomorrow.toISOString().split('T')[0];
    }
  });

  async function handleSubmit(e: Event) {
    e.preventDefault();
    if (!name || !startDate || !endDate) return;

    isSubmitting = true;
    error = null;

    try {
      const newTrip = await apiFetch<Trip>("/trips", {
        method: "POST",
        body: JSON.stringify({
          name: name.trim(),
          start_date: startDate,
          end_date: endDate,
          base_currency: baseCurrency,
          default_expense_currency: defaultExpenseCurrency
        }),
      });
      onsuccess(newTrip);
    } catch (err) {
      error = err instanceof Error ? err.message : $t("trip_form.create_error");
    } finally {
      isSubmitting = false;
    }
  }
</script>

<div 
  class="bg-white p-6 sm:p-8 rounded-xl border border-teren-border shadow-sm mb-10 animation-fade-in relative"
  transition:fade={{ duration: 250}}
  >
  <div class="mb-6">
    <h2 class="text-2xl font-bold text-teren-text-main tracking-tight">
      {$t("trip_form.title")}
    </h2>
  </div>

  {#if error}
    <div class="bg-red-50 text-red-700 p-3 rounded-lg text-sm mb-6 border border-red-100">
      {error}
    </div>
  {/if}

  <form onsubmit={handleSubmit} class="space-y-5">
    <div>
      <label for="name" class="block text-sm font-medium text-teren-text-main mb-1.5">
        {$t("trip_form.name")}
      </label>
      <input
        type="text"
        id="name"
        bind:value={name}
        placeholder={$t("trip_form.name_placeholder")}
        required
         class="w-full px-4 py-2.5 bg-surface-base border border-border-subtle rounded-lg text-text-main placeholder:text-text-muted focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary transition-all"
      />
    </div>

    <div class="grid grid-cols-2 gap-4">
      <div>
        <label for="trip-start-date" class="block text-sm font-medium text-teren-text-main mb-1.5">{$t("trip_form.start_date")}</label>
        <input 
          id="trip-start-date"
          type="date" 
          bind:value={startDate} 
          class="w-full px-4 py-2.5 bg-surface-base border border-border-subtle rounded-lg text-text-main placeholder:text-text-muted focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary transition-all" 
        />
      </div>
      <div>
        <label for="trip-end-date" class="block text-sm font-medium text-teren-text-main mb-1.5">{$t("trip_form.end_date")}</label>
        <input 
          id="trip-end-date"
          type="date" 
          bind:value={endDate} 
          class="w-full px-4 py-2.5 bg-white border border-teren-border rounded-lg focus:outline-none focus:ring-2 focus:ring-teren-primary/30 focus:border-teren-primary transition-colors text-teren-text-main" 
        />
      </div>
    </div>

    <div class="bg-surface-base border border-border-subtle rounded-xl p-3 sm:p-4">
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <label class="block">
          <span class="block text-sm font-medium text-teren-text-main mb-1.5">{$t("trip_form.currency")}</span>
          <CurrencySelector
            value={baseCurrency}
            onchange={(code) => {
              if (!code) return;
              const previousBase = baseCurrency;
              baseCurrency = code;
              if (defaultExpenseCurrency === previousBase) {
                defaultExpenseCurrency = code;
              }
            }}
            widthClass="w-full"
          />
        </label>

        <label class="block">
          <span class="block text-sm font-medium text-teren-text-main mb-1.5">{$t("trip_form.expense_currency")}</span>
          <CurrencySelector
            value={defaultExpenseCurrency}
            onchange={(code) => {
              if (code) defaultExpenseCurrency = code;
            }}
            widthClass="w-full"
          />
        </label>
      </div>
    </div>

    <div class="pt-4 flex justify-end gap-3 border-t border-gray-100">
      <button
        type="button"
        onclick={oncancel}
        disabled={isSubmitting}
        class="px-5 py-2.5 text-teren-text-muted hover:text-teren-text-main hover:bg-gray-50 font-medium rounded-lg transition-colors disabled:opacity-50"
    >
        {$t("trip_form.cancel")}
      </button>
      <button
        type="submit"
        disabled={isSubmitting}
        class="px-6 py-2.5 bg-teren-primary hover:bg-teren-primary-hover text-white font-semibold rounded-lg shadow-sm hover:shadow-md transition-all duration-200 active:scale-95 disabled:opacity-70 disabled:active:scale-100 flex items-center gap-2"
      >
        {#if isSubmitting}
          <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
          <span>{$t("trip_form.creating")}</span>
        {:else}
          <span>{$t("trip_form.submit")}</span>
        {/if}
      </button>
    </div>
  </form>
</div>

<style>
  .animation-fade-in {
    animation: fadeIn 0.3s ease-out;
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
      transform: translateY(-10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
</style>
