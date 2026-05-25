<!-- StationGuide.svelte -->
<script lang="ts">
  import { fade, slide } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import {
    findStationGuide,
    getTipsByPriority,
    type StationGuide
  } from '$lib/services/stationGuide';
  import { t } from '$lib/i18n/store';
  import StationMap from './StationMap.svelte';
  import ExitCard from './ExitCard.svelte';
  import TipItem from './TipItem.svelte';
  import LineBadge from './LineBadge.svelte';

  // Svelte 5 Props
  let {
    cityName,
    compact = false,
    locale
  }: {
    cityName: string;
    compact?: boolean;
    locale: string;
  } = $props();

  let expanded = $state(false);
  let activeTab = $state<'tips' | 'exits' | 'map'>('tips');

  let guide = $derived(findStationGuide(cityName));
  let tips = $derived(guide ? getTipsByPriority(guide) : []);

  let complexityText = $derived.by(() => {
    if (!guide) return '';
    return $t(`station_guide.complexity_${guide.complexity}` as any);
  });

  // Complexity visual styling tokens aligned to design system 3.12 (WCAG AA/AAA compliant)
  const complexityStyles: Record<StationGuide['complexity'], string> = {
    low: 'bg-[#D1FAE5] text-[#065F46] border-[#065F46]/20',
    medium: 'bg-[#DBEAFE] text-[#1E40AF] border-[#1E40AF]/20',
    high: 'bg-[#FEF3C7] text-[#92400E] border-[#92400E]/20',
    extreme: 'bg-[#FEE2E2] text-[#991B1B] border-[#991B1B]/20'
  };
</script>

