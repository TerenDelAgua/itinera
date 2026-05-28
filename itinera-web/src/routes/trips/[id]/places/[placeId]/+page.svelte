<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { apiFetch } from "$lib/api";
  import { fly, slide } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import { untrack, onMount } from "svelte";
  import { t, locale } from "$lib/i18n/store";
  import ActivityDrawer from "$lib/components/activities/ActivityDrawer.svelte";
  import { formatDate } from "$lib/utils/date";
  import { activityApi } from "$lib/api/activity";

  import type { Place } from "$lib/types/Place";
  import type { CategorySummary } from "$lib/types/Summary";
  import type { Expense, Expense_Category } from "$lib/index";
  import type { Trip } from "$lib/types/Trip";

  import ExpenseQuickAdd from "$lib/components/Expenses/ExpenseQuickAdd.svelte";
  import ExpenseDrawer from "$lib/components/Expenses/ExpenseDrawer.svelte";
  import ExpenseSummaryPills from "$lib/components/Expenses/ExpenseSummaryPills.svelte";
  import CurrencySelector from "$lib/components/currency/CurrencySelector.svelte";
  import UpcomingActivityCard from "$lib/components/activities/UpcomingActivityCard.svelte";
  import DetailHeader from "$lib/components/trip/DetailHeader.svelte";
  import ExpensesSummaryCard from "$lib/components/trip/ExpensesSummaryCard.svelte";
  import type { Activity } from "$lib/types/Activity";

  // Station Guides v1.2 imports
  import StationGuide from "$lib/components/station-guide/StationGuide.svelte";
  import { matchStationGuide } from "$lib/services/stationGuide";

  // Japan Context v1.3.1 imports
  import PlaceContextWidget from "$lib/components/japan-context/PlaceContextWidget.svelte";
  import PhraseDrawer from "$lib/components/japan-context/PhraseDrawer.svelte";
  import { getPlaceLevelRules } from "$lib/services/japanContext";

  let tripId = $state("");
  let placeId = $state("");

  let place = $state<Place | null>(null);
  let matchedGuide = $derived(place ? matchStationGuide(place.name, place.city) : null);
  let placeRules = $derived(place && place.city ? getPlaceLevelRules(place.name, place.city, $locale) : []);
  let expenses = $state<Expense[]>([]);
  let categorySummary = $state<CategorySummary[]>([]);
  let categories = $state<Expense_Category[]>([]);
  let tripBaseCurrency = $state("EUR");
  let tripDefaultCurrency = $state("EUR");
  let tripStartDate = $state("");
  let tripEndDate = $state("");

  let isDrawerOpen = $state(false);
  let isLoading = $state(true);

  let activities = $state<Activity[]>([]);
  let isAgendaOpen = $state(false);

  onMount(() => {
    tripId = $page.params.id ?? "";
    placeId = $page.params.placeId ?? "";
    if (tripId && placeId) {
      loadAllData();
    }
  });

  async function loadAllData() {
    if (!tripId || !placeId) return;

    if (!place) isLoading = true;

    try {
      const [placeData, tripData] = await Promise.all([
        apiFetch<Place>(`/trips/${tripId}/places/${placeId}`),
        apiFetch<Trip>(`/trips/${tripId}`),
      ]);

      place = placeData;
      tripBaseCurrency = tripData.base_currency || "EUR";
      tripDefaultCurrency =
        tripData.default_expense_currency || tripData.base_currency || "EUR";
      tripStartDate = tripData.start_date?.split("T")[0] || "";
      tripEndDate = tripData.end_date?.split("T")[0] || "";

      const [expensesData, summaryData, catsData, actsData] = await Promise.all(
        [
          apiFetch<Expense[]>(`/trips/${tripId}/places/${placeId}/expenses`),
          apiFetch<CategorySummary[]>(
            `/trips/${tripId}/places/${placeId}/expenses/summary?currency=${effectivePlaceCurrency}`,
          ),
          apiFetch<Expense_Category[]>(`/trips/${tripId}/expenses/categories`),
          activityApi.list(tripId),
        ],
      );

      expenses = expensesData;
      categorySummary = summaryData;
      categories = catsData;
      activities = actsData.filter((a) => a.place_id === placeId);
    } catch (e) {
      console.error("Failed to load place data", e);
      goto(`/trips/${tripId}`);
    } finally {
      isLoading = false;
    }
  }

  function calculateDuration(start?: string, end?: string) {
    if (!start || !end) return "";
    const diffDays = Math.ceil(
      (new Date(end).getTime() - new Date(start).getTime()) /
        (1000 * 60 * 60 * 24),
    );
    return diffDays > 0
      ? `${diffDays} ${$t("place.days")}`
      : `1 ${$t("place.day")}`;
  }

  function calculateTotal() {
    return categorySummary.reduce((sum, s) => sum + (s.total || 0), 0);
  }

  const effectivePlaceCurrency = $derived(
    place?.default_expense_currency || tripDefaultCurrency || tripBaseCurrency,
  );

  async function savePlaceCurrency(code?: string) {
    if (!place) return;

    const nextValue = code || undefined;
    const currentValue = place.default_expense_currency || undefined;

    if (nextValue === currentValue) return;

    try {
      const updated = await apiFetch<Place>(
        `/trips/${tripId}/places/${placeId}`,
        {
          method: "PUT",
          body: JSON.stringify({
            default_expense_currency: nextValue ?? null,
          }),
        },
      );
      place = updated;
    } catch (e) {
      console.error("Failed to update place currency", e);
    }
  }

  async function savePlaceInfo() {
    if (!place) return;
    try {
      const payload: any = {
        name: place.name,
      };
      if (place.start_date)
        payload.start_date = new Date(place.start_date).toISOString();
      if (place.end_date)
        payload.end_date = new Date(place.end_date).toISOString();
      payload.notes = place.notes || "";
      payload.city = place.city || "";

      const updated = await apiFetch<Place>(
        `/trips/${tripId}/places/${placeId}`,
        {
          method: "PUT",
          body: JSON.stringify(payload),
        },
      );
      place = updated;
    } catch (e) {
      console.error("Failed to update place", e);
    }
  }
