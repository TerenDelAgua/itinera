<script lang="ts">
  import type { ClimateDisplay } from '$lib/services/climateService';
  import { t } from '$lib/i18n/store';

  let { climate, isLarge = false } = $props<{ climate: ClimateDisplay | null, isLarge?: boolean }>();

  // Map our icon names to emojis
  function getEmoji(icon: string): string {
    const map: Record<string, string> = {
      'clear': '☀️',
      'partly-cloudy': '🌤️',
      'rainy': '🌧️',
      'snow': '❄️',
      'fog': '🌫️',
      'storm': '⛈️'
    };
    return map[icon] || '🌤️';
  }
</script>

{#if climate}
  <div class="inline-flex items-center gap-1.5 {isLarge ? 'text-sm' : 'text-xs'} font-medium text-teren-text-muted bg-teren-background/50 backdrop-blur-sm px-2 py-1 rounded-md border border-teren-border/50 shadow-sm" title={climate.notes}>
    <span class="text-teren-primary" aria-hidden="true">{getEmoji(climate.icon)}</span>
    <div class="flex items-center gap-1 tabular-nums">
      {#if climate.temp_current !== undefined}
        <span class="font-bold text-teren-text-main">{climate.temp_current}°C</span>
        <span class="text-[10px] opacity-60">({climate.temp_min}°-{climate.temp_max}°)</span>
      {:else}
        <span>{climate.temp_min}°C – {climate.temp_max}°C</span>
      {/if}
    </div>
  </div>
{/if}
