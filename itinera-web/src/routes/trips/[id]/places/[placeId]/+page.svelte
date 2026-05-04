<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { apiFetch } from "$lib/api";
  import { fly, slide } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import { untrack } from "svelte";
  import { t, locale } from "$lib/i18n/store";
  import { formatDate } from "$lib/utils/date";
  import ActivitiesDrawer from "$lib/components/ActivitiesDrawer.svelte";
  import { activityApi } from "$lib/api/activity";

  import type { Place } from "$lib/types/Place";
  import type { CategorySummary } from "$lib/types/Summary";
  import type { Expense, Expense_Category } from "$lib/index";
  import type { Trip } from "$lib/types/Trip";

  import ExpenseQuickAdd from "$lib/components/ExpenseQuickAdd.svelte";
  import ExpenseDrawer from "$lib/components/ExpenseDrawer.svelte";
  import ExpenseSummaryPills from "$lib/components/ExpenseSummaryPills.svelte";
  import CurrencySelector from "$lib/components/currency/CurrencySelector.svelte";
  import UpcomingActivityCard from "$lib/components/itinerary/UpcomingActivityCard.svelte";
  import DetailHeader from "$lib/components/trip/DetailHeader.svelte";
  import ExpensesSummaryCard from "$lib/components/trip/ExpensesSummaryCard.svelte";
  import type { Activity } from "$lib/types/Activity";

  let tripId = $state("");
  let placeId = $state("");

  let place = $state<Place | null>(null);
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
  let isMobileExpenseOpen = $state(false);

  $effect(() => {
    if ($page.url.pathname) {
      const parts = $page.url.pathname.split("/");
      tripId = parts[2] || "";
      placeId = parts[4] || "";
      if (tripId && placeId) untrack(() => loadAllData());
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

      const [expensesData, summaryData, catsData, actsData] = await Promise.all([
        apiFetch<Expense[]>(`/trips/${tripId}/places/${placeId}/expenses`),
        apiFetch<CategorySummary[]>(
          `/trips/${tripId}/places/${placeId}/expenses/summary?currency=${effectivePlaceCurrency}`,
        ),
        apiFetch<Expense_Category[]>(`/trips/${tripId}/expenses/categories`),
        activityApi.list(tripId),
      ]);

      expenses = expensesData;
      categorySummary = summaryData;
      categories = catsData;
      activities = actsData;
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
      if (place.start_date) payload.start_date = new Date(place.start_date).toISOString();
      if (place.end_date) payload.end_date = new Date(place.end_date).toISOString();
      payload.notes = place.notes || "";

      const updated = await apiFetch<Place>(`/trips/${tripId}/places/${placeId}`, {
        method: "PUT",
        body: JSON.stringify(payload),
      });
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
      defaultCurrency={place.default_expense_currency || ""}
      currencyFallbackLabel={`${$t("common.inherit")} (${tripDefaultCurrency})`}
      allowInheritCurrency={true}
      iconType="place"
      durationLabel={calculateDuration(place.start_date, place.end_date)}
      hideDescription={true}
      onSave={savePlaceInfo}
      onUpdateCurrency={savePlaceCurrency}
      onBack={() => goto(`/trips/${tripId}`)}
    />
  {/if}

  <main class="max-w-3xl mx-auto px-4 py-8 space-y-8">
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

      {#if place.notes}
        <section
          in:fly={{ y: 20, duration: 400, delay: 100 }}
          class="bg-teren-background p-6 rounded-xl border border-teren-border"
        >
          <h3
            class="text-sm font-semibold text-teren-text-muted uppercase tracking-wider mb-3"
          >
            {$t("detail.description")}
          </h3>
          <p class="text-teren-text-main leading-relaxed whitespace-pre-wrap">
            {place.notes}
          </p>
        </section>
      {/if}

      <UpcomingActivityCard
        {tripId}
        {placeId}
        tripStart={tripStartDate}
        tripEnd={tripEndDate}
        activities={activities.filter((a) => a.place_id === placeId)}
        onOpenDrawer={() => (isAgendaOpen = true)}
        onRefresh={loadAllData}
      />
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

  <ActivitiesDrawer
    isOpen={isAgendaOpen}
    {tripId}
    {placeId}
    tripStart={tripStartDate}
    tripEnd={tripEndDate}
    {activities}
    onRefresh={loadAllData}
    onClose={() => (isAgendaOpen = false)}
  />
</div>
