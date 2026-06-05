<script lang="ts">
  import { t } from '$lib/i18n/store';
  import { formatCurrency } from '$lib/utils';
  import type { CostEstimate } from '$lib/services/costPredictor';

  let {
    estimate
  }: {
    estimate: CostEstimate;
  } = $props();

  let hasFlight = $derived(estimate.flightCost && estimate.flightCost > 0);
  let dailyRemaining = $derived(estimate.dailyBudget.remaining);
  let isDailyOver = $derived(dailyRemaining < 0);

  // Color del remaining (basado en dailyBudget)
  let remainingColor = $derived(
    estimate.context.phase === 'pre_trip' 
      ? 'var(--color-teren-text-muted)'
      : isDailyOver
        ? 'var(--color-teren-error-base)'
        : estimate.status === 'halfway'
          ? '#D97706'
          : 'var(--color-teren-primary)'
  );
</script>

<div class="budget-status-card">
  <!-- Total (siempre incluye flight si existe) -->
  <div class="row">
    <span class="label">{$t('budget.actual')}</span>
    <span class="value">{formatCurrency(estimate.dailyBudget.spent + (estimate.flightCost || 0), estimate.total.currency)}</span>
  </div>

  {#if hasFlight && estimate.flightCost}
    <div class="row sub-row">
      <span class="label">✈️ {$t('budget.flight')}</span>
      <span class="value">{formatCurrency(estimate.flightCost, estimate.total.currency)}</span>
    </div>
  {/if}

  <!-- Estimated (siempre = daily + flight) -->
  <div class="row">
    <span class="label">{$t('budget.estimated')}</span>
    <span class="value">{formatCurrency(estimate.dailyBudget.estimated + (estimate.flightCost || 0), estimate.total.currency)}</span>
  </div>

  <div class="disclaimer">
    {#if hasFlight}
      {$t('budget.daily_excl_flights', { amount: formatCurrency(estimate.dailyBudget.estimated, estimate.total.currency) })}
    {:else}
      {$t('budget.excl_flights')}
    {/if}
  </div>

  <div class="divider"></div>

  <!-- Daily Budget breakdown -->
  <div class="row">
    <span class="label">{$t('budget.daily_spent')}</span>
    <span class="value">{formatCurrency(estimate.dailyBudget.spent, estimate.total.currency)}</span>
  </div>

  <div class="row">
    <span class="label">{$t('budget.daily_estimated')}</span>
    <span class="value">{formatCurrency(estimate.dailyBudget.estimated, estimate.total.currency)}</span>
  </div>

  <div class="divider"></div>

  <!-- Remaining (daily) -->
  <div class="row remaining">
    <span class="label">
      {isDailyOver ? $t('budget.daily_over') : $t('budget.daily_remaining')}
    </span>
    <span class="value" style="color: {remainingColor}">
      {formatCurrency(Math.abs(dailyRemaining), estimate.total.currency)}
    </span>
  </div>

  <!-- Contexto temporal (siempre abajo) -->
  {#if estimate.context.phase === 'pre_trip'}
    <div class="context-hint">
      {$t('budget.days_left', { count: String(estimate.context.daysUntilStart) })}
    </div>
  {:else if estimate.context.phase === 'active'}
    <div class="context-hint">
      {$t('budget.day_x_of_y', { day: String(estimate.context.daysElapsed), total: String(estimate.context.totalDays) })}
    </div>
  {:else}
    <div class="context-hint">
      {$t('budget.trip_completed')}
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

  .sub-row {
    padding-left: 20px;
    font-size: 13px;
  }

  .sub-row .label {
    color: var(--color-teren-text-muted);
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
    margin-top: 8px;
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
