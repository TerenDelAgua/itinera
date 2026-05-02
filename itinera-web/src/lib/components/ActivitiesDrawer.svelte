<script lang="ts">
  import { fly, fade } from 'svelte/transition';
  import type { Activity } from '$lib/types/Activity';
  import { activityApi } from '$lib/api/activity';
  import { SvelteMap } from 'svelte/reactivity';
  import { t } from '$lib/i18n/store';
  import ActivityQuickAdd from './itinerary/ActivityQuickAdd.svelte';

  let {
    isOpen,
    tripId,
    placeId = undefined,
    activities,
    onRefresh,
    onClose,
    tripStart,
    tripEnd
  }: {
    isOpen: boolean;
    tripId: string;
    placeId?: string;
    activities: Activity[];
    onRefresh: () => void;
    onClose: () => void;
    tripStart?: string;
    tripEnd?: string;
  } = $props();

  // Group activities by date
  let groupedActivities = $derived.by(() => {
    const groups = new SvelteMap<string, Activity[]>();
    
    const filtered = placeId 
      ? activities.filter(a => a.place_id === placeId)
      : activities;

    const sorted = [...filtered].sort((a, b) => {
        if (a.date !== b.date) return a.date.localeCompare(b.date);
        if (a.time && b.time) return a.time.localeCompare(b.time);
        return a.time ? -1 : 1;
    });

    for (const act of sorted) {
      if (!groups.has(act.date)) groups.set(act.date, []);
      groups.get(act.date)!.push(act);
    }
    return groups;
  });

  function formatDateHeader(dateStr: string): string {
    const today = new Date().toISOString().split('T')[0];
    if (dateStr === today) return $t('common.today');
    const d = new Date(dateStr);
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  }
</script>

{#if isOpen}
  <div 
    class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/40 backdrop-blur-sm" 
    transition:fade 
    role="presentation"
    onclick={(e) => e.target === e.currentTarget && onClose()}
    onkeydown={(e) => e.key === 'Escape' && onClose()}
  >
    <div 
      class="bg-teren-surface w-full sm:max-w-lg h-[90vh] sm:h-[80vh] rounded-t-2xl sm:rounded-2xl flex flex-col shadow-2xl overflow-hidden" 
      transition:fly={{ y: 40, duration: 300 }}
      role="presentation"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
    >
      <!-- Header -->
      <div class="p-6 border-b border-teren-border flex justify-between items-center bg-teren-background">
        <h2 class="text-xl font-bold text-teren-text-main">{$t('itinerary.title')}</h2>
        <button onclick={() => onClose()} class="text-teren-text-muted hover:text-teren-text-main p-2 rounded-lg hover:bg-white transition">
          ✕
        </button>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-y-auto p-6 space-y-6">
        
        <!-- List grouped by day -->
        {#each groupedActivities as [date, acts] (date)}
          <div class="space-y-3">
            <div class="sticky top-0 bg-teren-surface/95 backdrop-blur-sm py-2 z-10">
              <h4 class="text-sm font-bold text-teren-text-main border-b border-teren-border pb-1">{formatDateHeader(date)}</h4>
            </div>
            
            {#each acts as activity (activity.id)}
              <div class="group flex items-start gap-4 p-3 rounded-lg hover:bg-teren-background transition cursor-pointer">
                <div class="mt-1 flex flex-col items-center w-10">
                  {#if activity.time}
                    <span class="text-xs font-mono font-bold text-teren-text-muted">{activity.time}</span>
                  {:else}
                    <div class="w-1.5 h-1.5 rounded-full bg-teren-text-muted/30"></div>
                  {/if}
                </div>

                <div class="flex-1">
                  <h5 class="text-sm font-semibold text-teren-text-main">{activity.title}</h5>
                  {#if activity.notes}
                    <p class="text-xs text-teren-text-muted mt-0.5 line-clamp-1 italic">"{activity.notes}"</p>
                  {/if}
                  {#if activity.place_id}
                    <span class="inline-block mt-2 text-[10px] font-bold text-teren-primary bg-teren-primary-subtle px-2 py-0.5 rounded-full">{$t('itinerary.local_activity')}</span>
                  {/if}
                </div>

                <button onclick={() => activityApi.delete(tripId, activity.id).then(onRefresh)} class="opacity-0 group-hover:opacity-100 text-red-400 hover:text-red-600 p-1 transition">
                  🗑️
                </button>
              </div>
            {/each}
          </div>
        {/each}
        
        {#if activities.length === 0}
          <div class="text-center py-10 text-teren-text-muted text-sm">{$t('itinerary.empty_state')}</div>
        {/if}
      </div>

      <!-- Footer: Quick Add (Mirroring Itinerary Style) -->
      <div class="px-6 py-4 border-t border-teren-border bg-white shrink-0">
        <ActivityQuickAdd
          {tripId}
          {tripStart}
          {tripEnd}
          defaultDate={activities.length > 0 ? activities[0].date : undefined}
          onSuccess={() => {
            onRefresh();
          }}
        />
      </div>
    </div>
  </div>
{/if}