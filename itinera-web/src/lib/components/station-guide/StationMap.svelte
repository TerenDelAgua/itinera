<!-- StationMap.svelte -->
<script lang="ts">
  import { t } from '$lib/i18n/store';

  let { stationId, stationName }: { stationId: string; stationName: string } = $props();
</script>

<div
  class="relative overflow-hidden bg-teren-surface border border-teren-border rounded-2xl p-6 md:p-8 flex flex-col items-center justify-center text-center gap-4 transition-all duration-300"
>
  <!-- Background subtle warm gradient circle to look premium -->
  <div class="absolute -top-24 -left-24 w-48 h-48 rounded-full bg-teren-primary/5 blur-3xl pointer-events-none"></div>
  <div class="absolute -bottom-24 -right-24 w-48 h-48 rounded-full bg-teren-primary/5 blur-3xl pointer-events-none"></div>

  <!-- Sunrise Orange Glassmorphic Icon Wrapper -->
  <div
    class="w-16 h-16 rounded-2xl bg-gradient-to-br from-teren-primary/20 to-teren-primary/5 border border-teren-primary/25 flex items-center justify-center text-3xl shadow-sm shadow-teren-primary/5 animate-pulse"
  >
    🚉
  </div>

  <div class="max-w-md flex flex-col gap-2">
    <h4 class="text-base font-bold text-teren-text-main leading-snug">
      {$t('station_guide.map_placeholder')}
    </h4>
    <p class="text-xs text-teren-text-muted leading-relaxed font-medium">
      {$t('station_guide.map_placeholder_subtitle', { station_name: stationName })}
    </p>
  </div>

  <!-- Premium "Japan Edge" navigational philosophy badge -->
  <div
    class="mt-2 inline-flex items-center gap-1.5 px-3.5 py-2 rounded-xl bg-teren-primary-subtle text-teren-primary border border-teren-primary/10 text-[11px] font-bold tracking-tight max-w-sm"
  >
    💡 <span class="text-left leading-normal font-medium text-teren-primary/90">
      <strong>{$t('station_guide.map_insight_title')}:</strong> {$t('station_guide.map_insight_body')}
    </span>
  </div>
</div>
