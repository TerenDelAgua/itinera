<script lang="ts">
  import { t } from '$lib/i18n/store';
  import { formatCurrency } from '$lib/utils';
  import type { CostEstimate } from '$lib/services/costPredictor';

  let {
    estimate
  }: {
    estimate: CostEstimate;
  } = $props();

  let remaining = $derived(estimate.total.estimated - estimate.total.actual);
  let isOver = $derived(remaining < 0);

  // Color según fase y estado
  let statusColor = $derived(
    isOver 
      ? 'var(--color-teren-error-base)'
      : estimate.context.phase === 'pre_trip'
        ? 'var(--color-teren-text-muted)'
        : estimate.status === 'halfway'
          ? '#D97706'
          : 'var(--color-teren-primary)'
  );
</script>

<div class="budget-status-card">
  <!-- Actual vs Estimated -->
  <div class="row">
    <span class="label">{$t('budget.actual')}</span>
    <span class="value">{formatCurrency(estimate.total.actual, estimate.total.currency)}</span>
  </div>
  <div class="row">
    <span class="label">{$t('budget.estimated')}</span>
    <span class="value">{formatCurrency(estimate.total.estimated, estimate.total.currency)}</span>
  </div>

  <!-- Disclaimer siempre visible -->
  <div class="disclaimer">{$t('budget.excl_flights')}</div>

  <div class="divider"></div>

  <!-- Remaining / Over -->
  <div class="row remaining">
    <span class="label">
      {isOver ? $t('budget.over') : $t('budget.remaining')}
    </span>
    <span class="value" style="color: {statusColor}">
      {formatCurrency(Math.abs(remaining), estimate.total.currency)}
    </span>
  </div>

  <!-- Contexto temporal (opcional, solo en active) -->
  {#if estimate.context.phase === 'active'}
    <div class="context-hint">
      {$t('budget.day_x_of_y', { day: String(estimate.context.daysElapsedDisplay), total: String(estimate.context.totalDays) })}
    </div>
  {/if}
</div>

<style>
  .budget-status-card {
    background: var(--color-teren-surface);
    border: 1px solid var(--color-teren-border-subtle);
    border-radius: 12px;
    padding: 16px;
    margin-bottom: 16px;
  }

  .row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 4px 0;
  }

  .label {
    font-size: 14px;
    color: var(--color-teren-text-muted);
  }

  .value {
    font-size: 14px;
    font-weight: 600;
    color: var(--color-teren-text-main);
    font-variant-numeric: tabular-nums;
  }

  .disclaimer {
    font-size: 12px;
    color: var(--color-teren-text-muted);
    font-style: italic;
    margin-top: 4px;
    opacity: 0.7;
  }

  .divider {
    height: 1px;
    background: var(--color-teren-border-subtle);
    margin: 8px 0;
  }

  .remaining .value {
    font-size: 16px;
    font-weight: 700;
  }

  .context-hint {
    font-size: 12px;
    color: var(--color-teren-text-muted);
    text-align: right;
    margin-top: 4px;
  }

  :global(.dark) .budget-status-card {
    background: var(--color-teren-surface);
    border-color: var(--color-teren-border-subtle);
  }

  :global(.dark) .label,
  :global(.dark) .disclaimer,
  :global(.dark) .context-hint {
    color: var(--color-teren-text-muted);
  }

  :global(.dark) .value {
    color: var(--color-teren-text-main);
  }
</style>
