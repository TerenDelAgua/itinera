<script lang="ts">
  import { activityApi } from '$lib/api/activity';
  import { t } from '$lib/i18n/store';
  import { slide } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';

  let { 
    tripId, 
    placeId = undefined, 
    onSuccess 
  }: { 
    tripId: string; 
    placeId?: string; 
    onSuccess: () => void;
  } = $props();

  let newTitle = $state('');
  let newDate = $state(new Date().toISOString().split('T')[0]);
  let newTime = $state('');
  let newNotes = $state('');
  let isLoading = $state(false);

  async function handleCreate() {
    if (!newTitle.trim()) return;
    isLoading = true;
    try {
      await activityApi.create(tripId, {
        title: newTitle,
        date: newDate,
        time: newTime || undefined,
        notes: newNotes || undefined,
        place_id: placeId
      });
      newTitle = ''; newTime = ''; newNotes = '';
      onSuccess();
    } finally {
      isLoading = false;
    }
  }
</script>

<div class="bg-teren-background p-4 rounded-xl border-2 border-teren-primary/30 space-y-3" transition:slide={{ duration: 250, easing: cubicOut }}>
  <div class="flex flex-wrap gap-3">
    <input 
      type="text" 
      bind:value={newTitle} 
      placeholder={$t('itinerary.title_placeholder')} 
      class="w-full sm:flex-1 min-w-0 px-3 py-2 bg-white border border-teren-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-teren-primary/30" 
      autofocus
      onkeydown={e => e.key === 'Enter' && handleCreate()}
    />
    <input 
      type="date" 
      bind:value={newDate} 
      class="w-[calc(50%-0.375rem)] sm:w-32 min-w-0 px-2 py-2 bg-white border border-teren-border rounded-lg text-sm outline-none focus:ring-2 focus:ring-teren-primary/30 flex-shrink-0 sm:flex-shrink" 
    />
    <input 
      type="time" 
      bind:value={newTime} 
      class="w-[calc(50%-0.375rem)] sm:w-32 min-w-0 px-3 py-2 bg-white border border-teren-border rounded-lg text-sm outline-none focus:ring-2 focus:ring-teren-primary/30 flex-shrink-0 sm:flex-shrink" 
    />
    
    <!-- Separador invisible para forzar el salto de línea solo en escritorio -->
    <div class="hidden sm:block w-full h-0 m-0 p-0"></div>

    <div class="w-full flex gap-3 items-stretch">
      <input 
        type="text" 
        bind:value={newNotes} 
        placeholder={$t('itinerary.notes_placeholder')} 
        class="flex-1 min-w-0 px-3 py-2 bg-white border border-teren-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-teren-primary/30" 
        onkeydown={e => e.key === 'Enter' && handleCreate()}
      />
      <button 
        onclick={handleCreate} 
        disabled={isLoading || !newTitle} 
        class="flex-shrink-0 bg-teren-primary hover:bg-teren-primary-hover text-white px-4 rounded-lg font-bold shadow-sm active:scale-95 transition-all disabled:opacity-50 disabled:active:scale-100 flex items-center justify-center"
      >
        {#if isLoading}
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
