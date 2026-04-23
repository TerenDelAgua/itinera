<script lang="ts">
  import { SvelteMap } from 'svelte/reactivity';
  import { fade, fly } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import { apiFetch } from '$lib/api';
  import type { Expense, Category } from '$lib/types';

  let { tripId, categories, isOpen, onClose, onRefreshSummary }: { tripId: string; categories: Category[]; isOpen: boolean; onClose: () => void; onRefreshSummary: () => void; } = $props();

  let expenses = $state<Expense[]>([]);
  let loading = $state(true);
  let editingId = $state<string | null>(null);
  let draft = $state({ amount: '', date: '', notes: '', category_id: '' });

  $effect(() => { if (isOpen && tripId) loadExpenses(); });

  async function loadExpenses() {
    loading = true;
    try { expenses = await apiFetch<Expense[]>(`/trips/${tripId}/expenses`); } 
    finally { loading = false; }
  }

  function startEdit(exp: Expense) {
    editingId = exp.id;
    draft = { 
      amount: String(exp.amount), 
      date: exp.date.split('T')[0], 
      notes: exp.notes || '', 
      category_id: exp.category_id || '' 
    };
  }

  async function saveEdit(id: string) {
    if (!draft.amount || parseFloat(draft.amount) <= 0) return;
    const payload = {
      ...draft,
      amount: parseFloat(draft.amount),
      date: new Date(draft.date).toISOString()
    };
    await apiFetch(`/trips/${tripId}/expenses/${id}`, { method: 'PUT', body: JSON.stringify(payload) });
    editingId = null;
    loadExpenses();
    onRefreshSummary();
  }

  async function deleteExpense(id: string) {
    if (!confirm('Delete this expense?')) return;
    await apiFetch(`/trips/${tripId}/expenses/${id}`, { method: 'DELETE' });
    if (editingId === id) editingId = null;
    loadExpenses();
    onRefreshSummary();
  }

  let grouped = $derived.by(() => {
    const groups = new SvelteMap<string, Expense[]>();
    for (const exp of expenses) {
      const cat = categories.find(c => c.id === exp.category_id);
      const key = cat ? cat.slug : 'others';
      if (!groups.has(key)) groups.set(key, []);
      groups.get(key)!.push(exp);
    }
    return Array.from(groups.entries());
  });

  const emojiMap: Record<string, string> = { accommodation: '🏨', transport: '🚆', food: '🍔', leisure: '🎟️', shopping: '🛍️', others: '📦' };
</script>

