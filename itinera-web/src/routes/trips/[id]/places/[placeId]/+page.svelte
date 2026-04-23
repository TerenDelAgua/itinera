<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { apiFetch } from '$lib/api';
  import { t } from '$lib/i18n/store';
  import { tweened } from 'svelte/motion';
  import { cubicOut } from 'svelte/easing';
  
  // Types
  import type { Place } from '$lib/types/Place';
  import type { Expense } from '$lib/types/Expense';
  import type { CategorySummary } from '$lib/types/Summary';
  import type { ExpenseCategory } from '$lib/types/ExpenseCategory';
  import type { Trip } from '$lib/types/Trip';
  
  // Components
  import ExpenseQuickAdd from '$lib/components/ExpenseQuickAdd.svelte';
  import ExpenseDrawer from '$lib/components/ExpenseDrawer.svelte';
  import { getCurrencySymbol } from '$lib/utils';

  let tripId = '';
  let placeId = '';
  
  // State
  let place = $state<Place | null>(null);
  let expenses = $state<Expense[]>([]);
  let categorySummary = $state<CategorySummary[]>([]);
  let categories = $state<ExpenseCategory[]>([]);
  let tripCurrency = $state('EUR');
  
  let animatedTotal = tweened(0, { duration: 600, easing: cubicOut });
  
  let isDrawerOpen = $state(false);
  let isLoading = $state(true);

  // Extraer IDs de la URL
  $effect(() => {
    tripId = $page.params.id;
    placeId = $page.params.placeId;
    if (tripId && placeId) {
      loadAllData();
    }
  });

  async function loadAllData() {
    if (!tripId || !placeId) return;
    isLoading = true;
    try {
      const [placeData, expensesData, summaryData, catsData, tripData] = await Promise.all([
        apiFetch<Place>(`/trips/${tripId}/places/${placeId}`),
        apiFetch<Expense[]>(`/trips/${tripId}/places/${placeId}/expenses`),
        apiFetch<CategorySummary[]>(`/trips/${tripId}/places/${placeId}/expenses/summary`),
        apiFetch<ExpenseCategory[]>(`/trips/${tripId}/expenses/categories`),
        apiFetch<Trip>(`/trips/${tripId}`)
      ]);

      place = placeData;
      expenses = expensesData || [];
      categorySummary = summaryData || [];
      categories = catsData || [];
      tripCurrency = tripData?.base_currency || 'EUR';
      
      animatedTotal.set(calculateTotal());
    } catch (e) {
      console.error("Failed to load place data", e);
      goto(`/trips/${tripId}`);
    } finally {
      isLoading = false;
    }
  }

  // Utilidades
  const emojiMap: Record<string, string> = {
    accommodation: '🏨', transport: '🚆', food: '🍔',
    leisure: '🎟️', shopping: '🛍️', others: '📦'
  };

  function formatDate(dateStr?: string) {
    if (!dateStr) return 'No date';
    return new Date(dateStr).toLocaleDateString('en-US', { 
      month: 'short', 
      day: 'numeric', 
      year: 'numeric' 
    });
  }

  function calculateDuration(start?: string, end?: string) {
    if (!start || !end) return '';
    const startDate = new Date(start);
    const endDate = new Date(end);
    const diffTime = Math.abs(endDate.getTime() - startDate.getTime());
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    return diffDays > 0 ? `${diffDays} days` : '1 day';
  }

  function calculateTotal() {
    if (!expenses) return 0;
    return expenses.reduce((sum, exp) => sum + exp.amount, 0);
  }
</script>

<svelte:head>
  <title>{place ? `${place.name} | Itinera` : 'Loading...'}</title>
</svelte:head>

