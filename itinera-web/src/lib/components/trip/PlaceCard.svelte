<script lang="ts">
  import { t } from "$lib/i18n/store";
  import { getCurrencySymbol } from "$lib/utils";
  import { formatDisplayDate } from "$lib/utils/date";
  import { locale } from "$lib/i18n/store";
  import type { Place } from "$lib/types/Place";

  let {
    place,
    baseCurrency,
    onDelete,
    onClick,
  } = $props<{
    place: Place;
    baseCurrency: string;
    onDelete: (id: string) => void;
    onClick: (id: string) => void;
  }>();

  function formatSmartDate(dateStr?: string) {
    return formatDisplayDate(dateStr, $t, $locale);
  }
</script>

<div
  class="bg-white rounded-xl border border-teren-border p-4 hover:border-teren-primary/30 transition-all cursor-pointer group relative overflow-hidden flex flex-col sm:flex-row sm:items-center justify-between gap-4"
  onclick={() => onClick(place.id)}
>
  <div class="flex items-start gap-4">
    <div
      class="w-12 h-12 rounded-xl bg-teren-background flex items-center justify-center text-xl shrink-0 group-hover:scale-110 transition-transform"
    >
      📍
    </div>
    <div class="min-w-0">
      <h3 class="font-bold text-teren-text-main truncate pr-8 sm:pr-0">
        {place.name}
      </h3>
      <div class="flex flex-wrap items-center gap-x-3 gap-y-1 mt-1">
        <span class="text-xs text-teren-text-muted flex items-center gap-1">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="w-3 h-3"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <rect width="18" height="18" x="3" y="4" rx="2" ry="2" /><line
              x1="16"
              x2="16"
              y1="2"
              y2="6"
            /><line x1="8" x2="8" y1="2" y2="6" /><line
              x1="3"
              x2="21"
              y1="10"
              y2="10"
            />
          </svg>
          <span
            >{formatSmartDate(place.start_date)} — {place.end_date
              ? formatSmartDate(place.end_date)
              : $t("place.no_date")}</span
          >
        </span>
      </div>
    </div>
  </div>

  <div class="flex items-center justify-between sm:justify-end gap-4">
    <div class="text-right">
      <p class="text-[10px] font-bold text-teren-text-muted uppercase tracking-widest">
        {$t("place.local")}
      </p>
      <p class="text-lg font-black text-teren-text-main">
        <span class="text-[10px] text-teren-text-muted font-bold mr-0.5"
          >{getCurrencySymbol(baseCurrency)}</span
        >
        {(place.total_expenses || 0).toFixed(2)}
      </p>
    </div>

    <button
      class="p-2 text-teren-text-muted hover:text-error-base hover:bg-error-subtle rounded-lg transition-all opacity-100 sm:opacity-0 sm:group-hover:opacity-100"
      onclick={(e) => {
        e.stopPropagation();
        onDelete(place.id);
      }}
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        class="w-4 h-4"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path d="M3 6h18" /><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" /><path
          d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"
        />
      </svg>
    </button>
  </div>
</div>
