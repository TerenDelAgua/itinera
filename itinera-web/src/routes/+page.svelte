<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { apiFetch } from "$lib/api";
  import type { Trip } from "$lib/types/Trip";
  import { t, locale } from "$lib/i18n/store";
  import CreateTripForm from "$lib/components/CreateTripForm.svelte";
  import { getCurrencySymbol } from "$lib/utils";



  let trips = $state<Trip[]>([]);
  let isLoading = $state(true);
  let error = $state<string | null>(null);
  let isCreatingTrip = $state(false);

  onMount(async () => {
    try {
      const data = await apiFetch<Trip[]>("/trips");
      trips = data;
    } catch (err) {
      error = err instanceof Error ? err.message : "Failed to load trips";
    } finally {
      isLoading = false;
    }
  });

  function handleCreateTrip() {
    isCreatingTrip = true;
  }

  function handleTripCreated(newTrip: Trip) {
    trips = [newTrip, ...trips];
    isCreatingTrip = false;
  }

  function handleCancelCreate() {
    isCreatingTrip = false;
  }

  function formatDate(dateStr: string) {
    if (!dateStr) return "";
    // Note: ensure we treat the date string as local date to avoid TZ shifts
    const [year, month, day] = dateStr.split("-").map(Number);
    const date = new Date(year, month - 1, day);
    return new Intl.DateTimeFormat($locale, {
      day: "2-digit",
      month: "short",
      year: "numeric",
    }).format(date);
  }
</script>

<!-- Header -->
<section
  class="flex flex-col sm:flex-row sm:justify-between sm:items-start mb-10 gap-4 mt-4"
>
  <div>
    <h1
      class="text-[32px] sm:text-4xl font-bold text-teren-text-main tracking-tight"
    >
      {$t("dashboard.title")}
    </h1>
    <p class="text-teren-text-muted mt-2 text-base sm:text-lg">
      {$t("dashboard.subtitle")}
    </p>
  </div>

  <button
    onclick={handleCreateTrip}
    disabled={isCreatingTrip}
    class="bg-teren-primary hover:bg-teren-primary-hover text-white font-medium py-2.5 px-5 rounded-lg shadow-sm hover:shadow-md hover:shadow-orange-200 transition-all duration-200 flex items-center gap-2 active:scale-95 mt-1 sm:mt-0 disabled:opacity-50 disabled:active:scale-100"
  >
    <span class="text-lg leading-none">+</span>
    <span>{$t("dashboard.button_new")}</span>
  </button>
</section>

{#if isCreatingTrip}
  <CreateTripForm onsuccess={handleTripCreated} oncancel={handleCancelCreate} />
{/if}
{#if isLoading}
  <div class="flex justify-center items-center py-20">
    <div class="flex flex-col items-center gap-3">
      <div
        class="w-10 h-10 border-4 border-teren-primary/30 border-t-teren-primary rounded-full animate-spin"
      ></div>
      <p class="text-teren-text-muted font-medium animate-pulse">
        {$t("common.loading")}
      </p>
    </div>
  </div>
{:else if error}
  <div
    class="text-center py-12 bg-red-50 rounded-xl border border-red-100 text-red-700"
  >
    <p class="font-medium">⚠️ {error}</p>
    <button
      onclick={() => window.location.reload()}
      class="mt-4 text-sm underline hover:text-red-800">Try again</button
    >
  </div>
{:else if trips.length === 0}
  <div class="flex flex-col items-center justify-center py-20 text-center">
    <div
      class="w-16 h-16 mb-4 rounded-full bg-teren-surface border border-teren-border flex items-center justify-center"
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        class="h-8 w-8 text-teren-text-muted opacity-50"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="1.5"
          d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0121 18.382V7.618a1 1 0 01-.553-.894L15 7m0 13V7"
        />
      </svg>
    </div>

    <h3 class="text-xl font-medium text-teren-text-main mb-2">{$t("empty_state.title")}</h3>
    <p class="text-teren-text-muted max-w-xs mb-8">
      {$t("empty_state.description")}
    </p>

    <!-- Botón Secundario (Solo aquí, ya que el header siempre está visible) -->
    <button
      onclick={handleCreateTrip}
      class="text-teren-primary font-semibold hover:text-teren-primary-hover hover:underline transition-all"
    >
      {$t("empty_state.cta")} &rarr;
    </button>
  </div>
{:else}
  <div class="grid gap-6 sm:grid-cols-1 md:grid-cols-2 lg:grid-cols-2">
    {#each trips as trip (trip.id)}
      <div
        role="link"
        tabindex="0"
        onclick={() => goto(`/trips/${trip.id}`)}
        onkeydown={(e) => e.key === 'Enter' && goto(`/trips/${trip.id}`)}
        class="group bg-white rounded-xl border border-gray-200 hover:border-teren-primary/30 shadow-sm hover:shadow-xl hover:shadow-orange-900/5 transition-all duration-300 relative cursor-pointer hover:-translate-y-1"
      >
        <div class="p-6 flex flex-col h-full min-h-[180px]">
          <div class="flex justify-between items-start mb-2">
            <h2 class="text-xl font-bold text-teren-text-main group-hover:text-teren-primary transition-colors line-clamp-1">{trip.name}</h2>
            <span class="text-lg font-bold text-teren-primary whitespace-nowrap">
              {getCurrencySymbol(trip.base_currency)} {trip.total_spent?.toFixed(2) || '0.00'}
            </span>
          </div>
          
          <p class="text-teren-text-muted text-sm line-clamp-2 mb-6 flex-grow">
            {trip.description || ''}
          </p>

          <!-- Info Fila: Fechas y Destinos -->
            <div class="flex items-center gap-4 text-xs font-bold text-teren-text-muted mt-auto pt-4 border-t border-teren-border/50">
              <div class="flex items-center gap-1.5">
                <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" /></svg>
                {formatDate(trip.start_date)} - {formatDate(trip.end_date)}
              </div>
              {#if trip.place_count > 0}
                <div class="flex items-center gap-1.5">
                  <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
                  {trip.place_count} {trip.place_count === 1 ? $t('dashboard.destination') : $t('dashboard.destinations')}
                </div>
              {/if}
            </div>
        </div>
      </div>
    {/each}
  </div>
{/if}
