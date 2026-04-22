<script lang="ts">
  import { fade, fly } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import { apiFetch } from '$lib/api';
  import type { Expense, Category } from '$lib/types';

  let {
    tripId,
    categories,
    isOpen,
    onClose,
    onRefreshSummary
  }: {
    tripId: string;
    categories: Category[];
    isOpen: boolean;
    onClose: () => void;
    onRefreshSummary: () => void;
  } = $props();

  let expenses = $state<Expense[]>([]);
  let loading = $state(true);
  let editingId = $state<string | null>(null);
  let draft = $state({ amount: '', notes: '' });

  // Quick Add state
  let newAmount = $state('');
  let newCategoryId = $state(categories[0]?.id || '');
  let newNotes = $state('');
  let isAdding = $state(false);

  const emojiMap: Record<string, string> = {
    accommodation: '🏨', transport: '🚆', food: '🍔',
    leisure: '🎟️', shopping: '🛍️', others: '📦'
  };

  $effect(() => {
    if (isOpen && tripId) loadExpenses();
  });

  async function loadExpenses() {
    loading = true;
    try {
      const data = await apiFetch<Expense[]>(`/trips/${tripId}/expenses`);
      expenses = data.sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime());
    } finally {
      loading = false;
    }
  }

  function getCategoryForExpense(exp: Expense) {
    return categories.find(c => c.id === exp.category_id);
  }

  function getSlug(exp: Expense) {
    return getCategoryForExpense(exp)?.slug ?? 'others';
  }

  function startEdit(exp: Expense) {
    editingId = exp.id;
    draft = { amount: String(exp.amount), notes: exp.notes || '' };
  }

  function cancelEdit() {
    editingId = null;
  }

  async function saveEdit(id: string) {
    const parsed = parseFloat(draft.amount);
    if (!parsed || parsed <= 0) return;
    try {
      await apiFetch(`/expenses/${id}`, {
        method: 'PUT',
        body: JSON.stringify({ amount: parsed, notes: draft.notes })
      });
      editingId = null;
      loadExpenses();
      onRefreshSummary();
    } catch (err) {
      console.error('Update failed:', err);
    }
  }

  async function addExpense() {
    const parsed = parseFloat(newAmount);
    if (!parsed || parsed <= 0) return;
    isAdding = true;
    try {
      await apiFetch(`/trips/${tripId}/expenses`, {
        method: 'POST',
        body: JSON.stringify({
          amount: parsed,
          category_id: newCategoryId,
          date: new Date().toISOString(),
          notes: newNotes,
          currency: 'EUR'
        })
      });
      newAmount = '';
      newNotes = '';
      await loadExpenses();
      onRefreshSummary();
    } catch (err) {
      console.error('Add failed:', err);
    } finally {
      isAdding = false;
    }
  }

  // Group expenses by category slug
  let groupedExpenses = $derived.by(() => {
    const groups = new Map<string, Expense[]>();
    for (const exp of expenses) {
      const slug = getSlug(exp);
      if (!groups.has(slug)) groups.set(slug, []);
      groups.get(slug)!.push(exp);
    }
    return Array.from(groups.entries());
  });

  let totalAmount = $derived(
    expenses.reduce((sum, e) => sum + e.amount, 0)
  );
</script>