</script>

<svelte:head>
  <title>{place ? `${place.name} | Itinera` : $t("common.loading")}</title>
</svelte:head>

<div class="min-h-screen bg-teren-background pb-20">
  {#if place}
    <DetailHeader
      bind:name={place.name}
      bind:description={place.notes}
      bind:startDate={place.start_date}
      bind:endDate={place.end_date}
      bind:defaultCurrency={place.default_expense_currency}
      bind:city={place.city}
      currencyFallbackLabel={`${$t("common.inherit")} (${tripDefaultCurrency})`}
      allowInheritCurrency={true}
      iconType="place"
      durationLabel={calculateDuration(place.start_date, place.end_date)}
      hideDescription={false}
      onSave={savePlaceInfo}
      onUpdateCurrency={savePlaceCurrency}
      onBack={() => goto(`/trips/${tripId}`)}
    />
  {/if}

  <main 
    class="max-w-3xl mx-auto px-4 py-8 space-y-8"
    data-place-id={placeId}
  >
    {#if isLoading}
      <div class="animate-pulse space-y-6">
        <div
          class="h-48 bg-teren-surface rounded-xl border border-teren-border"
        ></div>
      </div>
    {:else if place}
      <ExpensesSummaryCard
        {tripId}
        {placeId}
        {categories}
        {categorySummary}
        displayCurrency={effectivePlaceCurrency}
        {tripDefaultCurrency}
        effectiveCurrency={effectivePlaceCurrency}
        grandTotalValue={calculateTotal()}
        titleBadge={$t("place.local")}
        isHighlighted={true}
        tripStart={tripStartDate}
        tripEnd={tripEndDate}
        onRefresh={loadAllData}
        onOpenDrawer={() => (isDrawerOpen = true)}
      />


      <UpcomingActivityCard
        {tripId}
        {placeId}
        city={place.city}
        tripStart={tripStartDate}
        tripEnd={tripEndDate}
        activities={activities.filter((a) => a.place_id === placeId)}
        onOpenDrawer={() => (isAgendaOpen = true)}
        onRefresh={loadAllData}
      />

      <PhraseDrawer />

      <!-- ============================================================ -->
      <!-- STATION GUIDE: Rendered fully expanded inside Place detail   -->
      <!-- ============================================================ -->
      {#if matchedGuide || placeRules.length > 0}
        <div class="space-y-4 pt-4 border-t border-dashed border-teren-border/60">
          {#if matchedGuide}
            <h2 class="text-lg font-semibold text-teren-text-main tracking-tight flex items-center gap-2">
              <span>🚉</span> {$t("station_guide.section_title" as any)}
            </h2>
            <StationGuide
              cityName={matchedGuide.id}
              compact={false}
              locale={$locale}
            />
          {/if}
          
          {#if placeRules.length > 0}
            <PlaceContextWidget rules={placeRules} />
          {/if}
        </div>
      {/if}
    {/if}
  </main>

  <ExpenseDrawer
    {tripId}
    {placeId}
    {categories}
    isOpen={isDrawerOpen}
    onClose={() => (isDrawerOpen = false)}
    onRefreshSummary={loadAllData}
  />

  <ActivityDrawer
    isOpen={isAgendaOpen}
    {tripId}
    {placeId}
    city={place?.city}
    tripStart={tripStartDate}
    tripEnd={tripEndDate}
    {activities}
    onRefresh={loadAllData}
    onClose={() => (isAgendaOpen = false)}
  />
</div>
