<script lang="ts">
  import { t, locale } from "$lib/i18n/store";
  import { formatDate } from "$lib/utils/date";
  import CurrencySelector from "$lib/components/currency/CurrencySelector.svelte";

  let {
    name = $bindable(),
    description = $bindable(),
    startDate = $bindable(),
    endDate = $bindable(),
    defaultCurrency,
    currencyFallbackLabel,
    allowInheritCurrency = false,
    iconType = "trip",
    durationLabel,
    hideDescription = false,
    onSave,
    onUpdateCurrency,
    onBack,
  } = $props<{
    name: string;
    description?: string;
    startDate?: string;
    endDate?: string;
    defaultCurrency: string;
    currencyFallbackLabel?: string;
    allowInheritCurrency?: boolean;
    iconType?: "trip" | "place";
    durationLabel?: string;
    hideDescription?: boolean;
    onSave: () => void;
    onUpdateCurrency: (code?: string) => void;
    onBack: () => void;
  }>();
</script>

<header
  class="bg-teren-background border-b border-teren-border py-4"
>
  <div class="max-w-3xl mx-auto px-4 flex items-start justify-between">
    <div class="flex items-start gap-3 w-full">
      <button
        onclick={onBack}
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
        <div class="flex items-center gap-2">
          {#if iconType === "place"}
            <svg
              class="w-5 h-5 text-teren-primary flex-shrink-0"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"
              />
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"
              />
            </svg>
          {/if}
          <input
            type="text"
            bind:value={name}
            onblur={onSave}
            onkeydown={(e) => e.key === "Enter" && e.currentTarget.blur()}
            placeholder={$t("trip_form.name")}
            class="bg-transparent border-none p-0 focus:ring-0 text-xl font-bold text-teren-text-main leading-tight outline-none w-full truncate"
          />
        </div>
        <div
          class="flex flex-wrap items-center gap-2 text-xs text-teren-text-muted font-medium mt-1"
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

          <div class="relative flex items-center cursor-pointer group px-1.5 py-0.5 rounded-md hover:bg-teren-primary-subtle transition-all">
            <span
              class="text-teren-text-main group-hover:text-teren-primary transition"
              >{formatDate(startDate, $locale)}</span
            >
            <input
              type="date"
              bind:value={startDate}
              onchange={onSave}
              onclick={(e) => e.currentTarget.showPicker()}
              class="absolute inset-0 opacity-0 cursor-pointer w-full h-full"
            />
          </div>

          <span class="opacity-70 mx-0.5">—</span>

          <div class="relative flex items-center cursor-pointer group px-1.5 py-0.5 rounded-md hover:bg-teren-primary-subtle transition-all">
            <span
              class="text-teren-text-main group-hover:text-teren-primary transition"
              >{formatDate(endDate, $locale)}</span
            >
            <input
              type="date"
              bind:value={endDate}
              onchange={onSave}
              onclick={(e) => e.currentTarget.showPicker()}
              class="absolute inset-0 opacity-0 cursor-pointer w-full h-full"
            />
          </div>

          {#if durationLabel}
            <span class="opacity-70 mx-1">·</span>
            <span class="text-teren-text-main font-medium">{durationLabel}</span>
          {/if}

          <div class="relative">
            <CurrencySelector
              value={defaultCurrency || ""}
              fallbackLabel={currencyFallbackLabel}
              allowInherit={allowInheritCurrency}
              onSave={onUpdateCurrency}
            />
          </div>
        </div>

        {#if !hideDescription}
          <input
            type="text"
            bind:value={description}
            onblur={onSave}
            onkeydown={(e) => e.key === "Enter" && e.currentTarget.blur()}
            placeholder={$t("detail.description")}
            class="bg-transparent border-none p-0 focus:ring-0 text-sm text-teren-text-muted outline-none w-full truncate mt-0.5 italic"
          />
        {/if}
      </div>
    </div>
  </div>
</header>
