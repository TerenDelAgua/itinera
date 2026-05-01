<script lang="ts">
  import { slide } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import { t } from "$lib/i18n/store";
  import { getCurrencySymbol } from "$lib/utils";
  import type { Expense_Category } from "$lib/types/Category";
  import type { TripExpenseSummary } from "$lib/types/Summary";
  import ExpenseSummaryPills from "$lib/components/ExpenseSummaryPills.svelte";
  import ExpenseQuickAdd from "$lib/components/ExpenseQuickAdd.svelte";

  let {
    tripId,
    categories,
    summary,
    baseCurrency,
    tripDefaultCurrency,
    effectiveCurrency,
    grandTotalValue,
    onRefresh,
    onOpenDrawer,
  } = $props<{
    tripId: string;
    categories: Expense_Category[];
    summary: TripExpenseSummary | null;
    baseCurrency: string;
    tripDefaultCurrency: string;
    effectiveCurrency: string;
    grandTotalValue: number;
    onRefresh: () => void;
    onOpenDrawer: () => void;
  }>();

  let isMobileExpenseOpen = $state(false);
</script>

<section
  class="bg-teren-surface p-6 rounded-xl border border-teren-border shadow-sm"
>
  <div
    class="flex flex-col sm:flex-row sm:justify-between sm:items-center mb-6 gap-3 sm:gap-0"
  >
    <div class="flex justify-between items-center">
      <h2 class="text-lg font-semibold text-teren-text-main tracking-tight">
        {$t("detail.expenses")}
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
        {grandTotalValue.toFixed(2)}
        {getCurrencySymbol(baseCurrency)}
      </span>
    </div>
  </div>

  <ExpenseSummaryPills
    {categories}
    summary={summary?.by_category || []}
    currency={baseCurrency}
  />

  <div class="hidden sm:block mt-6">
    <ExpenseQuickAdd
      {tripId}
      {categories}
      {baseCurrency}
      insertionCurrency={effectiveCurrency}
      onSuccess={onRefresh}
    />
  </div>

  <!-- Quick Add (Mobile Context) -->
  {#if isMobileExpenseOpen}
    <div
      class="sm:hidden mt-4"
      transition:slide={{ duration: 250, easing: cubicOut }}
    >
      <ExpenseQuickAdd
        {tripId}
        {categories}
        {baseCurrency}
        insertionCurrency={tripDefaultCurrency}
        onSuccess={() => {
          onRefresh();
          isMobileExpenseOpen = false;
        }}
      />
    </div>
  {/if}

  <button
    onclick={onOpenDrawer}
    class="mt-5 text-sm text-teren-text-muted hover:text-teren-primary transition flex items-center gap-1 group"
  >
    {$t("detail.view_all_expenses")}
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
