<script lang="ts">
  import { onMount } from "svelte";
  import { apiFetch } from "$lib/api";
  import type { Trip } from "$lib/types/Trip";
  import { t } from "$lib/i18n/store";

  import { Events } from "$lib/services/tracking";

  // Components
  import CreateTripForm from "$lib/components/dashboard/CreateTripForm.svelte";
  import ConfirmModal from "$lib/components/utils/ConfirmModal.svelte";
  import DashboardHeader from "$lib/components/dashboard/DashboardHeader.svelte";
  import EmptyState from "$lib/components/dashboard/EmptyState.svelte";
  import TripCard from "$lib/components/dashboard/TripCard.svelte";

  let trips = $state<Trip[]>([]);
  let isLoading = $state(true);
  let error = $state<string | null>(null);
  let isCreatingTrip = $state(false);

  // Derivados para separación visual
  let userTrips = $derived(trips.filter((t) => !t.is_public_demo));
  let demoTrips = $derived(trips.filter((t) => t.is_public_demo));

  // Estado para borrar
  let tripToDelete = $state<Trip | null>(null);

  onMount(async () => {
    try {
      const data = await apiFetch<Trip[]>("/trips");
      trips = data;
      Events.sessionStarted(document.referrer || "direct");
      Events.landingView();
    } catch (err) {
      error = err instanceof Error ? err.message : "Failed to load trips";
    } finally {
      isLoading = false;
    }
  });

  function handleCreateTrip() {
    isCreatingTrip = !isCreatingTrip;
  }

  function handleTripCreated(newTrip: Trip) {
    trips = [newTrip, ...trips];
    isCreatingTrip = false;
  }

  // Lógica de Borrado
  function handleDeleteClick(trip: Trip, event: Event) {
    event.stopPropagation(); // Evita navegar al hacer clic en borrar
    tripToDelete = trip;
  }

  async function confirmDelete() {
    if (!tripToDelete) return;
    const id = tripToDelete.id;
    try {
      await apiFetch(`/trips/${id}`, { method: "DELETE" });
      trips = trips.filter((t) => t.id !== id);
    } catch (err) {
      console.error("Error deleting trip", err);
    } finally {
      tripToDelete = null;
    }
  }

  function cancelDelete() {
    tripToDelete = null;
  }
</script>

<div class="max-w-5xl mx-auto px-4 py-8">
  <DashboardHeader {isCreatingTrip} onCreateClick={handleCreateTrip} />

  <!-- FORMULARIO INLINE -->
  {#if isCreatingTrip}
    <div class="mb-10 animate-in fade-in slide-in-from-top-4 duration-300">
      <CreateTripForm onsuccess={handleTripCreated} />
    </div>
  {/if}

  <!-- ESTADOS DE CARGA / ERROR / VACÍO -->
  {#if isLoading}
    <div class="flex justify-center items-center py-20">
      <div
        class="w-10 h-10 border-4 border-teren-primary/30 border-t-teren-primary rounded-full animate-spin"
      ></div>
    </div>
  {:else if error}
    <div
      class="text-center py-12 text-error-base bg-error-subtle rounded-xl border border-error-base/20"
    >
      <p class="font-medium">{$t("common.error")}: {error}</p>
      <button
        onclick={() => window.location.reload()}
        class="mt-4 text-sm underline hover:text-error-hover"
      >
        {$t("common.try_again")}
      </button>
    </div>
  {:else if trips.length === 0}
    <EmptyState onCreateClick={handleCreateTrip} />
  {:else}
    <div class="space-y-8">
      {#if userTrips.length > 0}
        <section class="animate-in fade-in slide-in-from-bottom-4 duration-500">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            {#each userTrips as trip (trip.id)}
              <TripCard {trip} onDeleteClick={handleDeleteClick} />
            {/each}
          </div>
        </section>
      {/if}

      {#if userTrips.length > 0 && demoTrips.length > 0}
        <div class="border-t border-teren-border/40"></div>
      {/if}

      {#if demoTrips.length > 0}
        <section class="animate-in fade-in slide-in-from-bottom-4 duration-700">
          <h3
            class="text-sm font-bold text-teren-text-muted uppercase tracking-wider mb-5 px-1 flex items-center gap-2"
          >
            <svg
              class="w-4 h-4 text-orange-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M13 10V3L4 14h7v7l9-11h-7z"
              />
            </svg>
            {$t("dashboard.inspiration")}
          </h3>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            {#each demoTrips as trip (trip.id)}
              <TripCard {trip} onDeleteClick={handleDeleteClick} />
            {/each}
          </div>
        </section>
      {/if}
    </div>
  {/if}

  <!-- MODAL DE BORRADO -->
  <ConfirmModal
    isOpen={tripToDelete !== null}
    title={$t("confirm.delete_trip_title")}
    message={$t("confirm.delete_trip_message", {
      tripName: tripToDelete?.name || "",
    })}
    confirmText={$t("common.delete")}
    cancelText={$t("common.cancel")}
    isDestructive={true}
    onConfirm={confirmDelete}
    onCancel={cancelDelete}
  />
</div>
