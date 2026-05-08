<script lang="ts">
  import { t } from "$lib/i18n/store";
  import { getCurrencySymbol } from "$lib/utils";
  import type { TripExpenseSummary } from "$lib/types/Summary";
  import type { Expense_Category } from "$lib/types/Category";

  let { summary, categories, baseCurrency } = $props<{
    summary: TripExpenseSummary | null;
    categories: Expense_Category[];
    baseCurrency: string;
  }>();

  function getCategoryTotal(catId: string) {
    return (
      summary?.by_category?.find((c: any) => c.category_id === catId)?.total || 0
    );
  }
</script>

<div class="bg-teren-card rounded-2xl border border-teren-border p-6 shadow-sm">
  <div class="flex items-center justify-between mb-6">
    <h3 class="font-bold text-teren-text-main flex items-center gap-2">
      <span class="text-teren-primary">📊</span>
      {$t("detail.expenses")}
    </h3>
    <button class="text-xs font-bold text-teren-primary hover:underline">
      {$t("detail.view_all_expenses")}
    </button>
  </div>

  <div class="space-y-4">
    {#each categories.slice(0, 4) as cat (cat.id)}
      <div class="space-y-1.5">
        <div
          class="flex justify-between text-[11px] font-bold uppercase tracking-wider"
        >
          <span class="text-teren-text-main">{cat.slug}</span>
          <span class="text-teren-text-muted">
            {getCurrencySymbol(baseCurrency)}
            {getCategoryTotal(cat.id).toFixed(0)}
          </span>
        </div>
        <div
          class="h-1.5 w-full bg-teren-background rounded-full overflow-hidden"
        >
          <div
            class="h-full transition-all duration-1000"
            style="width: {summary && summary.grand_total > 0
              ? (getCategoryTotal(cat.id) / summary.grand_total) * 100
              : 0}%; background-color: {cat.color_hex}"
          ></div>
        </div>
      </div>
    {/each}
  </div>
</div>
