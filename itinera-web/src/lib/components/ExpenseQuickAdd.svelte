<script lang="ts">
  import { apiFetch } from "$lib/api";
  import type { Expense, Category } from "$lib/types";
  import { t } from "$lib/i18n/store";
  import { getCategoryEmoji } from "$lib/utils";
  import { COMMON_CURRENCIES } from "$lib/types/Currency";

  let {
    tripId,
    categories,
    onSuccess,
    placeId,
    baseCurrency = "EUR",
    insertionCurrency = "EUR",
  }: {
    tripId: string;
    categories: Category[];
    onSuccess: (exp: Expense) => void;
    placeId?: string;
    baseCurrency?: string;
    insertionCurrency?: string;
  } = $props();

  let amount = $state("");
  let categoryId = $state("");
  let notes = $state("");
  let isSubmitting = $state(false);
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
      currency = effectiveInsertionCurrency;
    } catch (error) {
      console.error("Error al agregar gasto:", error);
    } finally {
      isSubmitting = false;
    }
  }
</script>

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
