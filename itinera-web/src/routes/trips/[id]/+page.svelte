<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { apiFetch } from '$lib/api';
  import { t, locale } from "$lib/i18n/store";
  import { slide } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import { tweened } from 'svelte/motion';

  // Types
  import type { Place } from '$lib/types/Place';
  import type { TripExpenseSummary } from '$lib/types/Summary';
  import type { Expense_Category, Trip } from '$lib/index';
  
  // Components
  import ExpenseQuickAdd from '$lib/components/ExpenseQuickAdd.svelte';
  import ExpenseDrawer from '$lib/components/ExpenseDrawer.svelte';
  import ExpenseSummaryPills from "$lib/components/ExpenseSummaryPills.svelte";
  import ConfirmModal from '$lib/components/ConfirmModal.svelte';
  import { getCurrencySymbol } from '$lib/utils';

  let tripId = $state('');
  
  // State
  let tripName = $state('');
  let tripDescription = $state('');
  let tripDates = $state({ start: '', end: '' });
  let currency = $state('EUR'); // Default
  
  let places = $state<Place[]>([]);
  let summary = $state<TripExpenseSummary | null>(null);
  let categories = $state<Expense_Category[]>([]);
  let animatedGrandTotal = tweened(0, { duration: 600, easing: cubicOut });
  
  let isDrawerOpen = $state(false);
  let isLoading = $state(true);

  let isCreatingPlace = $state(false);
  let newPlaceDraft = $state({ name: '', start_date: '', end_date: '' });
  let deletePlaceConfirmId = $state<string | null>(null);

  $effect(() => {
    if ($page.url.pathname) {
      tripId = $page.url.pathname.split('/').pop() || '';
      loadAllData();
    }
  });

  async function loadAllData() {
    if (!tripId) return;
    isLoading = true;
    try {
      const [tripData, placesData, summaryData, catsData] = await Promise.all([
        apiFetch<Trip>(`/trips/${tripId}`),
        apiFetch<Place[]>(`/trips/${tripId}/places`),
        apiFetch<TripExpenseSummary>(`/trips/${tripId}/expenses/summary`),
        apiFetch<Expense_Category[]>(`/trips/${tripId}/expenses/categories`)
      ]);

      tripName = tripData.name;
      tripDescription = tripData.description || '';
      tripDates = { start: tripData.start_date, end: tripData.end_date };
      currency = tripData.base_currency || 'EUR';
      
      places = placesData || [];
      summary = summaryData;
      categories = catsData || [];
      
      if (summary) {
        animatedGrandTotal.set(summary.grand_total || 0);
      }
    } catch (e) {
      console.error("Failed to load trip data", e);
    } finally {
      isLoading = false;
    }
  }

  async function saveTripInfo() {
    if (!tripName) return;
    try {
      const payload = {
        name: tripName,
        description: tripDescription,
        start_date: new Date(tripDates.start).toISOString(),
        end_date: new Date(tripDates.end).toISOString(),
        base_currency: currency
      };
      await apiFetch(`/trips/${tripId}`, {
        method: 'PUT',
        body: JSON.stringify(payload)
      });
      // reload to sync with summary if currency changed
      loadAllData();
    } catch (e) {
      console.error("Failed to update trip", e);
    }
  }

  async function createPlace() {
    if (!newPlaceDraft.name) return;
    try {
      const payload: any = { name: newPlaceDraft.name };
      if (newPlaceDraft.start_date) payload.start_date = new Date(newPlaceDraft.start_date).toISOString();
      if (newPlaceDraft.end_date) payload.end_date = new Date(newPlaceDraft.end_date).toISOString();
      
      await apiFetch(`/trips/${tripId}/places`, {
        method: 'POST',
        body: JSON.stringify(payload)
      });
      
      isCreatingPlace = false;
      newPlaceDraft = { name: '', start_date: '', end_date: '' };
      loadAllData();
    } catch (e) {
      console.error("Failed to create place", e);
    }
  }

  function requestDeletePlace(id: string) {
    deletePlaceConfirmId = id;
  }

  async function confirmDeletePlace() {
    if (!deletePlaceConfirmId) return;
    const id = deletePlaceConfirmId;
    deletePlaceConfirmId = null;
    
    try {
      await apiFetch(`/trips/${tripId}/places/${id}`, { method: 'DELETE' });
      loadAllData();
    } catch (e) {
      console.error("Failed to delete place", e);
    }
  }

  function cancelDeletePlace() {
    deletePlaceConfirmId = null;
  }

  const emojiMap: Record<string, string> = {
    accommodation: '🏨', transport: '🚆', food: '🍔',
    leisure: '🎟️', shopping: '🛍️', others: '📦'
  };

  function formatDate(dateStr?: string) {
    if (!dateStr) return '';
    return new Date(dateStr).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  }
</script>

