<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { apiFetch } from '$lib/api';
  import { fly } from 'svelte/transition';
  
  // Types
  import type { Place } from '$lib/types/Place';
  import type { CategorySummary } from '$lib/types/Summary';
  import type { Expense, Expense_Category } from '$lib/index';
  import type { Trip } from '$lib/types/Trip';
  
  // Components
  import ExpenseQuickAdd from '$lib/components/ExpenseQuickAdd.svelte';
  import ExpenseDrawer from '$lib/components/ExpenseDrawer.svelte';
  import ExpenseSummaryPills from '$lib/components/ExpenseSummaryPills.svelte';

  let tripId = $state('');
  let placeId = $state('');
  
  // State
  let place = $state<Place | null>(null);
  let expenses = $state<Expense[]>([]);
  let categorySummary = $state<CategorySummary[]>([]);
  let categories = $state<Expense_Category[]>([]);
  let tripCurrency = $state('EUR');
  
  let isDrawerOpen = $state(false);
  let isLoading = $state(true);

  // Extract IDs
  $effect(() => {
    if ($page.url.pathname) {
      const parts = $page.url.pathname.split('/');
      tripId = parts[2] || '';
      placeId = parts[4] || '';
      if (tripId && placeId) loadAllData();
    }
  });

  async function loadAllData() {
    if (!tripId || !placeId) return;
    isLoading = true;
    try {
      // Fetch place details and its expenses
      const [placeData, expensesData, summaryData, catsData, tripData] = await Promise.all([
        apiFetch<Place>(`/trips/${tripId}/places/${placeId}`),
        apiFetch<Expense[]>(`/trips/${tripId}/places/${placeId}/expenses`),
        apiFetch<CategorySummary[]>(`/trips/${tripId}/places/${placeId}/expenses/summary`),
        apiFetch<Expense_Category[]>(`/trips/${tripId}/expenses/categories`),
        apiFetch<Trip>(`/trips/${tripId}`)
      ]);

      place = placeData;
      expenses = expensesData;
      categorySummary = summaryData;
      categories = catsData;
      tripCurrency = tripData.base_currency || 'EUR';
    } catch (e) {
      console.error("Failed to load place data", e);
      goto(`/trips/${tripId}`);
    } finally {
      isLoading = false;
    }
  }

  function formatDate(dateStr?: string) {
    if (!dateStr) return 'No date';
    return new Date(dateStr).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  }

  function calculateDuration(start?: string, end?: string) {
    if (!start || !end) return '';
    const diffDays = Math.ceil((new Date(end).getTime() - new Date(start).getTime()) / (1000 * 60 * 60 * 24));
    return diffDays > 0 ? `${diffDays} days` : '1 day';
  }

  function calculateTotal() {
    return expenses.reduce((sum, exp) => sum + (exp.amount || 0), 0);
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
          <h1 class="text-xl font-bold text-teren-text-main leading-tight flex items-center gap-2">
            <svg class="w-5 h-5 text-teren-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            {place?.name || 'Loading...'}
          </h1>
          <p class="text-xs text-teren-text-muted font-medium">
            {formatDate(place?.start_date)} — {formatDate(place?.end_date)} • {calculateDuration(place?.start_date, place?.end_date)}
          </p>
        </div>
      </div>
    </div>
  </header>

  <main class="max-w-3xl mx-auto px-4 py-8 space-y-8">
    
    {#if isLoading}
      <div class="animate-pulse space-y-6">
        <div class="h-48 bg-teren-surface rounded-xl border border-teren-border"></div>
      </div>
    {:else if place}
      
      <!-- ========================================== -->
      <!-- 1. CONTEXTUAL EXPENSE CARD (The "Signature") -->
      <!-- ========================================== -->
      <section in:fly={{ y: 20, duration: 400, delay: 50 }} class="bg-teren-surface rounded-xl border border-teren-border shadow-sm border-l-4 border-l-teren-primary relative overflow-hidden">
        
        <!-- Header de la Tarjeta (Con degradado sutil) -->
        <div class="p-6 border-b border-teren-border flex justify-between items-center bg-gradient-to-r from-teren-surface to-teren-primary-subtle/20">
          <div>
            <h2 class="text-lg font-bold text-teren-text-main flex items-center gap-2">
              Expenses
              <span class="text-xs font-bold text-teren-primary bg-white px-2 py-0.5 rounded-full border border-teren-primary/20 shadow-sm">
                Local
              </span>
            </h2>
          </div>
          <span class="text-2xl font-bold text-teren-primary">
            {calculateTotal().toFixed(2)} {tripCurrency}
          </span>
        </div>

        <!-- Category Pills -->
        <div class="px-6 pt-4">
          {#if categorySummary.length > 0}
            <ExpenseSummaryPills {categories} summary={categorySummary} currency={tripCurrency} />
          {:else}
            <div class="pb-2 text-center">
              <p class="text-sm text-teren-text-muted italic">No local expenses yet.</p>
            </div>
          {/if}
        </div>

        <!-- Quick Add (Embedded for Speed) -->
        <div class="p-6 pb-2">
          <ExpenseQuickAdd 
            tripId={tripId} 
            placeId={placeId} 
            {categories} 
            onSuccess={loadAllData} 
          />
          <p class="text-xs text-teren-text-muted mt-3 text-center opacity-60 flex justify-center items-center gap-1">
            <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"/></svg>
            Expense automatically linked to {place.name}
          </p>
        </div>

        <!-- Link to Drawer (History) -->
        <button 
          onclick={() => isDrawerOpen = true} 
          class="w-full py-4 text-sm font-medium text-teren-text-muted hover:bg-teren-primary-subtle hover:text-teren-primary transition border-t border-teren-border mt-4"
        >
          View full history →
        </button>
      </section>

      <!-- ========================================== -->
      <!-- 2. PLACE NOTES (If any) -->
      <!-- ========================================== -->
      {#if place.notes}
        <section in:fly={{ y: 20, duration: 400, delay: 100 }} class="bg-teren-background p-6 rounded-xl border border-teren-border">
          <h3 class="text-sm font-semibold text-teren-text-muted uppercase tracking-wider mb-3">Notes</h3>
          <p class="text-teren-text-main leading-relaxed whitespace-pre-wrap">{place.notes}</p>
        </section>
      {/if}

      <!-- ========================================== -->
      <!-- 3. ACTIVITIES (Placeholder v2) -->
      <!-- ========================================== -->
      <section in:fly={{ y: 20, duration: 400, delay: 150 }} class="bg-teren-surface p-8 rounded-xl border border-teren-border border-dashed text-center">
        <div class="w-12 h-12 mx-auto mb-3 rounded-full bg-teren-background border border-teren-border flex items-center justify-center">
          <svg class="w-6 h-6 text-teren-text-muted opacity-60" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
        </div>
        <h3 class="text-teren-text-main font-semibold mb-1">Activities</h3>
        <p class="text-teren-text-muted text-sm">Coming soon in v2</p>
      </section>

    {/if}
  </main>

  <!-- Drawer (Only for viewing/editing history) -->
  <ExpenseDrawer 
    tripId={tripId}
    placeId={placeId}
    {categories}
    isOpen={isDrawerOpen}
    onClose={() => isDrawerOpen = false}
    onRefreshSummary={loadAllData}
  />
</div>