<script lang="ts">
  import { onMount } from 'svelte';
  import { tweened } from 'svelte/motion';
  import { cubicOut } from 'svelte/easing';
  import { t } from '$lib/i18n/store';

  let { count = 0 } = $props<{ count?: number }>();

  const displayCount = tweened(0, {
    duration: 800,
    easing: cubicOut
  });

  onMount(() => {
    // Small delay to ensure the animation is seen after page load
    setTimeout(() => {
      displayCount.set(count);
    }, 200);
  });
</script>

<div class="inline-flex items-center gap-3 px-4 py-2 bg-teren-primary-subtle border border-teren-primary/10 rounded-full shadow-sm">
  <div class="relative flex h-2 w-2">
    <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
    <span class="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
  </div>
  
  <p class="text-sm font-medium text-teren-text-muted">
    <span class="text-teren-text-main font-bold tabular-nums">
      {Math.round($displayCount)}
    </span>
    {$t('landing.travelers_planning')}
  </p>
</div>
