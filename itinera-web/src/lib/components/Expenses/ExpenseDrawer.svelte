<script lang="ts">
  import { fade, fly, slide } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import { SvelteMap, SvelteSet } from "svelte/reactivity";
  import { apiFetch } from "$lib/api";
  import type { Expense } from "$lib/types/Expense";
  import type { Expense_Category as Category } from "$lib/index";
  import ConfirmModal from "../utils/ConfirmModal.svelte";
  import {
    getCurrencySymbol,
    getCategoryEmoji,
    getCategoryName,
  } from "$lib/utils";
  import { t } from "$lib/i18n/store";
  import { COMMON_CURRENCIES } from "$lib/types/Currency";

  let {
    tripId,
    categories,
    isOpen,
    onClose,
    onRefreshSummary,
    placeId,
  }: {
    tripId: string;
    categories: Category[];
    isOpen: boolean;
    onClose: () => void;
    onRefreshSummary: () => void;
    placeId?: string;
  } = $props();

  let expenses = $state<Expense[]>([]);
  let loading = $state(true);
  let editingId = $state<string | null>(null);
  let deleteConfirmId = $state<string | null>(null);
  let draft = $state({
    amount: "",
    date: "",
    notes: "",
    category_id: "",
    currency: "",
  });

  $effect(() => {
    if (isOpen && tripId) loadExpenses();
  });

  async function loadExpenses() {
    loading = true;
    try {
      const endpoint = placeId
        ? `/trips/${tripId}/places/${placeId}/expenses`
        : `/trips/${tripId}/expenses`;
      expenses = await apiFetch<Expense[]>(endpoint);
    } finally {
      loading = false;
    }
  }

  function startEdit(exp: Expense) {
    editingId = exp.id;
    draft = {
      amount: String(exp.original_amount || exp.amount),
      date: exp.date.split("T")[0],
      notes: exp.notes || "",
      category_id: exp.category_id || "",
      currency: exp.original_currency || exp.currency || "EUR",
    };
  }

  async function saveEdit(id: string) {
    if (!draft.amount || parseFloat(draft.amount) <= 0) return;
    const payload = {
      ...draft,
      original_amount: parseFloat(draft.amount),
      original_currency: draft.currency,
      amount: parseFloat(draft.amount), // Fallback if backend doesn't recalculate
      date: new Date(draft.date).toISOString(),
    };
    await apiFetch(`/trips/${tripId}/expenses/${id}`, {
      method: "PUT",
      body: JSON.stringify(payload),
    });
    editingId = null;
    loadExpenses();
    onRefreshSummary();
  }

  function requestDelete(id: string) {
    deleteConfirmId = id;
  }

  async function confirmDeletion() {
    if (!deleteConfirmId) return;
    const id = deleteConfirmId;
    deleteConfirmId = null; // Cierra el modal inmediatamente

    await apiFetch(`/trips/${tripId}/expenses/${id}`, { method: "DELETE" });
    if (editingId === id) editingId = null;
    loadExpenses();
    onRefreshSummary();
  }

  function cancelDeletion() {
    deleteConfirmId = null;
  }

  let collapsedCategories = new SvelteSet<string>();
  function toggleCategory(slug: string) {
    if (collapsedCategories.has(slug)) {
      collapsedCategories.delete(slug);
    } else {
      collapsedCategories.add(slug);
    }
  }

  let grouped = $derived.by(() => {
    const groups = new SvelteMap<string, Expense[]>();
    for (const exp of expenses) {
      const cat = categories.find((c) => c.id === exp.category_id);
      const key = cat ? cat.slug : "others";
      if (!groups.has(key)) groups.set(key, []);
      groups.get(key)!.push(exp);
    }
    return Array.from(groups.entries());
  });
</script>

