<script lang="ts">
  import { COMMON_CURRENCIES } from '$lib/types/Currency';

  let {
    value,
    onchange,
    allowInherit = false,
    inheritLabel = 'Inherit',
    widthClass = 'w-28',
    label,
    compact = false
  }: {
    value?: string;
    onchange: (code?: string) => void;
    allowInherit?: boolean;
    inheritLabel?: string;
    widthClass?: string;
    label?: string;
    compact?: boolean;
  } = $props();
</script>

{#if compact}
  <div class={`relative inline-flex items-center rounded-full border border-teren-border bg-teren-surface px-3 py-1.5 shadow-sm ${widthClass}`}>
    {#if label}
      <span class="mr-2 text-[11px] font-semibold uppercase tracking-wide text-teren-text-muted">{label}</span>
    {/if}

    <select
      bind:value
      onchange={(e) => {
        const selected = (e.target as HTMLSelectElement).value;
        onchange(selected || undefined);
      }}
      class="appearance-none bg-transparent border-none pl-0 pr-5 py-0 text-xs font-semibold text-teren-text-main focus:outline-none focus:ring-0 cursor-pointer"
    >
      {#if allowInherit}
        <option value="">{inheritLabel}</option>
      {/if}
      {#each COMMON_CURRENCIES as c (c.code)}
        <option value={c.code}>{c.code}</option>
      {/each}
    </select>

    <div class="pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 text-teren-text-muted">
      <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </div>
  </div>
{:else}
  <div class={`relative flex-shrink-0 ${widthClass}`}>
    <select
      bind:value
      onchange={(e) => {
        const selected = (e.target as HTMLSelectElement).value;
        onchange(selected || undefined);
      }}
      class="appearance-none w-full bg-teren-background border border-teren-border rounded-lg px-3 py-2.5 text-sm font-medium text-teren-text-main focus:outline-none focus:ring-2 focus:ring-teren-primary/30 cursor-pointer"
    >
      {#if allowInherit}
        <option value="">{inheritLabel}</option>
      {/if}
      {#each COMMON_CURRENCIES as c (c.code)}
        <option value={c.code}>{c.code} ({c.symbol})</option>
      {/each}
    </select>

    <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-teren-text-muted">
      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </div>
  </div>
{/if}
