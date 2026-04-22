<script lang="ts">
  import type { Category, CategorySummary } from '$lib/types';
  
  let { categories = [], summary = [], currency }: { categories: Category[]; summary: CategorySummary[]; currency: string; } = $props();

  const total = $derived((summary || []).reduce((acc, curr) => acc + curr.total, 0));
  const catMap = $derived(new Map((categories || []).map(c => [c.id, c])));
</script>

{#if total > 0}
  <div class="flex flex-wrap gap-2 mb-6">
    {#each summary as item}
      {#if catMap.has(item.category_id)}
        <span class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-teren-primary-subtle text-teren-primary-hover rounded-full text-sm font-bold border border-teren-primary/20">
          {catMap.get(item.category_id)?.slug === 'food' ? '🍔' : 
           catMap.get(item.category_id)?.slug === 'transport' ? '🚆' : 
           catMap.get(item.category_id)?.slug === 'accommodation' ? '🏨' : '📦'}
          {item.total.toFixed(2)} {currency}
        </span>
      {/if}
    {/each}
  </div>
{/if}