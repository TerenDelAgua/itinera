<script lang="ts">
  import { t } from '$lib/i18n/store';
  import { apiFetch } from '$lib/api';
  import type { Category, Expense } from '$lib/types';

  let { tripId, categories, onSuccess }: { tripId: string; categories: Category[]; onSuccess: (exp: Expense) => void; } = $props();

  let amount = $state('');
  let categoryId = $state(categories[0]?.id || '');
  let isSubmitting = $state(false);

  // Mapeo rápido de slugs a emojis para reconocimiento instantáneo
  const emojiMap: Record<string, string> = {
    accommodation: '🏨', transport: '🚆', food: '🍔',
    leisure: '🎟️', shopping: '🛍️', others: '📦'
  };

  async function handleSubmit() {
    if (!amount || parseFloat(amount) <= 0) return;
    isSubmitting = true;
    try {
      const exp = await apiFetch<Expense>(`/trips/${tripId}/expenses`, {
        method: 'POST',
        body: JSON.stringify({
          amount: parseFloat(amount),
          category_id: categoryId,
          date: new Date().toISOString(),
          currency: 'EUR' // Se inyectará dinámicamente desde settings del viaje en v1.1
        })
      });
      onSuccess(exp);
      amount = ''; // Reset rápido para siguiente gasto
    } catch {
      // TEREN: Fallback silencioso o toast sutil. No bloquear flujo.
    } finally {
      isSubmitting = false;
    }
  }
</script>

<div class="flex items-center gap-3 bg-teren-surface p-3 rounded-xl border border-teren-border shadow-sm mb-6">
  
  <!-- Selector de Categoría (Icono + Dropdown nativo optimizado) -->
  <div class="relative">
    <select 
      bind:value={categoryId}
      class="appearance-none bg-gray-50 border border-teren-border rounded-lg pl-3 pr-8 py-2.5 text-2xl focus:outline-none focus:ring-2 focus:ring-teren-primary/30 focus:border-teren-primary transition-all cursor-pointer"
    >
      {#each categories as cat}
        <option value={cat.id}>{emojiMap[cat.slug] || '📦'}</option>
      {/each}
    </select>
    <!-- Flecha custom sutil -->
    <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-teren-text-muted">
      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" /></svg>
    </div>
  </div>

  <!-- Input Cantidad -->
  <div class="flex-1 relative">
    <span class="absolute left-3 top-1/2 -translate-y-1/2 text-teren-text-muted font-bold text-lg">€</span>
    <input 
      type="number" 
      inputmode="decimal"
      step="0.01"
      bind:value={amount}
      onkeydown={(e) => e.key === 'Enter' && handleSubmit()}
      placeholder="0.00"
      class="w-full pl-8 pr-4 py-2.5 bg-white border border-teren-border rounded-lg text-teren-text-main font-bold text-lg focus:outline-none focus:ring-2 focus:ring-teren-primary/30 focus:border-teren-primary transition-all placeholder:text-teren-text-muted/50"
    />
  </div>

  <!-- Botón Acción -->
  <button 
    onclick={handleSubmit}
    disabled={isSubmitting || !amount}
    class="bg-teren-primary hover:bg-teren-primary-hover text-white rounded-lg w-11 h-11 flex items-center justify-center shadow-sm active:scale-95 transition-all duration-200 disabled:opacity-50 disabled:active:scale-100">
    {#if isSubmitting}
      <div class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
    {:else}
      <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 4v16m8-8H4" />
      </svg>
    {/if}
  </button>
</div>