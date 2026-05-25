<!-- ExitCard.svelte -->
<script lang="ts">
  import type { Exit } from '$lib/services/stationGuide';
  import { t } from '$lib/i18n/store';

  let { exit, exitKey, locale }: { exit: Exit; exitKey: string; locale: string } = $props();

  let displayBestFor = $derived.by(() => {
    if (!exit.best_for || exit.best_for.length === 0) return '';
    const translated = exit.best_for.map(bf => {
      const key = `station_guide.best_for_${bf}`;
      const tVal = $t(key as any);
      return tVal === key ? bf : tVal;
    });
    return translated.join(', ');
  });

  let displayTime = $derived(
    $t('station_guide.exit_time_to_surface', { time: exit.time_to_surface })
  );

  // Localized exit name and tips for seamless ES/JA/EN fallbacks
  let exitName = $derived.by(() => {
    if (locale === 'ja' && exit.name_ja) return exit.name_ja;
    if (locale === 'es' && (exit as any).name_es) return (exit as any).name_es;
    return exit.name;
  });

  let exitTips = $derived.by(() => {
    if (locale === 'ja' && (exit as any).tips_ja) return (exit as any).tips_ja;
    if (locale === 'es' && (exit as any).tips_es) return (exit as any).tips_es;
    return exit.tips;
  });
</script>

<div
  class="bg-teren-surface border border-teren-border rounded-xl p-4 flex flex-col gap-2.5 hover:border-teren-primary/30 transition-all duration-200"
>
  <!-- Primera línea: el nombre de la salida (con icono de localización) -->
  <div class="flex items-center gap-1.5">
    <span class="font-bold text-[14px] sm:text-[15px] text-teren-text-main">📍 {exitName}</span>
  </div>

  <!-- Segunda línea: el nombre en japonés y el tiempo para llegar a la superficie -->
  <div class="flex flex-wrap items-center gap-2 text-xs text-teren-text-muted font-medium">
    {#if exit.name_ja}
      <span class="bg-teren-background text-teren-text-main px-2 py-0.5 rounded text-[10px] font-bold border border-teren-border/50">
        {exit.name_ja}
      </span>
      <span class="opacity-40 select-none">•</span>
    {/if}
    <span class="flex items-center gap-1 text-[11px] font-bold bg-teren-background text-teren-text-muted px-2 py-0.5 rounded-full border border-teren-border/30">
      ⏱️ {displayTime}
    </span>
  </div>

  <!-- Tercera línea: Texto ideal para... -->
  {#if exit.best_for && exit.best_for.length > 0}
    <p class="text-[12px] text-teren-primary font-bold uppercase tracking-wide leading-tight">
      {$t('station_guide.exit_best_for', { categories: displayBestFor })}
    </p>
  {/if}

  <!-- Landmarks (Tags) -->
  {#if exit.landmarks && exit.landmarks.length > 0}
    <div class="flex flex-wrap gap-1.5 mt-1">
      {#each exit.landmarks as landmark}
        <span class="text-[10px] font-semibold bg-teren-background text-teren-text-main px-2 py-0.5 rounded border border-teren-border/50 shadow-sm">
          {landmark}
        </span>
      {/each}
    </div>
  {/if}

  <!-- Exit Specific Tips -->
  {#if exitTips && exitTips.length > 0}
    <div class="border-t border-dashed border-teren-border/70 pt-2.5 mt-1.5">
      <ul class="space-y-1.5 text-xs text-teren-text-muted pl-4 list-disc marker:text-teren-primary">
        {#each exitTips as tip}
          <li class="leading-relaxed font-semibold">
            {tip}
          </li>
        {/each}
      </ul>
    </div>
  {/if}
</div>
