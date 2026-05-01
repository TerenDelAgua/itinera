<script lang="ts">
  import { apiFetch } from "$lib/api";
  import type { Expense, Category } from "$lib/types";
  import { t, locale } from "$lib/i18n/store";
  import { getCategoryEmoji } from "$lib/utils";
  import { getFriendlyErrorMessage } from "$lib/utils/errorMapper";
  import { fly } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import { COMMON_CURRENCIES } from "$lib/types/Currency";

  let {
    tripId,
    categories,
    onSuccess,
    placeId,
    baseCurrency = "EUR",
    insertionCurrency = "EUR",
    tripStart,
    tripEnd
  }: {
    tripId: string;
    categories: Category[];
    onSuccess: (exp: Expense) => void;
    placeId?: string;
    baseCurrency?: string;
    insertionCurrency?: string;
    tripStart?: string;
    tripEnd?: string;
  } = $props();

  let amount = $state("");
  let categoryId = $state("");
  let notes = $state("");
  let isSubmitting = $state(false);
  let errorMessage = $state<string | null>(null);
  const effectiveInsertionCurrency = $derived(
    insertionCurrency || baseCurrency || "EUR",
  );
  let currency = $state(insertionCurrency);

  $effect(() => {
    if (!categoryId && categories.length > 0) {
      categoryId = categories[0].id;
    }
  });

  $effect(() => {
    if (!amount) {
      currency = effectiveInsertionCurrency;
    }
  });

  async function handleSubmit() {
    if (!amount || parseFloat(amount) <= 0 || !categoryId) return;

    isSubmitting = true;

    try {
      const endpoint = placeId
        ? `/trips/${tripId}/places/${placeId}/expenses`
        : `/trips/${tripId}/expenses`;

      const exp = await apiFetch<Expense>(endpoint, {
        method: "POST",
        body: JSON.stringify({
          amount: parseFloat(amount),
          original_currency: currency,
          category_id: categoryId,
          date: new Date().toISOString(),
          notes: notes.trim(),
          place_id: placeId,
        }),
      });

      onSuccess(exp);
      amount = "";
      notes = "";
      errorMessage = null;
      currency = effectiveInsertionCurrency;
    } catch (error) {
      const formatOptions: Intl.DateTimeFormatOptions = { month: 'short', day: 'numeric', timeZone: 'UTC' };
      const parseYMD = (s: string) => {
        const [y, m, d] = s.split('-').map(Number);
        return new Date(Date.UTC(y, m - 1, d));
      };
      
      const startFormatted = tripStart ? new Intl.DateTimeFormat($locale, formatOptions).format(parseYMD(tripStart)) : "...";
      const endFormatted = tripEnd ? new Intl.DateTimeFormat($locale, formatOptions).format(parseYMD(tripEnd)) : "...";

      errorMessage = getFriendlyErrorMessage(error, {
        tripStart: startFormatted,
        tripEnd: endFormatted,
      });
    } finally {
      isSubmitting = false;
    }
  }
</script>

<div class="flex flex-col gap-2">
  {#if errorMessage}
    <div
      class="bg-error-subtle border border-error-base/20 rounded-xl p-3 flex items-start gap-3 shadow-sm"
      transition:fly={{ y: -5, duration: 200, easing: cubicOut }}
    >
      <div
        class="flex-shrink-0 w-5 h-5 flex items-center justify-center rounded-full bg-error-base/10 text-error-base mt-0.5"
      >
        <svg
          class="w-4 h-4"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
        >
          <circle cx="12" cy="12" r="10" />
          <line x1="12" y1="8" x2="12" y2="12" />
          <line x1="12" y1="16" x2="12.01" y2="16" />
        </svg>
      </div>
      <div class="flex-1 min-w-0">
        <p class="text-sm font-semibold text-error-base leading-tight">
          {errorMessage}
        </p>
      </div>
      <button
        onclick={() => (errorMessage = null)}
        class="text-error-base/40 hover:text-error-base transition-colors p-1 -mr-1"
        aria-label="Dismiss error"
      >
        <svg
          class="w-4 h-4"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
        >
          <path d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
  {/if}

  <div
    class="bg-teren-surface rounded-xl border border-teren-border shadow-sm overflow-hidden transition-all duration-300 focus-within:ring-2 focus-within:ring-teren-primary/30 focus-within:border-teren-primary focus-within:shadow-md"
  >
  <!-- FILA 1: Herramientas Financieras (Categoría | Divisa | Cantidad | Acción) -->
  <div class="flex items-stretch divide-x divide-teren-border">
    <!-- 1. Selector de Categoría -->
    <div class="flex-shrink-0 w-16 relative group">
      <select
        bind:value={categoryId}
        class="appearance-none w-full h-12 bg-transparent text-2xl text-center cursor-pointer focus:outline-none hover:bg-teren-primary-subtle transition-colors"
      >
        {#each categories as cat (cat.id)}
          <option value={cat.id}>{getCategoryEmoji(cat.slug)}</option>
        {/each}
      </select>
    </div>

    <!-- 2. Selector de Divisa -->
    <div class="flex-shrink-0 w-20 relative">
      <select
        bind:value={currency}
        class="appearance-none w-full h-12 bg-transparent text-lg font-bold text-teren-text-main text-center cursor-pointer focus:outline-none hover:bg-teren-primary-subtle transition-colors"
      >
        {#each COMMON_CURRENCIES as c (c.code)}
          <option value={c.code}>{c.symbol}</option>
        {/each}
      </select>
    </div>

    <!-- 3. Input de Cantidad (sin bordes) -->
    <div class="flex-1 min-w-0 px-4">
      <input
        type="number"
        inputmode="decimal"
        step="0.01"
        bind:value={amount}
        placeholder="0.00"
        class="w-full h-12 bg-transparent font-bold text-xl text-teren-text-main placeholder:text-teren-text-muted/30 focus:outline-none"
        onkeydown={(e) => e.key === "Enter" && handleSubmit()}
      />
    </div>
  </div>

  <!-- FILA 2: Notas (separador sutil arriba) -->
  <div class="flex flex-row items-stretch divide-x divide-teren-border">
    <div class="flex-1 h-12 pl-4 border-t border-teren-border/50">
      <input
        type="text"
        bind:value={notes}
        placeholder={$t("expenses.notes_placeholder")}
        class="w-full h-11 px-4 bg-transparent text-sm text-teren-text-muted placeholder:text-teren-text-muted/40 focus:outline-none"
        onkeydown={(e) => e.key === "Enter" && handleSubmit()}
      />
    </div>
    <!-- 4. Botón Añadir (integrado en el widget) -->
    <div class="flex-shrink-0 h-12 pl-4 pr-4">
      <button
        onclick={handleSubmit}
        disabled={isSubmitting || !amount}
        class="w-10 h-10 bg-teren-primary hover:bg-teren-primary-hover text-white font-semibold rounded-lg shadow-sm active:scale-95 transition-all disabled:opacity-50 flex items-center justify-center"
      >
        {#if isSubmitting}
          <div
            class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"
          ></div>
        {:else}
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-5 w-5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            ><path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2.5"
              d="M12 4v16m8-8H4"
            /></svg
          >
        {/if}
      </button>
    </div>
  </div>
  </div>
</div>