<div class="min-h-screen bg-teren-background pb-20">
  
  <!-- Header Fijo -->
  <header class="sticky top-0 z-40 bg-teren-background/90 backdrop-blur-md border-b border-teren-border">
    <div class="max-w-3xl mx-auto px-4 h-16 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <button 
          onclick={() => goto(`/trips/${tripId}`)} 
          class="p-2 -ml-2 text-teren-text-muted hover:text-teren-text-main hover:bg-gray-100 rounded-lg transition active:scale-95"
        >
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <div>
          <h1 class="text-xl font-bold text-teren-text-main leading-tight">{place?.name || 'Loading...'}</h1>
          {#if place}
            <p class="text-xs text-teren-text-muted font-medium">
              {formatDate(place.start_date)} — {formatDate(place.end_date)}
              {#if place.start_date && place.end_date}
                • {calculateDuration(place.start_date, place.end_date)}
              {/if}
            </p>
          {/if}
        </div>
      </div>
    </div>
  </header>

  <main class="max-w-3xl mx-auto px-4 py-8 space-y-10">
    
    {#if isLoading}
      <!-- Skeleton Loading -->
      <div class="animate-pulse space-y-6">
        <div class="h-40 bg-teren-surface rounded-xl border border-teren-border"></div>
        <div class="h-64 bg-teren-surface rounded-xl border border-teren-border"></div>
      </div>
    {:else if place}
      
      <!-- ========================================== -->
      <!-- 1. PLACE EXPENSES CARD -->
      <!-- ========================================== -->
      <section class="bg-teren-surface p-6 rounded-xl border border-teren-border shadow-sm">
        <div class="flex justify-between items-center mb-6">
          <h2 class="text-lg font-semibold text-teren-text-main tracking-tight">{$t('detail.expenses')}</h2>
          <span class="text-2xl font-bold text-teren-primary tabular-nums">
            {$animatedTotal.toFixed(2)} {getCurrencySymbol(tripCurrency)}
          </span>
        </div>

        <!-- Category Pills -->
        {#if categorySummary && categorySummary.length > 0}
          <div class="flex flex-wrap gap-2 mb-6">
            {#each categorySummary as cat (cat.id)}
              <span class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-teren-primary-subtle text-teren-primary-hover rounded-full text-sm font-bold border border-teren-primary/20">
                {emojiMap[cat.category_id] || '📦'} {cat.total.toFixed(2)} {getCurrencySymbol(tripCurrency)}
              </span>
            {/each}
          </div>
        {:else}
          <p class="text-teren-text-muted text-sm mb-6 italic">{$t('detail.expenses_empty')}</p>
        {/if}

        <!-- Quick Add (Place-scoped) -->
        <ExpenseQuickAdd 
          tripId={tripId} 
          placeId={placeId}
          {categories} 
          currency={tripCurrency}
          onSuccess={loadAllData} 
        />

        <!-- Link al Drawer -->
        <button 
          onclick={() => isDrawerOpen = true}
          class="w-full flex items-center justify-center gap-2 py-3 px-4 bg-teren-background hover:bg-teren-surface text-teren-primary font-bold text-sm rounded-xl border border-teren-primary/10 transition-all active:scale-[0.98] mt-5"
        >
          {$t('place.view_place_expenses')} →
        </button>
      </section>

      <!-- ========================================== -->
      <!-- 2. PLACE NOTES -->
      <!-- ========================================== -->
      {#if place.notes}
        <section class="bg-teren-background p-6 rounded-xl border border-teren-border">
          <h3 class="text-sm font-semibold text-teren-text-muted uppercase tracking-wider mb-3">Notes</h3>
          <p class="text-teren-text-main leading-relaxed whitespace-pre-wrap">{place.notes}</p>
        </section>
      {/if}

      <!-- ========================================== -->
      <!-- 3. ACTIVITIES (Placeholder v2) -->
      <!-- ========================================== -->
      <section class="bg-teren-surface p-6 rounded-xl border border-teren-border border-dashed">
        <div class="text-center py-8">
          <div class="w-12 h-12 mx-auto mb-3 rounded-full bg-teren-background border border-teren-border flex items-center justify-center">
            <svg class="w-6 h-6 text-teren-text-muted opacity-60" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </div>
          <h3 class="text-teren-text-main font-semibold mb-1">Activities</h3>
          <p class="text-teren-text-muted text-sm">Coming soon in v2</p>
        </div>
      </section>

    {/if}
  </main>

  <!-- Drawer para Gastos de este Place -->
  <ExpenseDrawer 
    tripId={tripId}
    placeId={placeId}
    {categories}
    isOpen={isDrawerOpen}
    onClose={() => isDrawerOpen = false}
    onRefreshSummary={loadAllData}
  />
</div>