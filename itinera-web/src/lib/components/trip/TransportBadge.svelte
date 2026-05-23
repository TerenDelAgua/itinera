<script lang="ts">
  import { calculateTransport, routesData } from "$lib/services/transport";
  import type { Place } from "$lib/types/Place";
  import { slide } from "svelte/transition";
  import { Events } from "$lib/services/tracking";
  import { t } from "$lib/i18n/store";
  import { getCurrencySymbol } from "$lib/utils";
  import { apiFetch } from "$lib/api";

  let {
    places,
    tripDurationDays,
    tripId,
    baseCurrency = "JPY",
  }: { places: Place[]; tripDurationDays: number; tripId: string; baseCurrency?: string } = $props();

  let result = $derived(
    places.length >= 2 ? calculateTransport(places, tripDurationDays) : null,
  );
  let expanded = $state(false);

  let tKey = $derived((key: string, params?: Record<string, string>) => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return $t(key as any, params);
  });

  let hasTracked = false;

  $effect(() => {
    if (result && !hasTracked && tripId && result.routesFound.length > 0) {
      Events.tripViewed(tripId, {
        transport_shown: true,
        transport_routes_count: result.routesFound.length,
        transport_savings_jpy: result.savings,
        transport_recommended_pass: result.recommendedPass,
        transport_routes_missing: result.routesMissing.length,
        transport_algorithm: result.algorithmUsed,
      });
      hasTracked = true;
    }
  });

  // Offline-friendly currency conversion rates (JPY to X)
  const JPY_TO_CURRENCY_FALLBACKS: Record<string, number> = {
    JPY: 1.0,
    EUR: 0.0060, // 1 EUR = ~166 JPY
    USD: 0.0064, // 1 USD = ~156 JPY
    GBP: 0.0050, // 1 GBP = ~200 JPY
    CHF: 0.0058, // 1 CHF = ~172 JPY
    CAD: 0.0087, // 1 CAD = ~115 JPY
    AUD: 0.0096, // 1 AUD = ~104 JPY
    PHP: 0.37,   // 1 JPY = 0.37 PHP
    IDR: 102.5,  // 1 JPY = 102.5 IDR
  };

  let currentRate = $state(1.0);

  $effect(() => {
    const target = baseCurrency || "JPY";
    if (target === "JPY") {
      currentRate = 1.0;
      return;
    }

    // Set offline-friendly fallback immediately
    currentRate = JPY_TO_CURRENCY_FALLBACKS[target] || 0.0060;

    // Fetch dynamic rate from local backend API to bypass browser CORS constraints
    apiFetch<{ rate: number }>(`/trips/${tripId}/rates?from=JPY&to=${target}`)
      .then((data) => {
        if (data && data.rate) {
          currentRate = data.rate;
        }
      })
      .catch((err) => {
        console.warn("Could not fetch JPY exchange rate - using offline fallback:", err);
      });
  });

  function formatAmount(jpyValue: number): string {
    const converted = jpyValue * currentRate;
    const symbol = getCurrencySymbol(baseCurrency);

    if (baseCurrency === "JPY") {
      return `¥${Math.round(converted).toLocaleString()}`;
    }

    return `${converted.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })} ${symbol}`;
  }
</script>

{#if result && result.routesFound.length > 0}
  <div
    class="bg-teren-surface border border-teren-border rounded-xl my-4 overflow-hidden border-l-4"
    class:border-l-green-500={result.savings > 10000}
    class:border-l-teren-primary={result.savings > 0 && result.savings <= 10000}
    class:border-l-teren-text-muted={result.savings <= 0}
  >
    <button
      class="flex items-center gap-3 p-4 w-full bg-transparent border-none cursor-pointer text-left font-inherit hover:bg-teren-surface-hover transition-colors"
      onclick={() => (expanded = !expanded)}
    >
      <span class="text-xl">🚄</span>
      <span class="flex-1 font-semibold text-teren-text">
        {#if result.savings > 10000}
          {$t("transport.savings_estimated")}: +{formatAmount(result.savings)}
        {:else if result.savings > 0}
          {$t("transport.jr_transport")}: {formatAmount(result.totalFare)}
        {:else}
          {$t("transport.no_savings", { total: formatAmount(result.totalFare) })}
        {/if}
      </span>
      <span
        class="transition-transform duration-200 text-teren-text-muted {expanded
          ? 'rotate-180'
          : ''}"
      >
        ▼
      </span>
    </button>

    {#if expanded}
      <div class="px-4 pb-4" transition:slide={{ duration: 200 }}>
        {#each result.routesFound as route}
          <div class="py-3 border-b border-teren-border last:border-b-0">
            <div class="font-semibold text-teren-text mb-1">
              {route.from} → {route.to}
            </div>
            <div class="flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-teren-text-muted">
              <span class="font-medium bg-teren-background px-2 py-0.5 rounded text-xs">{route.line}</span>
              <span>{route.timeMin} min</span>
              <span class="font-bold text-teren-text ml-auto tabular-nums">
                {formatAmount(route.fare)}
              </span>
            </div>
            {#if route.noteKey || route.notes}
              <div class="text-[11px] text-teren-text-muted/65 mt-1 italic">
                {route.noteKey ? tKey(route.noteKey, route.noteParams) : route.notes}
              </div>
            {/if}
          </div>
        {/each}

        <div class="h-px bg-teren-border my-3"></div>

        <div class="space-y-2 text-sm">
          <div class="flex justify-between items-baseline">
            <span class="text-teren-text-muted">{$t("transport.total_individual_tickets")}</span>
            <span class="font-semibold text-teren-text tabular-nums text-right min-w-[100px]"
              >{formatAmount(result.totalFare)}</span
            >
          </div>
          {#if result.recommendedPass}
            <div class="flex justify-between items-baseline">
              <span class="text-teren-text-muted"
                >{routesData.passes[result.recommendedPass].name}</span
              >
              <span class="font-semibold text-teren-text tabular-nums text-right min-w-[100px]"
                >{formatAmount(result.passPrice)}</span
              >
            </div>
          {/if}
          <div
            class="flex justify-between items-baseline font-bold border-t border-teren-border mt-2 pt-3"
          >
            <span class="text-teren-text pr-4">{tKey(result.honestMessageKey, result.honestMessageParams)}</span>
            <span
              class="ml-4 whitespace-nowrap tabular-nums text-right min-w-[100px] {result.savings > 0
                ? 'text-green-500'
                : 'text-teren-text'}"
            >
              {result.savings > 0 ? "+" : ""}{formatAmount(Math.abs(result.savings))}
            </span>
          </div>
        </div>

        {#if result.algorithmUsed === "sequential-fallback"}
          <div
            class="text-xs text-teren-text-muted mt-3 pt-2 border-t border-dashed border-teren-border"
          >
            {$t("transport.algorithm_note")}
          </div>
        {/if}
      </div>
    {/if}
  </div>
{/if}