{#if guide}
  <div class="w-full flex flex-col gap-2.5">
    {#if compact}
      <!-- Compact toggle button badge -->
      <button
        type="button"
        onclick={(e) => {
          e.preventDefault();
          e.stopPropagation();
          expanded = !expanded;
        }}
        class="flex items-center justify-between w-full px-4 py-3 rounded-xl border font-semibold text-xs transition-all duration-300 select-none cursor-pointer active:scale-[0.99]
               bg-teren-primary-subtle border-teren-primary/30 text-teren-primary hover:bg-teren-primary hover:text-white"
        class:rounded-b-none={expanded}
        aria-expanded={expanded}
      >
        <div class="flex items-center gap-2">
          <span class="text-sm shrink-0">🚉</span>
          <span class="truncate tracking-wide uppercase">
            {guide.name} — {complexityText}
          </span>
        </div>
        <span class="text-[10px] shrink-0 font-bold transition-transform duration-300" class:rotate-180={expanded}>
          ▼
        </span>
      </button>
    {/if}

    {#if expanded || !compact}
      <div
        class="bg-teren-surface border border-teren-border p-5 rounded-2xl flex flex-col gap-4 shadow-sm"
        class:rounded-t-none={compact && expanded}
        class:-mt-2.5={compact && expanded}
        transition:slide={{ duration: 250, easing: cubicOut }}
      >
        <!-- Header details -->
        <div class="flex justify-between items-start gap-4">
          <div class="min-w-0">
            <h3 class="font-black text-lg text-teren-text-main leading-tight flex items-baseline gap-2 flex-wrap">
              <span>{guide.name}</span>
              {#if guide.name_ja_romaji}
                <span class="text-[11px] font-bold text-teren-text-muted/60 tracking-wider font-sans uppercase">
                  ({guide.name_ja_romaji})
                </span>
              {/if}
            </h3>
            <p class="text-xs text-teren-text-muted mt-0.5 font-bold">
              {guide.name_ja} • {guide.city}, {guide.region.toUpperCase()}
            </p>
          </div>

          <!-- Complexity tag -->
          <span
            class="text-[11px] font-bold tracking-[0.05em] uppercase px-[10px] py-[4px] rounded-full border shadow-sm select-none {complexityStyles[guide.complexity]}"
          >
            {guide.complexity}
          </span>
        </div>

        <!-- Collapsible Train Operators Section -->
        {#if guide.operators && guide.operators.length > 0}
          <div class="flex flex-col gap-2 p-3 bg-teren-background/35 rounded-xl border border-teren-border/50">
            <span class="text-[10px] font-black tracking-widest uppercase text-teren-text-muted/70">
              🚉 Transit Operators & Lines
            </span>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 mt-1">
              {#each guide.operators as operator}
                <div class="flex flex-col gap-1.5 p-2 bg-teren-surface/40 rounded-lg border border-teren-border/20">
                  <span class="text-[10px] font-bold text-teren-text-main flex items-center gap-1">
                    🏢 {operator.name}
                    {#if operator.name_ja}
                      <span class="text-[9px] font-medium text-teren-text-muted">({operator.name_ja})</span>
                    {/if}
                  </span>
                  <div class="flex flex-col gap-1">
                    {#each operator.lines as line}
                      <LineBadge {line} />
                    {/each}
                  </div>
                </div>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Tab switches -->
        <div class="flex gap-1 border-b border-teren-border/75 pt-1" role="tablist">
          <button
            type="button"
            role="tab"
            aria-selected={activeTab === 'tips'}
            class="flex-1 text-center min-h-[40px] flex items-center justify-center border-b-2 font-bold text-xs tracking-wider uppercase transition-all duration-200 ease-out focus:outline-none cursor-pointer hover:bg-teren-interactive-hover hover:text-teren-text-main"
            class:border-teren-primary={activeTab === 'tips'}
            class:text-teren-primary={activeTab === 'tips'}
            class:bg-transparent={activeTab === 'tips'}
            class:border-transparent={activeTab !== 'tips'}
            class:text-teren-text-muted={activeTab !== 'tips'}
            onclick={(e) => { e.preventDefault(); e.stopPropagation(); activeTab = 'tips'; }}
          >
            {$t('station_guide.tabs_tips')}
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={activeTab === 'exits'}
            class="flex-1 text-center min-h-[40px] flex items-center justify-center border-b-2 font-bold text-xs tracking-wider uppercase transition-all duration-200 ease-out focus:outline-none cursor-pointer hover:bg-teren-interactive-hover hover:text-teren-text-main"
            class:border-teren-primary={activeTab === 'exits'}
            class:text-teren-primary={activeTab === 'exits'}
            class:bg-transparent={activeTab === 'exits'}
            class:border-transparent={activeTab !== 'exits'}
            class:text-teren-text-muted={activeTab !== 'exits'}
            onclick={(e) => { e.preventDefault(); e.stopPropagation(); activeTab = 'exits'; }}
          >
            {$t('station_guide.tabs_exits')}
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={activeTab === 'map'}
            class="flex-1 text-center min-h-[40px] flex items-center justify-center border-b-2 font-bold text-xs tracking-wider uppercase transition-all duration-200 ease-out focus:outline-none cursor-pointer hover:bg-teren-interactive-hover hover:text-teren-text-main"
            class:border-teren-primary={activeTab === 'map'}
            class:text-teren-primary={activeTab === 'map'}
            class:bg-transparent={activeTab === 'map'}
            class:border-transparent={activeTab !== 'map'}
            class:text-teren-text-muted={activeTab !== 'map'}
            onclick={(e) => { e.preventDefault(); e.stopPropagation(); activeTab = 'map'; }}
          >
            {$t('station_guide.tabs_map')}
          </button>
        </div>

        <!-- Rendered Tab Panels -->
        <div class="mt-2 min-h-[140px] flex flex-col gap-3">
          {#if activeTab === 'tips'}
            <div class="flex flex-col gap-3" role="tabpanel" in:fade={{ duration: 150 }}>
              {#each tips as tip}
                <TipItem {tip} {locale} />
              {/each}
            </div>
          {:else if activeTab === 'exits'}
            <div class="grid grid-cols-1 gap-3" role="tabpanel" in:fade={{ duration: 150 }}>
              {#each Object.entries(guide.exits) as [key, exit]}
                <ExitCard {exit} exitKey={key} {locale} />
              {/each}
            </div>
          {:else if activeTab === 'map'}
            <div role="tabpanel" in:fade={{ duration: 150 }}>
              <StationMap stationId={guide.id} stationName={guide.name} />
            </div>
          {/if}
        </div>
      </div>
    {/if}
  </div>
{/if}
