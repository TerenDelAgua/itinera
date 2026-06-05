<script lang="ts">
  import { slide } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import { t } from "$lib/i18n/store";
  import { getCurrencySymbol } from "$lib/utils";
  import type { Expense_Category } from "$lib/types/Category";
  import type { CategorySummary } from "$lib/types/Summary";
  import ExpenseSummaryPills from "$lib/components/Expenses/ExpenseSummaryPills.svelte";
  import ExpenseQuickAdd from "$lib/components/Expenses/ExpenseQuickAdd.svelte";
  import ExpenseEmptyState from "$lib/components/Expenses/ExpenseEmptyState.svelte";

  let {
    tripId,
    placeId,
    categories,
    categorySummary,
    displayCurrency,
    tripDefaultCurrency,
    effectiveCurrency,
    grandTotalValue,
    titleBadge,
    isHighlighted = false,
    tripStart,
    tripEnd,
    onRefresh,
    onOpenDrawer,
  } = $props<{
    tripId: string;
    placeId?: string;
    categories: Expense_Category[];
    categorySummary: CategorySummary[];
    displayCurrency: string;
    tripDefaultCurrency: string;
    effectiveCurrency: string;
    grandTotalValue: number;
    titleBadge?: string;
    isHighlighted?: boolean;
    tripStart?: string;
    tripEnd?: string;
    onRefresh: () => void;
    onOpenDrawer: () => void;
  }>();

  let isMobileExpenseOpen = $state(false);
  let showAddForm = $state(false);
  let preselectedCategorySlug = $state<string | undefined>(undefined);
</script>

<section
  class="bg-teren-surface p-6 rounded-xl border border-teren-border shadow-sm {isHighlighted ? 'border-l-4 border-l-teren-primary' : ''}"
>
  {#if grandTotalValue === 0 && !showAddForm}
    <ExpenseEmptyState
      onAddFlight={() => {
        preselectedCategorySlug = 'flight';
        showAddForm = true;
        isMobileExpenseOpen = true;
      }}
      onAddOther={() => {
        preselectedCategorySlug = undefined;
        showAddForm = true;
        isMobileExpenseOpen = true;
      }}
    />
  {:else}
    <div
      class="flex flex-col sm:flex-row sm:justify-between sm:items-center mb-6 gap-3 sm:gap-0"
    >
      <div class="flex justify-between items-center">
        <h2 class="text-lg font-semibold text-teren-text-main tracking-tight flex items-center gap-2">
          {$t("detail.expenses")}
          {#if titleBadge}
            <span class="text-xs font-bold text-teren-primary bg-teren-card px-2 py-0.5 rounded-full border border-teren-primary/20 shadow-sm">
              {titleBadge}
            </span>
          {/if}
        </h2>
        {#if !showAddForm}
          <button
            onclick={() => (isMobileExpenseOpen = !isMobileExpenseOpen)}
            class="sm:hidden text-sm font-medium text-teren-primary hover:text-teren-primary-hover transition px-3 py-1.5 rounded-lg bg-teren-primary-subtle active:scale-95 flex-shrink-0"
          >
            + {$t("expenses.add_expense")}
          </button>
        {/if}
      </div>
      <div class="flex items-center">
        <span
          data-testid="grand-total"
          class="text-3xl sm:text-2xl font-bold text-teren-primary tabular-nums leading-none"
        >
          {grandTotalValue.toFixed(2)}
          {getCurrencySymbol(displayCurrency)}
        </span>
      </div>
    </div>

    {#if grandTotalValue > 0}
      <ExpenseSummaryPills
        {categories}
        summary={categorySummary}
        currency={displayCurrency}
      />
    {/if}

    {#if showAddForm}
      <div class="mt-6">
        <ExpenseQuickAdd
          {tripId}
          {placeId}
          {categories}
          baseCurrency={displayCurrency}
          {tripStart}
          {tripEnd}
          initialCategorySlug={preselectedCategorySlug}
          insertionCurrency={effectiveCurrency}
          onSuccess={(exp) => {
            showAddForm = false;
            onRefresh();
          }}
        />
      </div>
    {:else}
      <div class="hidden sm:block mt-6">
        <ExpenseQuickAdd
          {tripId}
          {placeId}
          {categories}
          baseCurrency={displayCurrency}
          {tripStart}
          {tripEnd}
          insertionCurrency={effectiveCurrency}
          onSuccess={onRefresh}
        />
      </div>
    {/if}

    <!-- Quick Add (Mobile Context) -->
    {#if isMobileExpenseOpen && !showAddForm}
      <div
        class="sm:hidden mt-4"
        transition:slide={{ duration: 250, easing: cubicOut }}
      >
        <ExpenseQuickAdd
          {tripId}
          {placeId}
          {categories}
          baseCurrency={displayCurrency}
          {tripStart}
          {tripEnd}
          insertionCurrency={effectiveCurrency}
          onSuccess={() => {
            onRefresh();
          }}
        />
      </div>
    {/if}

    {#if showAddForm}
      <button
        onclick={() => showAddForm = false}
        class="mt-5 text-sm text-teren-text-muted hover:text-teren-primary transition flex items-center gap-1 group"
      >
        {$t("common.cancel")}
      </button>
    {:else}
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
    {/if}
  {/if}
</section>
