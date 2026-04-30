<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { apiFetch } from "$lib/api";
  import { fly, slide } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import { untrack } from "svelte";
  import { t, locale } from "$lib/i18n/store";
  import { formatDisplayDate, formatDate } from "$lib/utils/date";
  import UpcomingActivityCard from "$lib/components/upcomingActivityCard.svelte";
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
  import type { Activity } from "$lib/types/Activity";

  let tripId = $state("");
  let placeId = $state("");

  let place = $state<Place | null>(null);
  let expenses = $state<Expense[]>([]);
  let categorySummary = $state<CategorySummary[]>([]);
  let categories = $state<Expense_Category[]>([]);
  let tripBaseCurrency = $state("EUR");
  let tripDefaultCurrency = $state("EUR");

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
      const [
        placeData,
        expensesData,
        summaryData,
        catsData,
        tripData,
        actsData,
      ] = await Promise.all([
        apiFetch<Place>(`/trips/${tripId}/places/${placeId}`),
        apiFetch<Expense[]>(`/trips/${tripId}/places/${placeId}/expenses`),
        apiFetch<CategorySummary[]>(
          `/trips/${tripId}/places/${placeId}/expenses/summary`,
        ),
        apiFetch<Expense_Category[]>(`/trips/${tripId}/expenses/categories`),
        apiFetch<Trip>(`/trips/${tripId}`),
        activityApi.list(tripId),
      ]);

      place = placeData;
      expenses = expensesData;
      categorySummary = summaryData;
      categories = catsData;
      tripBaseCurrency = tripData.base_currency || "EUR";
      tripDefaultCurrency =
        tripData.default_expense_currency || tripData.base_currency || "EUR";
      activities = actsData;
    } catch (e) {
      console.error("Failed to load place data", e);
      goto(`/trips/${tripId}`);
    } finally {
      isLoading = false;
    }
  }

  function formatSmartDate(dateStr?: string) {
    if (!dateStr) return $t("place.no_date");
    return formatDisplayDate(dateStr, $t, $locale);
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
    return expenses.reduce((sum, exp) => sum + (exp.amount || 0), 0);
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
</script>

<svelte:head>
  <title>{place ? `${place.name} | Itinera` : $t("common.loading")}</title>
</svelte:head>

<div class="min-h-screen bg-teren-background pb-20">
  <header
    class="sticky top-0 z-40 bg-teren-background/90 backdrop-blur-md border-b border-teren-border"
  >
    <div class="max-w-3xl mx-auto px-4 py-3 flex items-start justify-between">
      <div class="flex items-start gap-3">
        <button
          title="Back"
          aria-label={$t("detail.back")}
          onclick={() => goto(`/trips/${tripId}`)}
          class="p-2 -ml-2 text-teren-text-muted hover:text-teren-text-main hover:bg-gray-100 rounded-lg transition active:scale-95"
        >
          <svg
            class="w-5 h-5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M10 19l-7-7m0 0l7-7m-7 7h18"
            />
          </svg>
        </button>

        <div class="min-w-0">
          <h1
            class="text-xl font-bold text-teren-text-main leading-tight flex items-center gap-2"
          >
            <svg
              class="w-5 h-5 text-teren-primary"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"
              />
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"
              />
            </svg>
            {place?.name || $t("common.loading")}
          </h1>

          <div
            class="flex flex-wrap items-center gap-2 text-xs text-teren-text-muted font-medium mt-1"
          >
            <span
              >{formatDate(place?.start_date, $locale)} - {formatDate(
                place?.end_date,
                $locale,
              )} - {calculateDuration(place?.start_date, place?.end_date)}</span
            >
            <CurrencySelector
              value={place?.default_expense_currency || ""}
              onchange={savePlaceCurrency}
              allowInherit={true}
              inheritLabel={`${$t("common.inherit")} (${tripDefaultCurrency})`}
              label={$t("detail.expense_currency_short")}
              compact={true}
              widthClass="w-auto"
            />
          </div>
        </div>
      </div>
    </div>
  </header>

  <main class="max-w-3xl mx-auto px-4 py-8 space-y-8">
    {#if isLoading}
      <div class="animate-pulse space-y-6">
        <div
          class="h-48 bg-teren-surface rounded-xl border border-teren-border"
        ></div>
      </div>
    {:else if place}
      <section
        in:fly={{ y: 20, duration: 400, delay: 50 }}
        class="bg-teren-surface rounded-xl border border-teren-border shadow-sm border-l-4 border-l-teren-primary relative overflow-hidden"
      >
        <div
          class="p-6 border-b border-teren-border flex flex-col sm:flex-row sm:justify-between sm:items-center bg-gradient-to-r from-teren-surface to-teren-primary-subtle/20 gap-3 sm:gap-0"
        >
          <div class="flex justify-between items-center">
            <h2
              class="text-lg font-bold text-teren-text-main flex items-center gap-2"
            >
              {$t("detail.expenses")}
              <span
                class="text-xs font-bold text-teren-primary bg-white px-2 py-0.5 rounded-full border border-teren-primary/20 shadow-sm"
              >
                {$t("place.local")}
              </span>
            </h2>
            <button
              onclick={() => (isMobileExpenseOpen = !isMobileExpenseOpen)}
              class="sm:hidden text-sm font-medium text-teren-primary hover:text-teren-primary-hover transition px-3 py-1.5 rounded-lg bg-teren-primary-subtle active:scale-95 flex-shrink-0"
            >
              + {$t("common.add")}
            </button>
          </div>
          <div class="flex items-center">
            <span
              class="text-3xl sm:text-2xl font-bold text-teren-primary tabular-nums leading-none"
            >
              {calculateTotal().toFixed(2)}
              {tripBaseCurrency}
            </span>
          </div>
        </div>

        <div class="px-6 pt-4">
          {#if categorySummary.length > 0}
            <ExpenseSummaryPills
              {categories}
              summary={categorySummary}
              currency={tripBaseCurrency}
            />
          {:else}
            <div class="pb-2 text-center">
              <p class="text-sm text-teren-text-muted italic">
                {$t("place.no_expenses")}
              </p>
            </div>
          {/if}
        </div>

        <div class="hidden sm:block p-6 pb-2">
          <ExpenseQuickAdd
            {tripId}
            {placeId}
            {categories}
            baseCurrency={tripBaseCurrency}
            defaultCurrency={effectivePlaceCurrency}
            onSuccess={loadAllData}
          />
        </div>

        {#if isMobileExpenseOpen}
          <div
            class="sm:hidden p-6 pb-2 pt-0"
            transition:slide={{ duration: 250, easing: cubicOut }}
          >
            <ExpenseQuickAdd
              {tripId}
              {placeId}
              {categories}
              baseCurrency={tripBaseCurrency}
              defaultCurrency={effectivePlaceCurrency}
              onSuccess={() => {
                loadAllData();
                isMobileExpenseOpen = false;
              }}
            />
          </div>
        {/if}

        <button
          onclick={() => (isDrawerOpen = true)}
          class="w-full py-4 px-6 text-left text-sm font-medium text-teren-text-muted hover:bg-teren-primary-subtle hover:text-teren-primary transition border-t border-teren-border mt-2 flex items-center gap-1 group"
        >
          {$t("place.view_history")}
          <svg
            class="w-4 h-4 group-hover:translate-x-1 transition-transform"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M17 8l4 4m0 0l-4 4m4-4H3"
            />
          </svg>
        </button>
      </section>

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
    {activities}
    onRefresh={loadAllData}
    onClose={() => (isAgendaOpen = false)}
  />
</div>
