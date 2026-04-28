<script lang="ts">
  import type { Activity } from '$lib/types/Activity';
  import { fly, slide } from 'svelte/transition';
  import { t, locale } from '$lib/i18n/store';
  import ActivityQuickAdd from './ActivityQuickAdd.svelte';

  let { tripId, placeId, activities, onOpenDrawer, onRefresh }: { 
    tripId: string;
    placeId?: string;
    activities: Activity[]; 
    onOpenDrawer: () => void;
    onRefresh: () => void;
  } = $props();

  let isAdding = $state(false);

  let upcomingActivity = $derived.by(() => {
    const today = new Date().toISOString().split('T')[0];
    return activities
      .filter(a => a.date >= today)
      .sort((a, b) => {
        if (a.date !== b.date) return a.date.localeCompare(b.date);
        if (a.time && b.time) return a.time.localeCompare(b.time);
        return a.time ? -1 : 1; 
      })[0];
  });

  function getBadgeText(dateStr: string) {
    const today = new Date();
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    
    const targetYMD = dateStr.includes('T') ? dateStr.split('T')[0] : dateStr;
    const formatYMD = (d: Date) => {
        const offset = d.getTimezoneOffset();
        d = new Date(d.getTime() - (offset*60*1000));
        return d.toISOString().split('T')[0];
    };

    if (targetYMD === formatYMD(today)) return $t('common.today_short').toUpperCase();
    if (targetYMD === formatYMD(tomorrow)) return $t('common.tomorrow_short').toUpperCase();
    
    const [year, month, day] = targetYMD.split("-").map(Number);
    const date = new Date(year, month - 1, day);
    return new Intl.DateTimeFormat($locale, { day: 'numeric', month: 'short' }).format(date).toUpperCase();
  }
</script>

<div class="bg-teren-surface p-5 rounded-xl border border-teren-border shadow-sm flex flex-col gap-4" transition:fly={{ y: -10, duration: 300 }}>
  
  <div class="flex justify-between items-center border-b border-teren-border pb-3">
    <h3 class="text-lg font-bold text-teren-text-main">{$t('itinerary.agenda')}</h3>
    <button 
      onclick={() => isAdding = !isAdding}
      class="text-sm font-medium text-teren-primary hover:text-teren-primary-hover transition px-3 py-1.5 rounded-lg bg-teren-primary-subtle active:scale-95"
    >
      + {$t('common.add')}
    </button>
  </div>

  {#if isAdding}
    <ActivityQuickAdd 
      {tripId} 
      {placeId}
      onSuccess={() => { onRefresh(); isAdding = false; }} 
    />
  {/if}

  {#if upcomingActivity}
    <div class="flex items-start gap-4">
      <!-- Indicador de Hora/Estado -->
      <div class="flex flex-col items-center justify-center h-12 w-14 bg-teren-primary-subtle rounded-lg border border-teren-primary/20 text-teren-primary shrink-0">
        <span class="text-[10px] font-bold opacity-70">NEXT</span>
        <span class="text-xs font-bold leading-tight text-center">{getBadgeText(upcomingActivity.date)}</span>
      </div>
      
      <div class="flex-1">
        <div class="flex items-center gap-2 mb-1">
          <span class="text-sm font-bold text-teren-text-main">{upcomingActivity.title}</span>
          {#if upcomingActivity.time}
            <span class="text-xs px-2 py-0.5 rounded-full bg-teren-background border border-teren-border text-teren-text-muted">
              {upcomingActivity.time}
            </span>
          {/if}
        </div>
        {#if upcomingActivity.notes}
          <p class="text-xs text-teren-text-muted line-clamp-1 italic">"{upcomingActivity.notes}"</p>
        {/if}
      </div>
    </div>
  {:else}
    <div class="text-center py-4">
      <p class="text-sm text-teren-text-muted opacity-60">{$t('itinerary.no_activities')}</p>
      </div>
  {/if}

  <!-- Navegación Footer -->
  <button 
    onclick={onOpenDrawer}
    class="w-full pt-2 text-left text-sm font-medium text-teren-text-muted hover:text-teren-primary transition flex items-center gap-1 group"
  >
    {$t('itinerary.view_all')} 
    <svg class="w-4 h-4 group-hover:translate-x-1 transition-transform" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 8l4 4m0 0l-4 4m4-4H3"/>
    </svg>
  </button>
</div>