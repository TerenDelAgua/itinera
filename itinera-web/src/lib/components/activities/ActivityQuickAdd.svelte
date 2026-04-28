<script lang="ts">
  import { activityApi } from '$lib/api/activity';
  import { slide } from 'svelte/transition';

  let { tripId, placeId, tripStartDate, tripEndDate, onActivityAdded }: {
    tripId: string;
    placeId?: string;
    tripStartDate?: string;
    tripEndDate?: string;
    onActivityAdded: () => void;
  } = $props();

  // Estados reactivos
  let title = $state('');
  let date = $state(getDefaultDate());
  let time = $state('');
  let notes = $state('');
  let isLoading = $state(false);
  let errorMessage = $state('');

  function getDefaultDate() {
    const today = new Date().toISOString().split('T')[0];
    if (tripStartDate && tripEndDate) {
      if (today >= tripStartDate && today <= tripEndDate) return today;
      return tripStartDate;
    }
    return today;
  }

  async function handleSubmit() {
    errorMessage = '';
    if (!title.trim()) return;

    // Validación de rango de fechas (TEREN requirement)
    if (tripStartDate && tripEndDate) {
        if (date < tripStartDate || date > tripEndDate) {
            errorMessage = "Activity date must be within trip dates";
            return;
        }
    }

    isLoading = true;
    try {
      await activityApi.create(tripId, {
        title: title.trim(),
        date,
        time: time || undefined,
        notes: notes || undefined,
        place_id: placeId
      });
      
      // Reset form on success
      title = '';
      time = '';
      notes = '';
      errorMessage = '';
      onActivityAdded();
    } catch (e) {
      console.error(e);
      errorMessage = "Failed to add activity. Please try again.";
    } finally {
      isLoading = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && title.trim()) {
      handleSubmit();
    }
  }
</script>

<div class="bg-white p-5 rounded-2xl border border-teren-border shadow-sm space-y-4">
  <div class="flex justify-between items-center">
    <h3 class="text-xs font-bold text-teren-text-muted uppercase tracking-widest">New Activity</h3>
  </div>

  <div class="space-y-3" onkeydown={handleKeydown}>
    <!-- Primera fila: Título y Fecha -->
    <div class="flex flex-col sm:flex-row gap-3">
      <div class="flex-1">
        <input 
          type="text" 
          bind:value={title} 
          placeholder="Activity name (e.g., Museum visit)"
          class="w-full px-4 py-2.5 bg-teren-background border border-teren-border rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-teren-primary/20 transition-all placeholder:text-teren-text-muted/50"
        />
      </div>
      <div class="w-full sm:w-40">
        <input 
          type="date" 
          bind:value={date}
          min={tripStartDate}
          max={tripEndDate}
          class="w-full px-3 py-2.5 bg-teren-background border border-teren-border rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-teren-primary/20 transition-all"
        />
      </div>
    </div>

    <!-- Segunda fila: Hora y Notas -->
    <div class="flex flex-col sm:flex-row gap-3">
      <div class="w-full sm:w-32">
        <input 
          type="time" 
          bind:value={time}
          class="w-full px-3 py-2.5 bg-teren-background border border-teren-border rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-teren-primary/20 transition-all"
        />
      </div>
      <div class="flex-1">
        <input 
          type="text" 
          bind:value={notes} 
          placeholder="Optional notes or link..."
          class="w-full px-4 py-2.5 bg-teren-background border border-teren-border rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-teren-primary/20 transition-all placeholder:text-teren-text-muted/50"
        />
      </div>
    </div>

    <!-- Error Message -->
    {#if errorMessage}
      <p class="text-xs text-red-500 font-medium px-1" transition:slide>
        {errorMessage}
      </p>
    {/if}

    <button 
      onclick={handleSubmit}
      disabled={isLoading || !title.trim()}
      class="w-full bg-teren-primary hover:bg-teren-primary-hover text-white font-bold py-3 rounded-xl text-sm transition-all active:scale-[0.98] disabled:opacity-50 disabled:grayscale disabled:cursor-not-allowed shadow-lg shadow-teren-primary/20"
    >
      {#if isLoading}
        <span class="flex items-center justify-center gap-2">
            <svg class="animate-spin h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Saving...
        </span>
      {:else}
        Add to Agenda
      {/if}
    </button>
  </div>
</div>
