<script lang="ts">
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { apiFetch } from "$lib/api";
  import { t } from "$lib/i18n/store";
  import { cubicOut } from "svelte/easing";
  import { tweened } from "svelte/motion";
  import { untrack } from "svelte";

  // Types
  import type { Place } from "$lib/types/Place";
  import type { TripExpenseSummary } from "$lib/types/Summary";
  import type { Expense_Category, Trip } from "$lib/index";

  // Components
  import ExpenseDrawer from "$lib/components/ExpenseDrawer.svelte";
  import ConfirmModal from "$lib/components/ConfirmModal.svelte";
  import { activityApi } from "$lib/api/activity";
  import type { Activity } from "$lib/types/Activity";
  import ItineraryDrawer from "$lib/components/itinerary/ItineraryDrawer.svelte";
  import UpcomingActivityCard from "$lib/components/itinerary/UpcomingActivityCard.svelte";
  import DetailHeader from "$lib/components/trip/DetailHeader.svelte";
  import ExpensesSummaryCard from "$lib/components/trip/ExpensesSummaryCard.svelte";
  import PlaceList from "$lib/components/trip/PlaceList.svelte";

  let tripId = $state("");

  // State
  let tripName = $state("");
  let tripDescription = $state("");
  let tripStartDate = $state("");
  let tripEndDate = $state("");
  let baseCurrency = $state("EUR");
  let tripDefaultCurrency = $state("EUR");
  let effectiveCurrency = $derived(
    tripDefaultCurrency || baseCurrency || "EUR",
  );

  let places = $state<Place[]>([]);
  let summary = $state<TripExpenseSummary | null>(null);
  let categories = $state<Expense_Category[]>([]);
  let animatedGrandTotal = tweened(0, { duration: 600, easing: cubicOut });

  let isDrawerOpen = $state(false);
  let isLoading = $state(true);
  let deletePlaceConfirmId = $state<string | null>(null);

  let activities = $state<Activity[]>([]);
  let isAgendaOpen = $state(false);
  const refreshActivities = async () =>
    (activities = await activityApi.list(tripId));

  let initialActivityDate = $derived.by(() => {
    const today = new Date().toISOString().split("T")[0];
    if (!tripStartDate || !tripEndDate) return today;
    if (today >= tripStartDate && today <= tripEndDate) return today;
    return tripStartDate;
  });

  $effect(() => {
    if ($page.url.pathname) {
      tripId = $page.url.pathname.split("/").pop() || "";
      untrack(() => loadAllData());
    }
  });

  async function loadAllData() {
    if (!tripId) return;

    if (!summary) isLoading = true;

    try {
      const [tripData, placesData, summaryData, catsData, actsData] =
        await Promise.all([
          apiFetch<Trip>(`/trips/${tripId}`),
          apiFetch<Place[]>(`/trips/${tripId}/places`),
          apiFetch<TripExpenseSummary>(`/trips/${tripId}/expenses/summary`),
          apiFetch<Expense_Category[]>(`/trips/${tripId}/expenses/categories`),
          activityApi.list(tripId),
        ]);

      tripName = tripData.name;
      tripDescription = tripData.description || "";
      tripStartDate = tripData.start_date?.split("T")[0] || "";
      tripEndDate = tripData.end_date?.split("T")[0] || "";
      baseCurrency = tripData.base_currency || "EUR";
      tripDefaultCurrency =
        tripData.default_expense_currency || tripData.base_currency || "EUR";

      places = (placesData || []).map((p) => ({
        ...p,
        total_expenses:
          summaryData.by_place?.find((s) => s.place_id === p.id)?.total || 0,
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
        start_date: tripStartDate
          ? new Date(tripStartDate).toISOString()
          : null,
        end_date: tripEndDate ? new Date(tripEndDate).toISOString() : null,
        base_currency: baseCurrency,
        default_expense_currency: tripDefaultCurrency,
      };
      await apiFetch(`/trips/${tripId}`, {
        method: "PUT",
        body: JSON.stringify(payload),
      });
      // reload to sync with summary if currency changed
      loadAllData();
    } catch (e) {
      console.error("Failed to update trip", e);
    }
  }

  async function updateTripCurrency(code?: string) {
    try {
      if (!code) return;
      const previousBase = baseCurrency;
      baseCurrency = code;

      // If the default expense currency was the same as the old base, update it too
      // so we don't leave it behind unintentionally.
      const newDefault =
        tripDefaultCurrency === previousBase ? code : tripDefaultCurrency;
      tripDefaultCurrency = newDefault;

      await apiFetch(`/trips/${tripId}`, {
        method: "PUT",
        body: JSON.stringify({
          base_currency: code,
          default_expense_currency: newDefault,
        }),
      });
      loadAllData();
    } catch (e) {
      console.error("Failed to update trip currency", e);
      loadAllData(); // reset on error
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
      await apiFetch(`/trips/${tripId}/places/${id}`, { method: "DELETE" });
      loadAllData();
    } catch (e) {
      console.error("Failed to delete place", e);
    }
  }

  function cancelDeletePlace() {
    deletePlaceConfirmId = null;
  }
</script>

<div class="min-h-screen bg-teren-background pb-20">
  <DetailHeader
    bind:name={tripName}
    bind:description={tripDescription}
    bind:startDate={tripStartDate}
    bind:endDate={tripEndDate}
    defaultCurrency={baseCurrency}
    onSave={saveTripInfo}
    onUpdateCurrency={updateTripCurrency}
    onBack={() => goto("/")}
  />

  <main class="max-w-3xl mx-auto px-4 py-8 space-y-12">
    {#if isLoading}
      <!-- Skeleton Loading -->
      <div class="animate-pulse space-y-6">
        <div
          class="h-40 bg-teren-surface rounded-xl border border-teren-border"
        ></div>
        <div
          class="h-64 bg-teren-surface rounded-xl border border-teren-border"
        ></div>
      </div>
    {:else}
      <!-- ========================================== -->
      <!-- 1. GLOBAL EXPENSES CARD -->
      <!-- ========================================== -->
      <ExpensesSummaryCard
        {tripId}
        {categories}
        categorySummary={summary?.by_category || []}
        {baseCurrency}
        {tripDefaultCurrency}
        {effectiveCurrency}
        grandTotalValue={$animatedGrandTotal}
        tripStart={tripStartDate}
        tripEnd={tripEndDate}
        onRefresh={loadAllData}
        onOpenDrawer={() => (isDrawerOpen = true)}
      />

      <!-- ========================================== -->
      <!-- 2. ACTIVITIES SECTION -->
      <!-- ========================================== -->

      <UpcomingActivityCard
        {activities}
        {tripId}
        tripStart={tripStartDate}
        tripEnd={tripEndDate}
        defaultDate={initialActivityDate}
        onRefresh={loadAllData}
        onOpenDrawer={() => (isAgendaOpen = true)}
      />

      <!-- ========================================== -->
      <!-- 3. DESTINATIONS (PLACES) -->
      <!-- ========================================== -->
      <PlaceList
        {tripId}
        {places}
        {baseCurrency}
        onRefresh={loadAllData}
        onRequestDelete={requestDeletePlace}
      />
    {/if}
  </main>

  <!-- Drawers and Modals outside of loading context -->
  <ExpenseDrawer
    {tripId}
    {categories}
    isOpen={isDrawerOpen}
    onClose={() => (isDrawerOpen = false)}
    onRefreshSummary={loadAllData}
  />

  <ConfirmModal
    isOpen={deletePlaceConfirmId !== null}
    title={$t("confirm.delete_place_title")}
    message={$t("confirm.delete_place_message")}
    confirmText={$t("common.delete")}
    cancelText={$t("common.cancel")}
    isDestructive={true}
    onConfirm={confirmDeletePlace}
    onCancel={cancelDeletePlace}
  />

  <ItineraryDrawer
    isOpen={isAgendaOpen}
    {tripId}
    tripStart={tripStartDate}
    tripEnd={tripEndDate}
    {activities}
    defaultDate={initialActivityDate}
    onRefresh={refreshActivities}
    onClose={() => (isAgendaOpen = false)}
  />
</div>
