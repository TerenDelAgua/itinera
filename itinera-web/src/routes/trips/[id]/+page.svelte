<script lang="ts">
  import { goto, afterNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import { onMount } from "svelte";
  import { apiFetch } from "$lib/api";
  import { t } from "$lib/i18n/store";
  import { Events } from "$lib/services/tracking";
  import { cubicOut } from "svelte/easing";
  import { tweened } from "svelte/motion";
  import { untrack } from "svelte";
  import { fly } from "svelte/transition";

  // Types
  import type { Place } from "$lib/types/Place";
  import type { TripExpenseSummary } from "$lib/types/Summary";
  import type { Expense_Category, Trip } from "$lib/index";

  // Components
  import ExpenseDrawer from "$lib/components/Expenses/ExpenseDrawer.svelte";
  import ConfirmModal from "$lib/components/utils/ConfirmModal.svelte";
  import { activityApi } from "$lib/api/activity";
  import type { Activity } from "$lib/types/Activity";
  import ActivityDrawer from "$lib/components/activities/ActivityDrawer.svelte";
  import UpcomingActivityCard from "$lib/components/activities/UpcomingActivityCard.svelte";
  import DetailHeader from "$lib/components/trip/DetailHeader.svelte";
  import ExpensesSummaryCard from "$lib/components/trip/ExpensesSummaryCard.svelte";
  import PlaceList from "$lib/components/trip/PlaceList.svelte";
  import TransportBadge from "$lib/components/trip/TransportBadge.svelte";

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

  let tripDurationDays = $derived.by(() => {
    if (!tripStartDate || !tripEndDate) return 0;
    const start = new Date(tripStartDate);
    const end = new Date(tripEndDate);
    const diff = end.getTime() - start.getTime();
    return Math.max(0, Math.ceil(diff / (1000 * 3600 * 24))) + 1;
  });

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

  onMount(() => {
    tripId = $page.params.id ?? "";
    if (tripId) {
      loadAllData();
    }
  });

  afterNavigate(({ to }) => {
    const newId = to?.params?.id;
    if (newId && newId !== tripId) {
      tripId = newId;
      loadAllData();
    }
  });

  async function loadAllData() {
    if (!tripId) return;

    if (!summary) isLoading = true;

    try {
      const tripData = await apiFetch<Trip>(`/trips/${tripId}`);

      tripName = tripData.name;
      tripDescription = tripData.description || "";
      tripStartDate = tripData.start_date?.split("T")[0] || "";
      tripEndDate = tripData.end_date?.split("T")[0] || "";
      baseCurrency = tripData.base_currency || "EUR";
      tripDefaultCurrency =
        tripData.default_expense_currency || tripData.base_currency || "EUR";

      const [placesData, summaryData, catsData, actsData] = await Promise.all([
        apiFetch<Place[]>(`/trips/${tripId}/places`),
        apiFetch<TripExpenseSummary>(
          `/trips/${tripId}/expenses/summary?currency=${baseCurrency}`,
        ),
        apiFetch<Expense_Category[]>(`/trips/${tripId}/expenses/categories`),
        activityApi.list(tripId),
      ]);

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

      if (tripData.is_public_demo) {
        Events.demoViewed(tripId, tripName);
      }
    } catch (e) {
      console.error("Failed to load trip data", e);
    } finally {
      isLoading = false;
    }
  }

  let isSaving = $state(false);
  let saveTimer: ReturnType<typeof setTimeout>;

  function saveTripInfo() {
    clearTimeout(saveTimer);
    saveTimer = setTimeout(async () => {
      if (!tripName || isSaving) return;
      isSaving = true;
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
      const response = await apiFetch<Trip>(`/trips/${tripId}`, {
        method: "PUT",
        body: JSON.stringify(payload),
      });

      // reload to sync with summary if currency changed
      loadAllData();
    } catch (e) {
      console.error("Failed to update trip", e);
    } finally {
        isSaving = false;
      }
    }, 300);
  }

  async function updateTripCurrency(code?: string) {
    if (!code || isSaving) return;
    isSaving = true;
    try {
      const previousBase = baseCurrency;
      baseCurrency = code;

      // If the default expense currency was the same as the old base, update it too
      // so we don't leave it behind unintentionally.
      const newDefault =
        tripDefaultCurrency === previousBase ? code : tripDefaultCurrency;
      tripDefaultCurrency = newDefault;

      const response = await apiFetch<Trip>(`/trips/${tripId}`, {
        method: "PUT",
        body: JSON.stringify({
          base_currency: code,
          default_expense_currency: newDefault,
        }),
      });

      // reload to sync with summary if currency changed
      loadAllData();
    } catch (e) {
      console.error("Failed to update currency", e);
    } finally {
      isSaving = false;
    }
  }

  function handleDeletePlace(id: string) {
    deletePlaceConfirmId = id;
  }

  async function confirmDeletePlace() {
    if (!deletePlaceConfirmId) return;
    try {
      await apiFetch(`/trips/${tripId}/places/${deletePlaceConfirmId}`, {
        method: "DELETE",
      });
      loadAllData();
    } catch (err) {
      console.error("Error deleting place", err);
    } finally {
      deletePlaceConfirmId = null;
    }
  }
