<script lang="ts">
  import { t } from '$lib/i18n/store';

  let {
    onAddFlight,
    onAddOther
  }: {
    onAddFlight: () => void;
    onAddOther: () => void;
  } = $props();
</script>

<div class="empty-state">
  <div class="empty-icon">✈️</div>
  <h3>{$t('expenses.empty_title')}</h3>
  <p>{$t('expenses.empty_subtitle')}</p>

  <button class="btn-primary" onclick={onAddFlight}>
    + {$t('expenses.add_flight')}
  </button>

  <button class="btn-link" onclick={onAddOther}>
    {$t('expenses.add_other')}
  </button>
</div>

<style>
  .empty-state {
    text-align: center;
    padding: 32px 24px;
  }

  .empty-icon {
    font-size: 48px;
    margin-bottom: 16px;
  }

  .empty-state h3 {
    font-size: 18px;
    margin-bottom: 8px;
    color: var(--color-teren-text-main);
  }

  .empty-state p {
    font-size: 14px;
    color: var(--color-teren-text-muted);
    margin-bottom: 24px;
    line-height: 1.5;
  }

  .btn-primary {
    display: block;
    width: 100%;
    padding: 12px;
    background: var(--color-teren-primary);
    color: white;
    border: none;
    border-radius: 8px;
    font-weight: 600;
    margin-bottom: 12px;
    cursor: pointer;
    transition: all 200ms ease-out;
  }

  .btn-primary:hover {
    background: var(--color-teren-primary-hover);
  }

  .btn-primary:active {
    transform: scale(0.95);
  }

  .btn-link {
    display: block;
    width: 100%;
    padding: 8px;
    background: transparent;
    color: var(--color-teren-text-muted);
    border: none;
    font-size: 14px;
    cursor: pointer;
    text-decoration: underline;
    text-underline-offset: 2px;
  }

  .btn-link:hover {
    color: var(--color-teren-text-main);
  }
</style>