{#if isOpen}
  <!-- Backdrop -->
  <div
    class="fixed inset-0 z-50 flex items-end sm:items-center justify-center"
    transition:fade={{ duration: 200 }}
  >
    <!-- Overlay -->
    <button
      class="absolute inset-0 bg-teren-text-main/40 backdrop-blur-[2px] cursor-default"
      onclick={onClose}
      aria-label="Cerrar"
    ></button>

    <!-- Drawer Panel -->
    <div
      class="relative w-full sm:max-w-lg h-[88vh] sm:h-[80vh] bg-teren-background rounded-t-2xl sm:rounded-2xl shadow-2xl flex flex-col overflow-hidden"
      transition:fly={{ y: 48, duration: 280, easing: cubicOut }}
      onclick={(e) => e.stopPropagation()}
    >

      <!-- Handle bar (mobile) -->
      <div class="sm:hidden flex justify-center pt-3 pb-1 flex-shrink-0">
        <div class="w-10 h-1 bg-teren-border rounded-full"></div>
      </div>

      <!-- Header -->
      <header class="flex items-center justify-between px-5 py-4 border-b border-teren-border bg-teren-surface flex-shrink-0">
        <div>
          <h2 class="text-base font-semibold text-teren-text-main tracking-tight">Gastos</h2>
          {#if !loading}
            <p class="text-xs text-teren-text-muted mt-0.5">
              {expenses.length} {expenses.length === 1 ? 'gasto' : 'gastos'} ·
              <span class="font-semibold text-teren-primary">{totalAmount.toFixed(2)} EUR</span>
            </p>
          {/if}
        </div>
        <button
          onclick={onClose}
          class="w-8 h-8 flex items-center justify-center rounded-lg text-teren-text-muted hover:text-teren-text-main hover:bg-teren-border/40 transition-all duration-200 active:scale-95"
          aria-label="Cerrar"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </header>

      <!-- Scrollable content -->
      <main class="flex-1 overflow-y-auto overscroll-contain">
        {#if loading}
          <!-- Skeleton -->
          <div class="p-5 space-y-3">
            {#each [1,2,3] as _}
              <div class="animate-pulse h-16 bg-teren-border/50 rounded-xl"></div>
            {/each}
          </div>

        {:else if expenses.length === 0}
          <div class="flex flex-col items-center justify-center py-16 px-6 text-center">
            <div class="w-14 h-14 rounded-full bg-teren-primary-subtle flex items-center justify-center mb-4">
              <span class="text-2xl">💸</span>
            </div>
            <p class="font-medium text-teren-text-main">Sin gastos aún</p>
            <p class="text-sm text-teren-text-muted mt-1">Añade tu primer gasto abajo.</p>
          </div>

        {:else}
          <div class="p-4 space-y-5">
            {#each groupedExpenses as [slug, items]}
              <section>
                <!-- Category header -->
                <div class="flex items-center gap-2 mb-2 px-1">
                  <span class="text-base">{emojiMap[slug] || '📦'}</span>
                  <span class="text-xs font-semibold text-teren-text-muted uppercase tracking-widest">
                    {slug.charAt(0).toUpperCase() + slug.slice(1)}
                  </span>
                  <div class="flex-1 h-px bg-teren-border"></div>
                  <span class="text-xs text-teren-text-muted">
                    {items.reduce((s, e) => s + e.amount, 0).toFixed(2)} EUR
                  </span>
                </div>

                <!-- Expense rows -->
                <div class="space-y-2">
                  {#each items as exp (exp.id)}
                    {#if editingId === exp.id}
                      <!-- Edit mode: Tap-to-Transform -->
                      <div
                        class="bg-teren-surface border border-teren-primary/30 rounded-xl p-4 shadow-sm space-y-3"
                        transition:fly={{ y: -4, duration: 150, easing: cubicOut }}
                      >
                        <div class="flex gap-3">
                          <div class="flex-1">
                            <label class="text-xs font-medium text-teren-text-muted mb-1.5 block">Importe</label>
                            <div class="relative">
                              <span class="absolute left-3 top-1/2 -translate-y-1/2 text-teren-text-muted font-bold text-sm">€</span>
                              <input
                                type="number"
                                step="0.01"
                                bind:value={draft.amount}
                                autofocus
                                class="w-full pl-7 pr-3 py-2 bg-teren-background border border-teren-border rounded-lg text-teren-text-main font-semibold focus:outline-none focus:ring-2 focus:ring-teren-primary/30 focus:border-teren-primary transition-all text-sm"
                                onkeydown={(e) => {
                                  if (e.key === 'Enter') saveEdit(exp.id);
                                  if (e.key === 'Escape') cancelEdit();
                                }}
                              />
                            </div>
                          </div>
                        </div>
                        <textarea
                          bind:value={draft.notes}
                          placeholder="Nota (opcional)"
                          rows="2"
                          class="w-full px-3 py-2 bg-teren-background border border-teren-border rounded-lg text-teren-text-main text-sm placeholder:text-teren-text-muted/50 resize-none focus:outline-none focus:ring-2 focus:ring-teren-primary/30 focus:border-teren-primary transition-all"
                        ></textarea>
                        <div class="flex justify-end gap-2">
                          <button
                            onclick={cancelEdit}
                            class="px-3 py-1.5 text-sm text-teren-text-muted hover:text-teren-text-main hover:bg-teren-border/40 rounded-lg transition-all"
                          >
                            Cancelar
                          </button>
                          <button
                            onclick={() => saveEdit(exp.id)}
                            class="px-4 py-1.5 text-sm bg-teren-primary hover:bg-teren-primary-hover text-white font-semibold rounded-lg transition-all active:scale-95"
                          >
                            Guardar
                          </button>
                        </div>
                      </div>
                    {:else}
                      <!-- Read mode -->
                      <button
                        class="w-full text-left bg-teren-surface border border-teren-border hover:border-teren-primary/30 rounded-xl px-4 py-3 transition-all duration-200 group active:scale-[0.99]"
                        onclick={() => startEdit(exp)}
                      >
                        <div class="flex items-center justify-between">
                          <div class="flex items-center gap-3">
                            <span class="text-lg select-none">{emojiMap[slug] || '📦'}</span>
                            <div>
                              <p class="font-semibold text-teren-text-main text-sm leading-tight">
                                {exp.amount.toFixed(2)} {exp.currency}
                              </p>
                              {#if exp.notes}
                                <p class="text-xs text-teren-text-muted mt-0.5 line-clamp-1 italic">
                                  "{exp.notes}"
                                </p>
                              {/if}
                              <p class="text-xs text-teren-text-muted/70 mt-0.5">
                                {new Date(exp.date).toLocaleDateString('es-ES', { day: 'numeric', month: 'short', year: 'numeric' })}
                              </p>
                            </div>
                          </div>
                          <!-- Edit hint -->
                          <span class="text-xs text-teren-primary opacity-0 group-hover:opacity-100 transition-opacity duration-200 font-medium">
                            Editar
                          </span>
                        </div>
                      </button>
                    {/if}
                  {/each}
                </div>
              </section>
            {/each}
          </div>
        {/if}
      </main>

      <!-- Quick-add footer -->
      <footer class="border-t border-teren-border bg-teren-surface px-4 py-3 flex-shrink-0">
        <div class="flex items-center gap-2">
          <!-- Category selector -->
          <select
            bind:value={newCategoryId}
            class="text-xl bg-teren-background border border-teren-border rounded-lg px-2 py-2.5 focus:outline-none focus:ring-2 focus:ring-teren-primary/30 focus:border-teren-primary transition-all cursor-pointer appearance-none"
          >
            {#each categories as cat}
              <option value={cat.id}>{emojiMap[cat.slug] || '📦'}</option>
            {/each}
          </select>

          <!-- Amount input -->
          <div class="relative flex-1">
            <span class="absolute left-3 top-1/2 -translate-y-1/2 text-teren-text-muted font-bold text-sm select-none">€</span>
            <input
              type="number"
              inputmode="decimal"
              step="0.01"
              bind:value={newAmount}
              placeholder="0.00"
              class="w-full pl-7 pr-3 py-2.5 bg-teren-background border border-teren-border rounded-lg text-teren-text-main font-bold text-base focus:outline-none focus:ring-2 focus:ring-teren-primary/30 focus:border-teren-primary transition-all placeholder:text-teren-text-muted/40"
              onkeydown={(e) => e.key === 'Enter' && addExpense()}
            />
          </div>

          <!-- Notes input (hidden on very small, visible on sm+) -->
          <input
            type="text"
            bind:value={newNotes}
            placeholder="Nota"
            class="hidden sm:block flex-1 px-3 py-2.5 bg-teren-background border border-teren-border rounded-lg text-teren-text-main text-sm focus:outline-none focus:ring-2 focus:ring-teren-primary/30 focus:border-teren-primary transition-all placeholder:text-teren-text-muted/40"
            onkeydown={(e) => e.key === 'Enter' && addExpense()}
          />

          <!-- Add button -->
          <button
            onclick={addExpense}
            disabled={isAdding || !newAmount || parseFloat(newAmount) <= 0}
            class="w-11 h-11 flex items-center justify-center bg-teren-primary hover:bg-teren-primary-hover text-white rounded-lg shadow-sm transition-all duration-200 active:scale-95 disabled:opacity-40 disabled:cursor-not-allowed disabled:active:scale-100 flex-shrink-0"
          >
            {#if isAdding}
              <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
            {:else}
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
              </svg>
            {/if}
          </button>
        </div>
      </footer>
    </div>
  </div>
{/if}