</script>

<svelte:head>
  <title>{tripName ? `${tripName} | Itinera` : $t("common.loading")}</title>
</svelte:head>

<div class="min-h-screen bg-teren-background pb-20">
  <DetailHeader
    bind:name={tripName}
    bind:description={tripDescription}
    bind:startDate={tripStartDate}
    bind:endDate={tripEndDate}
    bind:defaultCurrency={baseCurrency}
    onSave={saveTripInfo}
    onUpdateCurrency={updateTripCurrency}
    onBack={() => goto("/trips")}
  />

  <main class="max-w-3xl mx-auto px-4 py-8 space-y-10">
    {#if isLoading}
      <div class="animate-pulse space-y-6">
        <div
          class="h-48 bg-teren-surface rounded-xl border border-teren-border"
        ></div>
        <div
          class="h-32 bg-teren-surface rounded-xl border border-teren-border"
        ></div>
      </div>
    {:else}
      <div in:fly={{ y: 20, duration: 400, easing: cubicOut }}>
        <UpcomingActivityCard
          {tripId}
          {activities}
          tripStart={tripStartDate}
          tripEnd={tripEndDate}
          onOpenDrawer={() => (isAgendaOpen = true)}
          onRefresh={loadAllData}
        />
      </div>

      <ExpensesSummaryCard
        {tripId}
        {categories}
        categorySummary={summary?.by_category || []}
        displayCurrency={baseCurrency}
        tripDefaultCurrency={tripDefaultCurrency}
        effectiveCurrency={effectiveCurrency}
        grandTotalValue={$animatedGrandTotal}
        tripStart={tripStartDate}
        tripEnd={tripEndDate}
        onRefresh={loadAllData}
        onOpenDrawer={() => (isDrawerOpen = true)}
      />

      <TransportBadge {places} {tripDurationDays} {tripId} {baseCurrency} />

      <PlaceList
        {places}
        {tripId}
        {baseCurrency}
        onRefresh={loadAllData}
        onRequestDelete={handleDeletePlace}
      />
    {/if}
  </main>

  <ExpenseDrawer
    {tripId}
    {categories}
    isOpen={isDrawerOpen}
    onClose={() => (isDrawerOpen = false)}
    onRefreshSummary={loadAllData}
  />

  <ActivityDrawer
    isOpen={isAgendaOpen}
    {tripId}
    tripStart={tripStartDate}
    tripEnd={tripEndDate}
    {activities}
    onRefresh={loadAllData}
    onClose={() => (isAgendaOpen = false)}
  />

  <ConfirmModal
    isOpen={deletePlaceConfirmId !== null}
    title={$t("confirm.delete_place_title")}
    message={$t("confirm.delete_place_message")}
    confirmText={$t("common.delete")}
    cancelText={$t("common.cancel")}
    isDestructive={true}
    onConfirm={confirmDeletePlace}
    onCancel={() => (deletePlaceConfirmId = null)}
  />
</div>
