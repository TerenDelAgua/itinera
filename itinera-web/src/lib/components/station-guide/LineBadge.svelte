<!-- LineBadge.svelte -->
<script lang="ts">
  import type { Line } from '$lib/services/stationGuide';

  let { line }: { line: Line } = $props();
</script>

<div class="flex items-start gap-2.5 py-2 px-2.5 bg-teren-surface rounded-lg border border-teren-border/50 shadow-sm">
  <span class="w-3 h-3 rounded-full flex-shrink-0 mt-0.5 shadow-sm" style="background-color: {line.color};"></span>
  <div class="flex-1 min-w-0">
    <div class="flex items-center gap-1.5 flex-wrap">
      <span class="text-xs font-bold text-teren-text-main leading-tight">{line.name}</span>
      {#if line.name_ja}
        <span class="text-[10px] text-teren-text-muted">{line.name_ja}</span>
      {/if}
    </div>
    {#if (line.platforms && line.platforms.length > 0) || (line.directions && line.directions.length > 0)}
      <div class="flex flex-wrap items-center gap-x-2 gap-y-0.5 text-[10px] text-teren-text-muted mt-0.5 font-medium">
        {#if line.platforms && line.platforms.length > 0}
          <span>Plat: <span class="font-bold text-teren-text-main">{line.platforms.join(', ')}</span></span>
        {/if}
        {#if line.directions && line.directions.length > 0}
          <span class="opacity-40">•</span>
          <span class="truncate">Dir: {line.directions.join('/')}</span>
        {/if}
      </div>
    {/if}
  </div>
</div>
