<script lang="ts">
  import { t } from '$lib/i18n/store';
  import { formatCurrency } from '$lib/utils';
  import type { CostEstimate } from '$lib/services/costPredictor';

  let {
    estimate,
    onClick
  }: {
    estimate: CostEstimate;
    onClick: () => void;
  } = $props();

  let phase = $derived(estimate.context.phase);
  let daysUntilStart = $derived(estimate.context.daysUntilStart);

  let formattedActual = $derived(formatCurrency(estimate.total.actual, estimate.total.currency).replace(/\.00$/, ''));
  let formattedEstimated = $derived(formatCurrency(estimate.total.estimated, estimate.total.currency).replace(/\.00$/, ''));

  let overAmount = $derived(Math.max(0, estimate.total.actual - estimate.total.estimated));
  let formattedOver = $derived(formatCurrency(overAmount, estimate.total.currency).replace(/\.00$/, ''));

  let statusColor = $derived({
    on_track: 'var(--color-teren-primary)',
    halfway: '#D97706',
    over_budget: 'var(--color-teren-error-base)'
  }[estimate.status]);

</script>

<button 
  class="budget-badge"
  onclick={onClick}
  aria-label={$t('budget.aria_label')}
>
  {#if phase === 'pre_trip'}
    <span class="status-text" style="color: var(--color-teren-text-main)">
      ⚪ {#if daysUntilStart === 0}{$t('budget.trip_starts_today')}{:else if daysUntilStart === 1}{$t('budget.trip_starts_tomorrow')}{:else}{$t('budget.days_left', { count: String(daysUntilStart) })}{/if}
    </span>
  {:else if phase === 'active'}
    <span class="status-text" style="color: {statusColor}">
      {#if estimate.status === 'on_track'}
        🟢 {$t('budget.on_track')}
      {:else if estimate.status === 'halfway'}
        🟡 {$t('budget.halfway')}
      {:else}
        🔴 +{formattedOver}
      {/if}
    </span>
  {:else if phase === 'completed'}
    <span class="status-text" style="color: {estimate.total.actual <= estimate.total.estimated ? 'var(--color-teren-primary)' : 'var(--color-teren-error-base)'}">
      {#if estimate.total.actual <= estimate.total.estimated}
        🟢 {$t('budget.on_track')}
      {:else}
        🔴 +{formattedOver}
      {/if}
    </span>
  {/if}

  <span class="separator">·</span>
  <span class="amounts" style="color: var(--color-teren-text-main)">
    {formattedActual} / {formattedEstimated}
  </span>
</button>

<style>
  .budget-badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 4px 12px;
    background: var(--color-teren-surface);
    border: 1px solid var(--color-teren-border-subtle);
    border-radius: 9999px;
    font-size: 13px;
    font-weight: 500;
    color: var(--color-teren-text-main);
    cursor: pointer;
    transition: all 200ms ease-out;
    white-space: nowrap;
  }

  .budget-badge:hover {
    border-color: var(--color-teren-primary);
    background: var(--color-teren-primary-subtle);
  }

  .status-text {
    font-weight: 600;
  }

  .separator {
    color: var(--color-teren-border-subtle);
    flex-shrink: 0;
  }

  .amounts {
    font-variant-numeric: tabular-nums;
    flex-shrink: 0;
    font-weight: 600;
  }

  /* Clean up dark mode specific hardcoded colors to rely on DS variables */
  :global(.dark) .budget-badge {
    background: var(--color-teren-surface);
    border-color: var(--color-teren-border-subtle);
    color: var(--color-teren-text-main);
  }

  :global(.dark) .budget-badge:hover {
    border-color: var(--color-teren-primary);
    background: var(--color-teren-dark-primary-subtle);
  }
</style>
