<script lang="ts">
  import { t } from "$lib/i18n/store";
  import { formatDate } from "$lib/utils/date";
  import { getCurrencySymbol } from "$lib/utils";
  import { locale } from "$lib/i18n/store";

  let {
    tripStartDate,
    tripEndDate,
    baseCurrency,
    animatedGrandTotalValue,
  } = $props<{
    tripStartDate: string;
    tripEndDate: string;
    baseCurrency: string;
    animatedGrandTotalValue: number;
  }>();
</script>

<div class="space-y-6">
  <!-- Trip Info & Quick Summary -->
  <div class="flex flex-col md:flex-row md:items-end justify-between gap-4">
    <div class="space-y-1">
      <div class="flex items-center gap-2 text-teren-text-muted text-xs font-semibold uppercase tracking-wider">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <rect width="18" height="18" x="3" y="4" rx="2" ry="2"/><line x1="16" x2="16" y1="2" y2="6"/><line x1="8" x2="8" y1="2" y2="6"/><line x1="3" x2="21" y1="10" y2="10"/>
        </svg>
        <span>{formatDate(tripStartDate, $locale)} — {formatDate(tripEndDate, $locale)}</span>
      </div>
    </div>

    <div class="bg-white px-5 py-3 rounded-2xl border border-teren-border shadow-sm flex items-center gap-4 group transition-all hover:border-teren-primary/30">
      <div class="w-10 h-10 rounded-full bg-teren-primary-subtle flex items-center justify-center text-teren-primary">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="12" x2="12" y1="1" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/>
        </svg>
      </div>
      <div>
        <p class="text-[10px] font-bold text-teren-text-muted uppercase tracking-widest mb-0.5">
          {$t("detail.summary_currency_short")}
        </p>
        <p class="text-xl font-black text-teren-text-main flex items-baseline gap-1">
          <span class="text-xs text-teren-text-muted font-bold">{getCurrencySymbol(baseCurrency)}</span>
          {animatedGrandTotalValue.toFixed(2)}
        </p>
      </div>
    </div>
  </div>
</div>
