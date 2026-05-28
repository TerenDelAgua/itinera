<script lang="ts">
  import { slide } from 'svelte/transition';
  import { t, locale } from '$lib/i18n/store';
  import { getPhrasesForContext } from '$lib/services/japanContext';
  import PhraseItem from './PhraseItem.svelte';

  let isExpanded = $state(false);
  let activeCategory = $state('restaurant');

  const categories = [
    { id: 'restaurant', labelKey: 'japan_context.ui.category_restaurant' },
    { id: 'transport', labelKey: 'japan_context.ui.category_transport' },
    { id: 'emergency', labelKey: 'japan_context.ui.category_emergency' },
    { id: 'all', labelKey: 'japan_context.ui.category_general' }
  ];

  let currentPhrases = $derived(getPhrasesForContext(activeCategory, $locale));
</script>

<div class="bg-teren-surface border border-teren-border rounded-xl overflow-hidden shadow-sm">
  <button
    class="w-full px-5 py-4 flex justify-between items-center bg-teren-primary-subtle cursor-pointer hover:bg-teren-primary/10 transition-colors focus:outline-none"
    onclick={() => isExpanded = !isExpanded}
  >
    <h2 class="text-lg font-semibold text-teren-primary-hover flex items-center gap-2">
      {$t('japan_context.ui.useful_phrases' as any)}
    </h2>
    <svg
      class="w-5 h-5 text-teren-primary transition-transform duration-300 {isExpanded ? 'rotate-180' : ''}"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
    >
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
    </svg>
  </button>

  {#if isExpanded}
    <div transition:slide={{ duration: 300 }}>
      <!-- Tabs -->
      <div class="flex overflow-x-auto border-b border-teren-border no-scrollbar px-2 bg-teren-background">
        {#each categories as cat}
          <button
            class="px-4 py-3 text-sm font-semibold whitespace-nowrap transition-colors border-b-2 {activeCategory === cat.id ? 'border-teren-primary text-teren-primary' : 'border-transparent text-teren-text-muted hover:text-teren-text-main'}"
            onclick={() => activeCategory = cat.id}
          >
            {$t(cat.labelKey as any)}
          </button>
        {/each}
      </div>

      <!-- Phrase List -->
      <div class="p-4 flex flex-col gap-3 bg-teren-background/30 max-h-96 overflow-y-auto">
        {#each currentPhrases as phrase (phrase.id)}
          <PhraseItem {phrase} />
        {/each}
        {#if currentPhrases.length === 0}
          <div class="text-center py-6 text-teren-text-muted text-sm">
            No phrases found for this category.
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
  /* Hide scrollbar for Chrome, Safari and Opera */
  .no-scrollbar::-webkit-scrollbar {
    display: none;
  }
  /* Hide scrollbar for IE, Edge and Firefox */
  .no-scrollbar {
    -ms-overflow-style: none;  /* IE and Edge */
    scrollbar-width: none;  /* Firefox */
  }
</style>
