<script lang="ts">
  import { goto } from "$app/navigation";
  import { t, locale } from "$lib/i18n/store";
  import { formatDate } from "$lib/utils/date";
  import CurrencySelector from "$lib/components/currency/CurrencySelector.svelte";

  let {
    tripName = $bindable(),
    tripDescription = $bindable(),
    tripStartDate = $bindable(),
    tripEndDate = $bindable(),
    tripDefaultCurrency,
    onSave,
    onUpdateCurrency,
  } = $props<{
    tripName: string;
    tripDescription: string;
    tripStartDate: string;
    tripEndDate: string;
    tripDefaultCurrency: string;
    onSave: () => void;
    onUpdateCurrency: (code?: string) => void;
  }>();
</script>

<header
  class="sticky top-0 z-40 bg-teren-background/90 backdrop-blur-md border-b border-teren-border py-2"
>
  <div class="max-w-3xl mx-auto px-4 flex items-start justify-between">
    <div class="flex items-start gap-3 w-full">
      <button
        onclick={() => goto("/")}
        aria-label={$t("detail.back")}
        class="p-2 -ml-2 mt-0.5 text-teren-text-muted hover:text-teren-text-main hover:bg-gray-100 rounded-lg transition active:scale-95 flex-shrink-0"
      >
        <svg
          class="w-5 h-5"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M10 19l-7-7m0 0l7-7m-7 7h18"
          />
        </svg>
      </button>
      <div class="flex-1 min-w-0 flex flex-col gap-1">
        <div class="flex items-center">
          <input
            type="text"
            bind:value={tripName}
            onblur={onSave}
            onkeydown={(e) => e.key === "Enter" && e.currentTarget.blur()}
            placeholder={$t("trip_form.name")}
            class="bg-transparent border-none p-0 focus:ring-0 text-xl font-bold text-teren-text-main leading-tight outline-none w-full truncate"
          />
        </div>
        <div
          class="flex flex-wrap items-center gap-2 text-xs text-teren-text-muted font-medium"
        >
          <svg
            class="w-3.5 h-3.5 opacity-70 flex-shrink-0"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            ><path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
            /></svg
          >

          <div class="relative flex items-center cursor-pointer group">
            <span
              class="text-teren-text-main group-hover:text-teren-primary transition"
              >{formatDate(tripStartDate, $locale)}</span
            >
            <input
              type="date"
              bind:value={tripStartDate}
              onchange={onSave}
              class="absolute inset-0 opacity-0 cursor-pointer w-full h-full pointer-events-auto"
            />
          </div>

          <span class="opacity-70 mx-0.5">—</span>

          <div class="relative flex items-center cursor-pointer group">
            <span
              class="text-teren-text-main group-hover:text-teren-primary transition"
              >{formatDate(tripEndDate, $locale)}</span
            >
            <input
              type="date"
              bind:value={tripEndDate}
              onchange={onSave}
              class="absolute inset-0 opacity-0 cursor-pointer w-full h-full pointer-events-auto"
            />
          </div>

          <div class="relative">
            <CurrencySelector
              value={tripDefaultCurrency || ""}
              onSave={onUpdateCurrency}
            />
          </div>
        </div>

        <input
          type="text"
          bind:value={tripDescription}
          onblur={onSave}
          onkeydown={(e) => e.key === "Enter" && e.currentTarget.blur()}
          placeholder={$t("detail.description")}
          class="bg-transparent border-none p-0 focus:ring-0 text-sm text-teren-text-muted outline-none w-full truncate mt-0.5 italic"
        />
      </div>
    </div>
  </div>
</header>
