<script lang="ts">
  import { onMount } from "svelte";
  import { apiFetch } from "$lib/api";
  import type { Trip } from "$lib/types/trip";

  let trips = $state<Trip[]>([]);
  let isLoading = $state(true);
  let error = $state<string | null>(null);

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
    console.log("New experience");
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
      My trips
    </h1>
    <p class="text-teren-text-muted mt-2 text-base sm:text-lg">
      Plan and organize all your adventures
    </p>
  </div>

  <button
    onclick={handleCreateTrip}
    class="bg-teren-primary hover:bg-teren-primary-hover text-white font-medium py-2.5 px-5 rounded-lg shadow-sm hover:shadow-md hover:shadow-orange-200 transition-all duration-200 flex items-center gap-2 active:scale-95 mt-1 sm:mt-0"
  >
    <span class="text-lg leading-none">+</span>
    <span>New trip</span>
  </button>
</section>

<!-- Trips grid -->
{#if isLoading}
  <div class="flex justify-center items-center py-20">
    <div class="flex flex-col items-center gap-3">
      <div
        class="w-10 h-10 border-4 border-teren-primary/30 border-t-teren-primary rounded-full animate-spin"
      ></div>
      <p class="text-teren-text-muted font-medium animate-pulse">
        Loading adventures...
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
  <div
    class="text-center py-20 bg-teren-surface rounded-xl border border-teren-border shadow-sm"
  >
    <div class="text-5xl mb-4">🧳</div>
    <h3 class="text-xl font-bold text-teren-text-main mb-2">No trips</h3>
    <p class="text-teren-text-muted mb-6">
      Start planning your first adventure!
    </p>
    <button
      title="New trip"
      onclick={handleCreateTrip}
      class="bg-teren-primary hover:bg-teren-primary-hover text-white font-bold py-3 px-8 rounded-lg shadow-md transition-all active:scale-95"
    >
      Start your new adventure!
    </button>
  </div>
{:else}
  <div class="grid gap-6 sm:grid-cols-1 md:grid-cols-2 lg:grid-cols-2">
    {#each trips as trip}
      <!-- Card Estilo TEREN -->
      <div
        class="group bg-white p-6 rounded-xl border border-gray-200 hover:border-teren-primary/30 shadow-sm hover:shadow-xl hover:shadow-orange-900/5 transition-all duration-300 relative cursor-pointer hover:-translate-y-1"
      >
        <!-- Header Card: Nombre + Menú 3 puntos -->
        <div class="flex justify-between items-start mb-1">
          <h3
            class="text-xl font-bold text-teren-text-main tracking-tight group-hover:text-teren-primary-hover transition-colors"
          >
            {trip.name}
          </h3>
          <button
            title="Options"
            class="text-gray-400 hover:text-teren-text-main hover:bg-gray-50 rounded p-1.5 transition-colors -mr-1.5 -mt-1.5 active:scale-95"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <circle cx="12" cy="12" r="1"></circle>
              <circle cx="12" cy="5" r="1"></circle>
              <circle cx="12" cy="19" r="1"></circle>
            </svg>
          </button>
        </div>

        <!-- Duración -->
        <p class="text-[15px] text-teren-text-muted mb-5 font-medium">
          {trip.start_date} - {trip.end_date}
        </p>

        <!-- Fila Central: Iconos + Datos -->
        <div
          class="flex items-center gap-5 text-[15px] text-gray-500 mb-6 font-medium"
        >
          <!-- Destinos -->
          <div class="flex items-center gap-1.5">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="text-gray-400"
            >
              <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path>
              <circle cx="12" cy="10" r="3"></circle>
            </svg>
            <span>{trip.description || "0"} destinos</span>
          </div>

          <!-- Fechas -->
          <div class="flex items-center gap-1.5">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="text-gray-400"
            >
              <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
              <line x1="16" y1="2" x2="16" y2="6"></line>
              <line x1="8" y1="2" x2="8" y2="6"></line>
              <line x1="3" y1="10" x2="21" y2="10"></line>
            </svg>
            <span>{trip.start_date} - {trip.end_date}</span>
          </div>
        </div>

        <!-- Footer Card: Gasto Total (Badge) -->
        <div>
          <span
            class="inline-block bg-teren-primary-subtle text-teren-primary-hover text-[13px] font-bold px-3 py-1 rounded-full uppercase"
          >
            {trip.base_currency} 0.00
          </span>
        </div>
      </div>
    {/each}
  </div>
{/if}
