<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { apiFetch } from '$lib/api';
  import { t, locale } from "$lib/i18n/store";
  import { slide } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import { tweened } from 'svelte/motion';
  import { untrack } from 'svelte';
  import { formatDisplayDate, formatDate } from '$lib/utils/date';

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
  import { activityApi } from '$lib/api/activity';
  import type { Activity } from '$lib/types/Activity';
  import UpcomingActivityCard from '$lib/components/upcomingActivityCard.svelte';
  import ActivitiesDrawer from '$lib/components/ActivitiesDrawer.svelte';

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
  let isMobileExpenseOpen = $state(false);
  let newPlaceDraft = $state({ name: '', start_date: '', end_date: '' });
  let deletePlaceConfirmId = $state<string | null>(null);

  let activities = $state<Activity[]>([]);
  let isAgendaOpen = $state(false);

  $effect(() => {
    if ($page.url.pathname) {
      tripId = $page.url.pathname.split('/').pop() || '';
      untrack(() => loadAllData());
    }
  });

  async function loadAllData() {
    if (!tripId) return;
    
    // Only show skeleton on initial load
    if (!summary) isLoading = true; 
    
    try {
      const [tripData, placesData, summaryData, catsData, actsData] = await Promise.all([
        apiFetch<Trip>(`/trips/${tripId}`),
        apiFetch<Place[]>(`/trips/${tripId}/places`),
        apiFetch<TripExpenseSummary>(`/trips/${tripId}/expenses/summary`),
        apiFetch<Expense_Category[]>(`/trips/${tripId}/expenses/categories`),
        activityApi.list(tripId)
      ]);

      tripName = tripData.name;
      tripDescription = tripData.description || '';
      tripDates = { start: tripData.start_date, end: tripData.end_date };
      currency = tripData.base_currency || 'EUR';
      
      places = (placesData || []).map(p => ({
        ...p,
        total_expenses: summaryData.by_place?.find(s => s.place_id === p.id)?.total || 0
      }));
      summary = summaryData;
      categories = catsData || [];
      activities = actsData || [];
      
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
      const payload = { name: newPlaceDraft.name };
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


  function formatSmartDate(dateStr?: string) {
    return formatDisplayDate(dateStr, $t, $locale);
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
              placeholder={$t('trip_form.name')}
              class="bg-transparent border-none p-0 focus:ring-0 text-xl font-bold text-teren-text-main leading-tight outline-none w-full truncate" 
            />
            <div class="relative ml-3">
              <select 
                bind:value={currency} 
                onchange={saveTripInfo} 
                class="appearance-none bg-teren-primary-subtle text-teren-primary border border-teren-primary/20 rounded-full px-2.5 py-0.5 text-[10px] font-bold cursor-pointer hover:bg-teren-primary hover:text-white transition-all outline-none"
              >
                <option value="EUR">EUR</option>
                <option value="USD">USD</option>
                <option value="GBP">GBP</option>
                <option value="JPY">JPY</option>
              </select>
            </div>
          </div>
          <div class="flex items-center gap-1.5 text-xs text-teren-text-muted font-medium">
            <svg class="w-3.5 h-3.5 opacity-70 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/></svg>
            
            <div class="relative flex items-center">
              <span class="text-teren-text-main">{formatDate(tripDates.start, $locale)}</span>
              <input type="date" bind:value={tripDates.start} onchange={saveTripInfo} class="absolute inset-0 opacity-0 cursor-pointer w-full" />
            </div>

            <span class="opacity-70 mx-0.5">—</span>

            <div class="relative flex items-center">
              <span class="text-teren-text-main">{formatDate(tripDates.end, $locale)}</span>
              <input type="date" bind:value={tripDates.end} onchange={saveTripInfo} class="absolute inset-0 opacity-0 cursor-pointer w-full" />
            </div>
          </div>
          <input 
            type="text" 
            bind:value={tripDescription} 
            onblur={saveTripInfo}
            onkeydown={(e) => e.key === 'Enter' && e.currentTarget.blur()}
            placeholder={$t('detail.description')}
            class="bg-transparent border-none p-0 focus:ring-0 text-sm text-teren-text-muted outline-none w-full truncate mt-0.5 italic" 
          />
        </div>
      </div>
    </div>
  </header>

  <main class="max-w-3xl mx-auto px-4 py-8 space-y-12">
    
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
        <div class="flex flex-col sm:flex-row sm:justify-between sm:items-center mb-6 gap-3 sm:gap-0">
          <div class="flex justify-between items-center">
            <h2 class="text-lg font-semibold text-teren-text-main tracking-tight">{$t('detail.expenses')}</h2>
            <button 
              onclick={() => isMobileExpenseOpen = !isMobileExpenseOpen} 
              class="sm:hidden text-sm font-medium text-teren-primary hover:text-teren-primary-hover transition px-3 py-1.5 rounded-lg bg-teren-primary-subtle active:scale-95 flex-shrink-0"
            >
              + {$t('common.add')}
            </button>
          </div>
          <div class="flex items-center">
            <span class="text-3xl sm:text-2xl font-bold text-teren-primary tabular-nums leading-none">
              {$animatedGrandTotal.toFixed(2)} {getCurrencySymbol(currency)}
            </span>
          </div>
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

        <!-- Quick Add (Desktop Context) -->
        <div class="hidden sm:block mt-6">
          <ExpenseQuickAdd 
            tripId={tripId} 
            {categories} 
            {currency}
            onSuccess={loadAllData} 
          />
        </div>

        <!-- Quick Add (Mobile Context) -->
        {#if isMobileExpenseOpen}
          <div class="sm:hidden mt-4" transition:slide={{ duration: 250, easing: cubicOut }}>
            <ExpenseQuickAdd 
              tripId={tripId} 
              {categories} 
              {currency}
              onSuccess={() => { loadAllData(); isMobileExpenseOpen = false; }} 
            />
          </div>
        {/if}

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
      <!-- 2. ACTIVITIES SECTION -->
      <!-- ========================================== -->
      <UpcomingActivityCard 
        {tripId}
        activities={activities} 
        onOpenDrawer={() => isAgendaOpen = true} 
        onRefresh={loadAllData}
      />

      <!-- ========================================== -->
      <!-- 3. DESTINATIONS (PLACES) -->
      <!-- ========================================== -->
      <section>
        <div class="flex justify-between items-center mb-4">
          <h2 class="text-lg font-semibold text-teren-text-main tracking-tight">{$t('detail.destinations')}</h2>
          <button 
            onclick={() => isCreatingPlace = !isCreatingPlace}
            class="text-sm font-medium text-teren-primary hover:text-teren-primary-hover transition px-3 py-1.5 rounded-lg hover:bg-teren-primary-subtle active:scale-95"
          >
            + {$t('common.add')}
          </button>
        </div>

        {#if isCreatingPlace}
          <div class="mb-4 p-4 bg-teren-background border-2 border-teren-primary/30 rounded-xl space-y-3" transition:slide={{ duration: 250, easing: cubicOut }}>
            <input 
              type="text" 
              bind:value={newPlaceDraft.name} 
              placeholder={$t('place_form.name_placeholder')} 
              class="w-full px-3 py-2 text-sm font-bold bg-white border border-teren-border rounded-lg focus:ring-2 focus:ring-teren-primary/30 outline-none" 
              autofocus 
              onkeydown={e => e.key === 'Enter' && createPlace()}
            />
            <div class="flex flex-wrap gap-3 items-end">
              <div class="w-[calc(50%-0.375rem)] sm:flex-1 min-w-0">
                <label class="block text-xs text-teren-text-muted mb-1 ml-1 font-medium">{$t('place_form.start_date')}</label>
                <input 
                  type="date" 
                  bind:value={newPlaceDraft.start_date} 
                  class="w-full min-w-0 px-2 sm:px-3 py-2 text-sm bg-white border border-teren-border rounded-lg focus:ring-2 focus:ring-teren-primary/30 outline-none" 
                  onkeydown={e => e.key === 'Enter' && createPlace()}
                />
              </div>
              <div class="w-[calc(50%-0.375rem)] sm:flex-1 min-w-0">
                <label class="block text-xs text-teren-text-muted mb-1 ml-1 font-medium">{$t('place_form.end_date')}</label>
                <input 
                  type="date" 
                  bind:value={newPlaceDraft.end_date} 
                  class="w-full min-w-0 px-2 sm:px-3 py-2 text-sm bg-white border border-teren-border rounded-lg focus:ring-2 focus:ring-teren-primary/30 outline-none" 
                  onkeydown={e => e.key === 'Enter' && createPlace()}
                />
              </div>
              
              <!-- Separador para forzar el botón a la 3ra fila en móvil -->
              <div class="w-full h-0 sm:hidden"></div>

              <div class="w-full sm:w-auto flex justify-end">
                <button 
                  onclick={createPlace} 
                  disabled={isLoading || !newPlaceDraft.name} 
                  class="flex-shrink-0 bg-teren-primary hover:bg-teren-primary-hover text-white h-[38px] px-8 sm:px-4 rounded-lg font-bold shadow-sm active:scale-95 transition-all disabled:opacity-50 disabled:active:scale-100 flex items-center justify-center"
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
        {/if}

        {#if !places || places.length === 0}
          <!-- Empty State: Destinations -->
          <div class="text-center py-12 bg-teren-surface rounded-xl border border-teren-border border-dashed">
            <p class="text-teren-text-muted text-sm">{$t('detail.destinations_empty')}</p>
            <p class="text-xs text-teren-text-muted mt-1">{$t('detail.destinations_empty_sub')}</p>
          </div>
        {:else}
          <!-- Lista de Places -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            {#each places as place (place.id)}
        <a href="/trips/{tripId}/places/{place.id}" class="group block bg-teren-surface p-5 rounded-xl border border-teren-border hover:border-teren-primary/30 hover:shadow-md transition-all cursor-pointer relative">
          
          <!-- Contenido Principal -->
          <div class="flex flex-col gap-1">
            <h3 class="text-lg font-semibold text-teren-text-main group-hover:text-teren-primary-hover transition-colors pr-8">
              {place.name}
            </h3>
            
            <div class="flex justify-between items-center">
              <div class="text-sm text-teren-text-muted flex items-center gap-1.5">
                <svg class="w-4 h-4 opacity-60" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
                </svg>
                <span>{formatSmartDate(place.start_date)} — {place.end_date ? formatSmartDate(place.end_date) : 'No end'}</span>
              </div>
 
              {#if place.total_expenses !== undefined}
                <div class="text-right pl-2">
                  <span class="font-bold {place.total_expenses > 0 ? 'text-teren-primary' : 'text-[11px] text-teren-text-muted opacity-50'}">
                    {place.total_expenses.toFixed(2)} {getCurrencySymbol(currency)}
                  </span>
                </div>
              {/if}
            </div>
          </div>  

          <!-- Botón Borrar (Aparece en Hover) -->
          <button 
            onclick={(e) => { e.preventDefault(); e.stopPropagation(); requestDeletePlace(place.id); }}
            class="absolute top-3 right-3 opacity-0 group-hover:opacity-100 text-red-400 hover:text-red-600 p-2 rounded-lg hover:bg-red-50 transition active:scale-95"
            aria-label="Delete place"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </button>
        </a>
      {/each}
    </div>
    {/if}
  </section>

  {/if}
  </main>


  <!-- Drawers and Modals outside of loading context -->
  <ExpenseDrawer 
    tripId={tripId}
    {categories}
    isOpen={isDrawerOpen}
    onClose={() => isDrawerOpen = false}
    onRefreshSummary={loadAllData}
  />
  
  <ConfirmModal
    isOpen={deletePlaceConfirmId !== null}
    title={$t('confirm.delete_place_title')}
    message={$t('confirm.delete_place_message')}
    confirmText={$t('common.delete')}
    cancelText={$t('common.cancel')}
    isDestructive={true}
    onConfirm={confirmDeletePlace}
    onCancel={cancelDeletePlace}
  />

  <ActivitiesDrawer 
    isOpen={isAgendaOpen}
    tripId={tripId}
    activities={activities}
    onRefresh={loadAllData}
    onClose={() => isAgendaOpen = false}
  />
</div>