{#if isOpen}
  <div
    class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/40 backdrop-blur-sm"
    transition:fade={{ duration: 200 }}
    onclick={(e) => e.target === e.currentTarget && onClose()}
  >
    <div
      class="bg-teren-background w-full sm:max-w-2xl h-[85vh] sm:h-[80vh] rounded-t-2xl sm:rounded-2xl shadow-2xl flex flex-col overflow-hidden"
      transition:fly={{ y: 40, duration: 250, easing: cubicOut }}
      onclick={(e) => e.stopPropagation()}
    >
      <header
        class="flex items-center justify-between px-6 py-4 border-b border-teren-border"
      >
        <h2 class="text-lg font-semibold text-teren-text-main tracking-tight">
          {$t("detail.expenses")}
        </h2>
        <button
          onclick={onClose}
          class="text-teren-text-muted hover:text-teren-text-main p-2 rounded-lg hover:bg-gray-100 transition active:scale-95"
          aria-label="Close"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-5 w-5"
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
      </header>

      <main class="flex-1 overflow-y-auto p-6 space-y-6">
        {#if loading}
          <div class="flex justify-center py-10">
            <div
              class="w-8 h-8 border-3 border-teren-primary/30 border-t-teren-primary rounded-full animate-spin"
            ></div>
          </div>
        {:else if expenses.length === 0}
          <div class="text-center py-12 text-teren-text-muted">
            {$t("detail.expenses_empty")}
          </div>
        {:else}
          {#each grouped as [slug, items] (slug)}
            <section class="border-b border-teren-border/50 last:border-0 pb-2">
              <button
                onclick={() => toggleCategory(slug)}
                class="w-full text-left py-3 flex items-center gap-2 group/header focus:outline-none"
              >
                <span class="text-xl select-none">{getCategoryEmoji(slug)}</span
                >
                <h3
                  class="text-sm font-semibold text-teren-text-muted uppercase tracking-wider group-hover/header:text-teren-primary transition-colors"
                >
                  {getCategoryName(slug)}
                </h3>
                <span
                  class="ml-auto font-medium text-xs text-teren-text-muted/60"
                >
                  ({items.length})
                </span>
                <svg
                  class="w-4 h-4 text-teren-text-muted/40 transition-transform duration-200 {collapsedCategories.has(
                    slug,
                  )
                    ? '-rotate-90'
                    : ''}"
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
              </button>

              {#if !collapsedCategories.has(slug)}
                <div
                  class="space-y-3 pb-4"
                  transition:slide={{ duration: 200, easing: cubicOut }}
                >
                  {#each items as exp (exp.id)}
                    {#if editingId === exp.id}
                      <!-- Modo Edición Mejorado -->
                      <div
                        class="p-4 bg-teren-surface border-2 border-teren-primary/30 rounded-xl space-y-4 shadow-sm"
                      >
                        <!-- Fila 1: Categoría, Divisa, Cantidad -->
                        <div class="flex items-stretch gap-2">
                          <!-- Selector de Categoría -->
                          <div
                            class="relative w-16 h-11 bg-white border border-teren-border rounded-lg overflow-hidden focus-within:ring-2 focus-within:ring-teren-primary/30"
                          >
                            <select
                              bind:value={draft.category_id}
                              class="appearance-none w-full h-full bg-transparent text-xl text-center cursor-pointer focus:outline-none"
                            >
                              {#each categories as cat (cat.id)}
                                <option value={cat.id}
                                  >{getCategoryEmoji(cat.slug)}</option
                                >
                              {/each}
                            </select>
                          </div>

                          <!-- Selector de Divisa -->
                          <div
                            class="relative w-20 h-11 bg-white border border-teren-border rounded-lg overflow-hidden focus-within:ring-2 focus-within:ring-teren-primary/30"
                          >
                            <select
                              bind:value={draft.currency}
                              class="appearance-none w-full h-full bg-transparent text-sm font-bold text-teren-text-main text-center cursor-pointer focus:outline-none"
                            >
                              {#each COMMON_CURRENCIES as c (c.code)}
                                <option value={c.code}>{c.code}</option>
                              {/each}
                            </select>
                          </div>

                          <!-- Input de Cantidad -->
                          <div class="flex-1">
                            <input
                              type="number"
                              step="0.01"
                              bind:value={draft.amount}
                              class="w-full h-11 px-3 text-sm font-bold bg-white border border-teren-border rounded-lg focus:ring-2 focus:ring-teren-primary/30 outline-none"
                              autofocus
                            />
                          </div>
                        </div>

                        <!-- Fila 2: Fecha, Notas -->
                        <div class="grid grid-cols-1 sm:grid-cols-3 gap-2">
                          <div class="relative flex items-center group sm:col-span-1">
                            <svg
                              class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-teren-text-muted pointer-events-none"
                              fill="none"
                              viewBox="0 0 24 24"
                              stroke="currentColor"
                              ><path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="1.5"
                                d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                              /></svg
                            >
                            <input
                              type="date"
                              bind:value={draft.date}
                              class="w-full h-11 pl-10 pr-3 text-sm bg-white border border-teren-border rounded-lg focus:ring-2 focus:ring-teren-primary/30 outline-none appearance-none [&::-webkit-calendar-picker-indicator]:opacity-0"
                            />
                          </div>
                          <div class="sm:col-span-2">
                            <input
                              type="text"
                              bind:value={draft.notes}
                              placeholder={$t("detail.notes_optional")}
                              class="w-full h-11 px-4 text-sm bg-white border border-teren-border rounded-lg focus:ring-2 focus:ring-teren-primary/30 outline-none"
                              onkeydown={(e) =>
                                e.key === "Enter" && saveEdit(exp.id)}
                            />
                          </div>
                        </div>

                        <!-- Fila 3: Botones -->
                        <div class="flex justify-end gap-3 pt-2">
                          <button
                            onclick={() => (editingId = null)}
                            class="px-5 py-2 text-sm font-medium text-teren-text-muted hover:text-teren-text-main hover:bg-gray-100 rounded-xl transition"
                          >
                            {$t("common.cancel")}
                          </button>
                          <button
                            onclick={() => saveEdit(exp.id)}
                            class="px-5 py-2 text-sm bg-teren-primary hover:bg-teren-primary-hover text-white font-semibold rounded-xl shadow-sm transition active:scale-95"
                          >
                            {$t("common.save")}
                          </button>
                        </div>
                      </div>
                    {:else}
                      <!-- Vista Compacta (Fecha a la derecha) -->
                      <div
                        class="group p-4 bg-teren-surface border border-teren-border rounded-xl hover:border-teren-primary/20 transition cursor-pointer"
                        onclick={() => startEdit(exp)}
                      >
                        <div class="flex justify-between items-start gap-3">
                          <div class="flex gap-3 flex-1 min-w-0">
                            <!-- Icono con Tooltip -->
                            <span
                              class="text-xl select-none relative group/icon flex-shrink-0"
                            >
                              {getCategoryEmoji(slug)}
                              <span
                                class="absolute bottom-full left-1/2 -translate-x-1/2 mb-1 px-2 py-1 bg-teren-text-main text-white text-xs rounded opacity-0 group-hover/icon:opacity-100 transition pointer-events-none whitespace-nowrap z-10"
                              >
                                {getCategoryName(slug)}
                              </span>
                            </span>

                            <!-- Contenido -->
                            <div class="flex-1 min-w-0">
                              <div class="flex justify-between items-baseline">
                                <span
                                  class="font-bold text-teren-text-main text-base"
                                >
                                  {exp.amount.toFixed(2)}
                                  {getCurrencySymbol(exp.currency)}
                                </span>
                                <span
                                  class="text-xs text-teren-text-muted flex-shrink-0 ml-2"
                                >
                                  {new Date(exp.date).toLocaleDateString(
                                    "en-US",
                                    {
                                      month: "short",
                                      day: "numeric",
                                    },
                                  )}
                                </span>
                              </div>
                              {#if exp.notes}
                                <p
                                  class="text-sm text-teren-text-muted mt-0.5 line-clamp-1 italic opacity-80"
                                >
                                  {exp.notes}
                                </p>
                              {/if}
                            </div>
                          </div>

                          <!-- Botón Delete (solo hover en desktop) -->
                          <button
                            onclick={(e) => {
                              e.stopPropagation();
                              requestDelete(exp.id);
                            }}
                            class="opacity-100 sm:opacity-0 sm:group-hover:opacity-100 text-error-base/70 hover:text-error-base p-1.5 rounded-lg hover:bg-error-subtle transition active:scale-95 flex-shrink-0"
                            aria-label="Delete"
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
                        </div>
                      </div>
                    {/if}
                  {/each}
                </div>
              {/if}
            </section>
          {/each}
        {/if}
      </main>
    </div>
  </div>

  <ConfirmModal
    isOpen={deleteConfirmId !== null}
    title={$t("confirm.delete_expense_title")}
    message={$t("confirm.delete_expense_message")}
    confirmText={$t("common.delete")}
    cancelText={$t("common.cancel")}
    isDestructive={true}
    onConfirm={confirmDeletion}
    onCancel={cancelDeletion}
  />
{/if}