{#if isOpen}
  <div class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/40 backdrop-blur-sm" transition:fade={{ duration: 200 }} onclick={e => e.target === e.currentTarget && onClose()}>
    <div class="bg-teren-surface w-full sm:max-w-2xl h-[85vh] sm:h-[80vh] rounded-t-2xl sm:rounded-2xl shadow-2xl flex flex-col overflow-hidden" transition:fly={{ y: 40, duration: 250, easing: cubicOut }} onclick={(e) => e.stopPropagation()}>
      <header class="flex items-center justify-between px-6 py-4 border-b border-teren-border">
        <h2 class="text-lg font-semibold text-teren-text-main tracking-tight">Expenses</h2>
        <button onclick={onClose} class="text-teren-text-muted hover:text-teren-text-main p-2 rounded-lg hover:bg-gray-100 transition active:scale-95" aria-label="Close">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
        </button>
      </header>

      <main class="flex-1 overflow-y-auto p-6 space-y-6">
        {#if loading}
          <div class="flex justify-center py-10"><div class="w-8 h-8 border-3 border-teren-primary/30 border-t-teren-primary rounded-full animate-spin"></div></div>
        {:else if expenses.length === 0}
          <div class="text-center py-12 text-teren-text-muted">No expenses yet.</div>
        {:else}
          {#each grouped as [slug, items] (slug)}
            <section>
              <h3 class="text-sm font-semibold text-teren-text-muted uppercase tracking-wider mb-3 flex items-center gap-2">
                {emojiMap[slug] || '📦'} {slug.charAt(0).toUpperCase() + slug.slice(1)}
                <span class="ml-auto font-normal normal-case text-xs">({items.length})</span>
              </h3>
              <div class="space-y-3">
                {#each items as exp (exp.id)}
                  {#if editingId === exp.id}
                    <div class="p-4 bg-teren-background border border-teren-primary/30 rounded-xl shadow-sm space-y-3">
                      <div class="flex flex-wrap gap-3">
                        <div class="flex-1 min-w-[120px]">
                          <label class="text-xs text-teren-text-muted mb-1 block">Category</label>
                          <select 
                            bind:value={draft.category_id}
                            class="w-full px-3 py-2 bg-teren-surface border border-teren-border rounded-lg text-sm focus:ring-2 focus:ring-teren-primary/30 outline-none cursor-pointer appearance-none"
                          >
                            {#each categories as cat (cat.id)}
                              <option value={cat.id}>{emojiMap[cat.slug] || '📦'} {cat.slug}</option>
                            {/each}
                          </select>
                        </div>
                        <div class="flex-1 min-w-[120px]">
                          <label class="text-xs text-teren-text-muted mb-1 block">Amount</label>
                          <input type="number" step="0.01" bind:value={draft.amount} class="w-full px-3 py-2 bg-teren-surface border border-teren-border rounded-lg focus:ring-2 focus:ring-teren-primary/30 outline-none" onkeydown={e => e.key === 'Enter' && saveEdit(exp.id)} autofocus />
                        </div>
                        <div class="w-full sm:w-32">
                          <label class="text-xs text-teren-text-muted mb-1 block">Date</label>
                          <input type="date" bind:value={draft.date} class="w-full px-2 py-2 bg-teren-surface border border-teren-border rounded-lg text-sm focus:ring-2 focus:ring-teren-primary/30 outline-none" />
                        </div>
                      </div>
                      <textarea bind:value={draft.notes} placeholder="Add a note..." class="w-full px-3 py-2 bg-teren-surface border border-teren-border rounded-lg text-sm resize-none focus:ring-2 focus:ring-teren-primary/30 outline-none" rows="2" />
                      <div class="flex justify-end gap-2 pt-1">
                        <button onclick={() => editingId = null} class="px-4 py-2 text-sm text-teren-text-muted hover:text-teren-text-main hover:bg-gray-100 rounded-lg transition">Cancel</button>
                        <button onclick={() => saveEdit(exp.id)} class="px-4 py-2 text-sm bg-teren-primary hover:bg-teren-primary-hover text-white font-medium rounded-lg transition active:scale-95">Save</button>
                      </div>
                    </div>
                  {:else}
                    <div class="group p-4 bg-teren-surface border border-teren-border rounded-xl hover:border-teren-primary/20 transition cursor-pointer flex justify-between items-start" onclick={() => startEdit(exp)}>
                      <div class="flex gap-3">
                        <span class="text-xl select-none">{emojiMap[slug] || '📦'}</span>
                        <div>
                          <div class="font-bold text-teren-text-main">{exp.amount.toFixed(2)} {exp.currency}</div>
                          <div class="text-sm text-teren-text-muted">{new Date(exp.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}</div>
                          {#if exp.notes}
                            <div class="text-sm text-teren-text-muted mt-1 line-clamp-1 opacity-70 italic">"{exp.notes}"</div>
                          {/if}
                        </div>
                      </div>
                      <button onclick={() => deleteExpense(exp.id)} class="opacity-0 group-hover:opacity-100 text-red-400 hover:text-red-600 p-1.5 rounded-lg hover:bg-red-50 transition active:scale-95" aria-label="Delete">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                      </button>
                    </div>
                  {/if}
                {/each}
              </div>
            </section>
          {/each}
        {/if}
      </main>
    </div>
  </div>
{/if}