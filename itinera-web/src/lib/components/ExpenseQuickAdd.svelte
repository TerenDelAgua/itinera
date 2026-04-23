<script lang="ts">
  import { apiFetch } from '$lib/api';
  import type { Expense, Category } from '$lib/types';
  import { t } from '$lib/i18n/store';
  import { getCurrencySymbol } from '$lib/utils';

  let { tripId, categories, onSuccess, placeId, currency = 'EUR' }: { tripId: string; categories: Category[]; onSuccess: (exp: Expense) => void; placeId?: string; currency?: string; } = $props();

  let amount = $state('');
  let categoryId = $state(categories[0]?.id || '');
  let notes = $state('');
  let isSubmitting = $state(false);

  const emojiMap: Record<string, string> = {
    accommodation: '🏨', transport: '🚆', food: '🍔',
    leisure: '🎟️', shopping: '🛍️', others: '📦'
  };

  async function handleSubmit() {
    if (!amount || parseFloat(amount) <= 0 || !categoryId) return;
    isSubmitting = true;
    try {
      const exp = await apiFetch<Expense>(`/trips/${tripId}/expenses`, {
        method: 'POST',
        body: JSON.stringify({
          amount: parseFloat(amount),
          category_id: categoryId,
          date: new Date().toISOString(),
          currency: currency,
          notes: notes.trim(),
          place_id: placeId
        })
      });
      onSuccess(exp);
      amount = ''; notes = '';
    } finally { 
      isSubmitting = false; 
    }
  }
</script>

<div class="relative group">
  <!-- Glowing effect en el hover (muy sutil) -->
  <div class="absolute -inset-0.5 bg-gradient-to-r from-teren-primary/0 via-teren-primary/10 to-teren-primary/0 rounded-2xl blur opacity-0 group-hover:opacity-100 transition duration-500"></div>
  
  <div class="relative bg-teren-surface rounded-2xl border border-teren-border/80 shadow-sm focus-within:border-teren-primary/50 focus-within:ring-4 focus-within:ring-teren-primary/10 transition-all overflow-hidden p-1">
    
    <!-- Fila Superior: Categoría y Cantidad -->
    <div class="flex items-stretch h-14 bg-teren-surface">
      <!-- Selector de Categoría (Solo Emoji visualmente) -->
      <div class="relative flex items-center justify-center w-16 hover:bg-teren-background transition-colors cursor-pointer border-r border-teren-border/50">
        <select 
          bind:value={categoryId} 
          class="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
        >
          {#each categories as cat (cat.id)}
            <option value={cat.id}>{cat.name}</option>
          {/each}
        </select>
        <span class="text-2xl pointer-events-none">{emojiMap[categories.find(c => c.id === categoryId)?.slug || ''] || '📦'}</span>
      </div>

      <!-- Input de Cantidad -->
      <div class="flex-1 relative flex items-center ml-2">
        <span class="absolute left-3 text-teren-text-muted font-bold text-lg pointer-events-none select-none">{getCurrencySymbol(currency)}</span>
        <input 
          type="number" 
          inputmode="decimal" 
          step="0.01" 
          bind:value={amount} 
          placeholder="0.00"
          class="w-full h-11 pl-8 pr-4 bg-transparent border-none text-xl font-bold text-teren-text-main focus:ring-0 placeholder:text-teren-text-muted/30"
          onkeydown={e => e.key === 'Enter' && handleSubmit()} 
        />
      </div>
    </div>

    <div class="mx-4 h-px bg-teren-border/50"></div>

    <!-- Fila Inferior: Notas y Botón -->
    <div class="flex items-center h-14 bg-teren-background/50 pl-4 pr-2 py-2">
      <input 
        type="text" 
        bind:value={notes} 
        placeholder={$t('detail.notes_optional')}
        class="flex-1 bg-transparent border-none p-0 text-sm h-11 focus:ring-0 placeholder:text-teren-text-muted/60 text-teren-text-main"
        onkeydown={e => e.key === 'Enter' && handleSubmit()} 
      />
      <button 
        onclick={handleSubmit} 
        disabled={isSubmitting || !amount}
        class="ml-3 bg-teren-primary hover:bg-teren-primary-hover text-white h-11 px-6 rounded-xl font-bold shadow-sm active:scale-95 transition-all disabled:opacity-50 disabled:active:scale-100 flex items-center justify-center"
      >
        {#if isSubmitting}
          <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
        {:else}
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 4v16m8-8H4" />
          </svg>
        {/if}
      </button>
    </div>
  </div>
</div>