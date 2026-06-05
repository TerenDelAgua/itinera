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
  let underAmount = $derived(Math.max(0, estimate.total.estimated - estimate.total.actual));
  let formattedOver = $derived(formatCurrency(overAmount, estimate.total.currency).replace(/\.00$/, ''));
  let formattedUnder = $derived(formatCurrency(underAmount, estimate.total.currency).replace(/\.00$/, ''));

  let statusColor = $derived(
    phase === 'pre_trip'
      ? '#A8A29E'
      : phase === 'active'
        ? {
            on_track: 'var(--color-teren-primary)',
            halfway: '#D97706',
            over_budget: 'var(--color-teren-error-base)'
          }[estimate.status]
        : estimate.total.actual <= estimate.total.estimated
          ? 'var(--color-teren-primary)'
          : 'var(--color-teren-error-base)'
  );

  let statusText = $derived.by(() => {
    if (phase === 'pre_trip') {
      if (daysUntilStart === 0) return $t('budget.trip_starts_today');
      if (daysUntilStart === 1) return $t('budget.trip_starts_tomorrow');
      return $t('budget.days_left', { count: String(daysUntilStart) });
    } else if (phase === 'active') {
      if (estimate.status === 'on_track') return $t('budget.on_track');
      if (estimate.status === 'halfway') return $t('budget.halfway');
      return $t('budget.over_by', { amount: formattedOver });
    } else {
      if (estimate.total.actual <= estimate.total.estimated) {
        return $t('budget.under_budget', { amount: formattedUnder });
      }
      return $t('budget.over_budget', { amount: formattedOver });
    }
  });
</script>

<button 
  class="budget-badge"
  onclick={onClick}
  aria-label={$t('budget.aria_label')}
>
  <span class="status-dot" style="background: {statusColor}"></span>
  <span class="status-text" style="color: {phase === 'pre_trip' ? 'var(--color-teren-text-main)' : statusColor}">
    {statusText}
  </span>
  <span class="separator">·</span>
  <span class="amounts">
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
    color: var(--color-teren-text-muted);
    cursor: pointer;
    transition: all 200ms ease-out;
    white-space: nowrap;
  }

  .budget-badge:hover {
    border-color: var(--color-teren-primary);
    background: var(--color-teren-primary-subtle);
  }

  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 9999px;
    flex-shrink: 0;
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
    color: var(--color-teren-text-main);
  }

  :global(.dark) .budget-badge {
    background: var(--color-teren-surface);
    border-color: var(--color-teren-border-subtle);
    color: var(--color-teren-text-muted);
  }

  :global(.dark) .budget-badge:hover {
    border-color: var(--color-teren-primary);
    background: var(--color-teren-dark-primary-subtle);
  }
</style>
