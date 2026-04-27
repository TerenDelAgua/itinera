<script lang="ts">
  import type { Expense_Category as Category, CategorySummary } from '$lib';
  import { getCurrencySymbol, getCategoryEmoji } from '$lib/utils';
  
  let { categories = [], summary = [], currency }: { categories: Category[]; summary: CategorySummary[]; currency: string; } = $props();

  const total = $derived((summary || []).reduce((acc, curr) => acc + curr.total, 0));
</script>

{#if total > 0}
  <div class="flex flex-wrap gap-2 mb-6">
    {#each summary as item (item.category_id)}
      {@const cat = categories.find(c => c.id === item.category_id || c.slug === item.category_id)}
      <span class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-teren-primary-subtle text-teren-primary-hover rounded-full text-sm font-bold border border-teren-primary/20">
        {getCategoryEmoji(cat?.slug || item.category_id)}
        {item.total.toFixed(2)} {getCurrencySymbol(currency)}
      </span>
    {/each}
  </div>
{/if}