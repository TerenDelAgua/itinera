<script lang="ts">
  import { onMount } from "svelte";
  import { apiFetch } from "$lib/api";
  import type { Trip } from "$lib/types/Trip";
  import { t } from "$lib/i18n/store";

  // Components
  import CreateTripForm from "$lib/components/CreateTripForm.svelte";
  import ConfirmModal from "$lib/components/ConfirmModal.svelte";
  import DashboardHeader from "$lib/components/dashboard/DashboardHeader.svelte";
  import EmptyState from "$lib/components/dashboard/EmptyState.svelte";
  import TripCard from "$lib/components/dashboard/TripCard.svelte";

  let trips = $state<Trip[]>([]);
  let isLoading = $state(true);
  let error = $state<string | null>(null);
  let isCreatingTrip = $state(false);

  // Estado para borrar
  let tripToDelete = $state<Trip | null>(null);

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

  // Lógica de Borrado
  function handleDeleteClick(trip: Trip, event: Event) {
    event.stopPropagation(); // Evita navegar al hacer clic en borrar
    tripToDelete = trip;
  }

  async function confirmDelete() {
    if (!tripToDelete) return;
    try {
      await apiFetch(`/trips/${tripToDelete.id}`, { method: "DELETE" });
      trips = trips.filter((t) => t.id !== tripToDelete.id);
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
      <CreateTripForm
        onsuccess={handleTripCreated}
        oncancel={handleCancelCreate}
      />
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
    <!-- GRID DE TARJETAS -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      {#each trips as trip (trip.id)}
        <TripCard {trip} onDeleteClick={handleDeleteClick} />
      {/each}
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
