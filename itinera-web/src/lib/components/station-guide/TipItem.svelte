<!-- TipItem.svelte -->
<script lang="ts">
  import type { Tip } from '$lib/services/stationGuide';
  import { t } from '$lib/i18n/store';

  let { tip, locale }: { tip: Tip; locale: string } = $props();

  let categoryEmoji = $derived.by(() => {
    switch (tip.category) {
      case 'navigation': return '🧭';
      case 'practical': return '💡';
      case 'safety': return '⚠️';
      case 'cultural': return '⛩️';
      default: return '📌';
    }
  });

  let displayCategory = $derived($t(`station_guide.tip_category_${tip.category}` as any));
  let displayPriority = $derived($t(`station_guide.priority_${tip.priority}` as any));

  // Supporting robust Spanish and Japanese translations with default English fallbacks
  let displayContent = $derived.by(() => {
    if (locale === 'ja' && tip.text_ja) return tip.text_ja;
    if (locale === 'es' && (tip as any).text_es) return (tip as any).text_es;
    return tip.text;
  });
</script>

<div
  class="flex gap-3.5 p-4 rounded-xl border border-l-4 transition-all duration-200 hover:shadow-sm bg-teren-surface border-teren-border text-teren-text-main"
  class:border-l-red-500={tip.priority === 'critical'}
  class:border-l-orange-500={tip.priority === 'high'}
  class:border-l-amber-500={tip.priority === 'medium'}
  class:border-l-stone-300={tip.priority === 'low'}
>
  <span class="text-xl flex-shrink-0 select-none mt-0.5" role="img" aria-label={tip.category}>
    {categoryEmoji}
  </span>
  <div class="flex-1 min-w-0">
    <div class="flex items-center gap-2 mb-1 flex-wrap">
      <span
        class="text-[10px] font-bold tracking-wider uppercase px-2 py-0.5 rounded-full"
        class:bg-red-50={tip.priority === 'critical'}
        class:text-red-600={tip.priority === 'critical'}
        class:bg-orange-50={tip.priority === 'high'}
        class:text-orange-600={tip.priority === 'high'}
        class:bg-amber-50={tip.priority === 'medium'}
        class:text-amber-700={tip.priority === 'medium'}
        class:bg-stone-100={tip.priority === 'low'}
        class:text-stone-600={tip.priority === 'low'}
      >
        {displayPriority}
      </span>
      <span class="text-[11px] font-semibold text-teren-text-muted opacity-75">
        • {displayCategory}
      </span>
    </div>
    <p class="text-[13px] leading-relaxed text-teren-text-main whitespace-pre-line font-medium">
      {displayContent}
    </p>
  </div>
</div>
