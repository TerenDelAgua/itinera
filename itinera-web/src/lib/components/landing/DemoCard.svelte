<script lang="ts">
  import { t } from "$lib/i18n/store";

  let { id, name, day_count, destination_count, base_currency } = $props<{
    id: string;
    name: string;
    day_count: number;
    destination_count: number;
    base_currency: string;
  }>();
</script>

<div
  class="bg-teren-card border border-teren-border border-t-white/5 rounded-2xl p-6 transition-all duration-300 hover:border-teren-primary/30 hover:-translate-y-1.5 hover:shadow-xl group flex flex-col h-full"
>
  <div class="mb-4">
    <span
      class="inline-flex items-center px-3 py-1 rounded-full bg-teren-primary-subtle text-teren-primary text-[11px] font-bold tracking-wider uppercase border border-teren-primary/10"
    >
      {$t("dashboard.inspiration_badge")}
    </span>
  </div>

  <h3 class="text-xl font-bold text-teren-text-main mb-4 group-hover:text-teren-primary transition-colors">
    {name}
  </h3>

  <div class="mt-auto space-y-4">
    <div class="flex items-center gap-4 text-sm text-teren-text-muted">
      <div class="flex items-center gap-1.5">
        <svg class="w-4 h-4 opacity-70" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
        <span class="tabular-nums">{day_count} {$t("place.days")}</span>
      </div>
      <div class="flex items-center gap-1.5">
        <svg class="w-4 h-4 opacity-70" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
        </svg>
        <span class="tabular-nums">{destination_count} {$t("dashboard.destinations")}</span>
      </div>
      <div class="flex items-center gap-1.5 font-mono text-xs opacity-60">
         {base_currency}
      </div>
    </div>

    <a
      href="/trips/{id}"
      class="inline-flex items-center gap-2 text-teren-primary font-bold text-sm hover:gap-3 transition-all group/link"
    >
      {$t("landing.customize")}
      <svg class="w-4 h-4 transition-transform group-hover/link:translate-x-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M13 7l5 5m0 0l-5 5m5-5H6" />
      </svg>
    </a>
  </div>
</div>