<!-- Layout Principal -->
<div class="min-h-screen bg-teren-background pb-20">
  
  <!-- Header Fijo (Back + Título) -->
  <header class="sticky top-0 z-40 bg-teren-background/90 backdrop-blur-md border-b border-teren-border py-2">
    <div class="max-w-3xl mx-auto px-4 flex items-start justify-between">
      <div class="flex items-start gap-3 w-full">
        <button 
          onclick={() => goto('/')} 
          class="p-2 -ml-2 mt-0.5 text-teren-text-muted hover:text-teren-text-main hover:bg-gray-100 rounded-lg transition active:scale-95 flex-shrink-0"
        >
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
        </button>
        <div class="flex-1 min-w-0 flex flex-col gap-1">
          <div class="flex justify-between items-center">
            <input 
              type="text" 
              bind:value={tripName}   
              onblur={saveTripInfo}
              onkeydown={(e) => e.key === 'Enter' && e.currentTarget.blur()}
              placeholder="Trip Name"
              class="bg-transparent border-none p-0 focus:ring-0 text-xl font-bold text-teren-text-main leading-tight outline-none w-full truncate" 
            />
            <select 
              bind:value={currency} 
              onchange={saveTripInfo} 
              class="ml-3 bg-transparent text-sm font-bold text-teren-text-muted hover:text-teren-primary transition-colors border-none focus:ring-0 cursor-pointer outline-none flex-shrink-0"
            >
              <option value="EUR">€</option>
              <option value="USD">$</option>
              <option value="GBP">£</option>
              <option value="JPY">¥</option>
            </select>
          </div>
          <div class="flex items-center gap-1.5 text-xs text-teren-text-muted font-medium">
            <svg class="w-3.5 h-3.5 opacity-70 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/></svg>
            <input type="date" bind:value={tripDates.start} onchange={saveTripInfo} class="bg-transparent border-none p-0 focus:ring-0 outline-none cursor-pointer w-auto" />
            <span class="opacity-70">—</span>
            <input type="date" bind:value={tripDates.end} onchange={saveTripInfo} class="bg-transparent border-none p-0 focus:ring-0 outline-none cursor-pointer w-auto" />
          </div>
          <input 
            type="text" 
            bind:value={tripDescription} 
            onblur={saveTripInfo}
            onkeydown={(e) => e.key === 'Enter' && e.currentTarget.blur()}
            placeholder="Add a description... (inline edit)"
            class="bg-transparent border-none p-0 focus:ring-0 text-sm text-teren-text-muted outline-none w-full truncate mt-0.5" 
          />
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
    {:else}
      
      <!-- ========================================== -->
      <!-- 1. GLOBAL EXPENSES CARD -->
      <!-- ========================================== -->
      <section class="bg-teren-surface p-6 rounded-xl border border-teren-border shadow-sm">
        <div class="flex justify-between items-center mb-6">
          <h2 class="text-lg font-semibold text-teren-text-main tracking-tight">{$t('detail.expenses')}</h2>
          <!-- Muestra solo el total GLOBAL animado -->
          <span class="text-2xl font-bold text-teren-primary tabular-nums">
            {$animatedGrandTotal.toFixed(2)} {getCurrencySymbol(currency)}
          </span>
        </div>

        <!-- Category Pills (Globales) -->
         <ExpenseSummaryPills {categories} summary={summary?.by_category || []} currency={currency} />

        <!-- {#if summary?.by_category && summary.by_category.length > 0}
          <div class="flex flex-wrap gap-2 mb-6">
            {#each summary.by_category as cat (cat.category_id)}
              <span class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-teren-primary-subtle text-teren-primary-hover rounded-full text-sm font-bold border border-teren-primary/20">
                {emojiMap[cat.category_id] || '📦'} {cat.total?.toFixed(2) || '0.00'}
              </span>
            {/each}
          </div>
        {:else}
          <p class="text-teren-text-muted text-sm mb-6 italic">No global expenses added yet.</p>
        {/if} -->

        <!-- Quick Add (Global Context) -->
        <ExpenseQuickAdd 
          tripId={tripId} 
          {categories} 
          {currency}
          onSuccess={loadAllData} 
        />

        <!-- Link al Drawer (Ver todo) -->
        <button 
          onclick={() => isDrawerOpen = true} 
          class="mt-5 text-sm text-teren-text-muted hover:text-teren-primary transition flex items-center gap-1 group"
        >
          {$t('detail.view_all_expenses')} 
          <svg class="w-4 h-4 group-hover:translate-x-1 transition-transform" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 8l4 4m0 0l-4 4m4-4H3"/>
          </svg>
        </button>
      </section>

      <!-- ========================================== -->
      <!-- 2. DESTINATIONS (PLACES) -->
      <!-- ========================================== -->
      <section>
        <div class="flex justify-between items-center mb-4">
          <h2 class="text-lg font-semibold text-teren-text-main tracking-tight">{$t('detail.destinations')}</h2>
          <button 
            onclick={() => isCreatingPlace = true}
            class="text-sm font-medium text-teren-primary hover:text-teren-primary-hover transition px-3 py-1.5 rounded-lg hover:bg-teren-primary-subtle active:scale-95"
          >
            + {$t('detail.add_destination')}
          </button>
        </div>

        {#if isCreatingPlace}
          <div class="mb-4 p-4 bg-teren-background border-2 border-teren-primary/30 rounded-xl space-y-3" transition:slide={{ duration: 250, easing: cubicOut }}>
            <input 
              type="text" 
              bind:value={newPlaceDraft.name} 
              placeholder="Destination name (e.g. Kyoto)" 
              class="w-full px-3 py-2 text-sm font-bold bg-white border border-teren-border rounded-lg focus:ring-2 focus:ring-teren-primary/30 outline-none" 
              autofocus 
            />
            <div class="flex gap-3">
              <div class="flex-1">
                <label class="block text-xs text-teren-text-muted mb-1 ml-1 font-medium">Start Date (Optional)</label>
                <input 
                  type="date" 
                  bind:value={newPlaceDraft.start_date} 
                  class="w-full px-3 py-2 text-sm bg-white border border-teren-border rounded-lg focus:ring-2 focus:ring-teren-primary/30 outline-none" 
                />
              </div>
              <div class="flex-1">
                <label class="block text-xs text-teren-text-muted mb-1 ml-1 font-medium">End Date (Optional)</label>
                <input 
                  type="date" 
                  bind:value={newPlaceDraft.end_date} 
                  class="w-full px-3 py-2 text-sm bg-white border border-teren-border rounded-lg focus:ring-2 focus:ring-teren-primary/30 outline-none" 
                />
              </div>
            </div>
            <div class="flex justify-end gap-2 pt-2">
              <button 
                onclick={() => isCreatingPlace = false} 
                class="px-4 py-2 text-sm text-teren-text-muted hover:text-teren-text-main hover:bg-gray-100 rounded-lg transition"
              >
                Cancel
              </button>
              <button 
                onclick={createPlace} 
                disabled={!newPlaceDraft.name}
                class="px-4 py-2 text-sm bg-teren-primary hover:bg-teren-primary-hover text-white font-medium rounded-lg transition active:scale-95 disabled:opacity-50 disabled:pointer-events-none"
              >
                Save
              </button>
            </div>
          </div>
        {/if}

        {#if !places || places.length === 0}
          <!-- Empty State: Destinations -->
          <div class="text-center py-12 bg-teren-surface rounded-xl border border-teren-border border-dashed">
            <p class="text-teren-text-muted text-sm">{$t('detail.destinations_empty')}</p>
            <p class="text-xs text-teren-text-muted mt-1">{$t('detail.destinations_empty_sub')}</p>
          </div>
        {:else}
          <!-- Lista de Places -->
          <div class="space-y-3">
            {#each places as place (place.id)}
              <a 
                href="/trips/{tripId}/places/{place.id}" 
                class="block bg-teren-surface p-5 rounded-xl border border-teren-border hover:border-teren-primary/30 hover:shadow-md transition-all cursor-pointer group"
              >
                <div class="flex justify-between items-center mb-2">
                  <h3 class="text-lg font-semibold text-teren-text-main group-hover:text-teren-primary-hover transition-colors">{place.name}</h3>
                  <button 
                    onclick={(e) => { e.preventDefault(); e.stopPropagation(); requestDeletePlace(place.id); }}
                    class="opacity-0 group-hover:opacity-100 text-red-400 hover:text-red-600 p-1.5 rounded-lg hover:bg-red-50 transition active:scale-95 flex-shrink-0"
                    aria-label="Delete place"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
                
                <p class="text-sm text-teren-text-muted flex items-center gap-2">
                  <svg class="w-4 h-4 opacity-60" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/></svg>
                  {formatDate(place.start_date)} — {formatDate(place.end_date)}
                </p>
              </a>
            {/each}
          </div>
        {/if}
      </section>

    {/if}
  </main>

  <!-- Drawer para Gastos Globales -->
  <ExpenseDrawer 
    tripId={tripId}
    {categories}
    isOpen={isDrawerOpen}
    onClose={() => isDrawerOpen = false}
    onRefreshSummary={loadAllData}
  />
  
  <ConfirmModal
    isOpen={deletePlaceConfirmId !== null}
    title="Delete Destination"
    message="Are you sure you want to delete this destination? All associated activities will be lost."
    confirmText="Delete"
    cancelText="Cancel"
    isDestructive={true}
    onConfirm={confirmDeletePlace}
    onCancel={cancelDeletePlace}
  />
</div>