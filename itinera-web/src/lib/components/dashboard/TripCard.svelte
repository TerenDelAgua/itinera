<script lang="ts">
  import { t, locale } from "$lib/i18n/store";
  import { formatDisplayDate } from "$lib/utils/date";
  import type { Trip } from "$lib/types/Trip";
  import { resolve } from "$app/paths";

  let {
    trip,
    onDeleteClick,
  }: {
    trip: Trip;
    onDeleteClick: (trip: Trip, event: Event) => void;
  } = $props();

  function formatDate(dateStr: string) {
    return formatDisplayDate(
      dateStr,
      $t("common.today_short"),
      $t("common.tomorrow_short"),
      $locale,
    );
  }
</script>

<article
  class="group relative bg-teren-card rounded-xl border border-teren-border transition-all duration-300 hover:-translate-y-1 overflow-hidden focus-within:ring-2 focus-within:ring-teren-primary focus-within:ring-offset-2
         {trip.is_public_demo
    ? 'hover:border-teren-primary/30 shadow-sm hover:shadow-lg hover:shadow-teren-primary/10'
    : 'hover:border-teren-primary/30 shadow-sm hover:shadow-lg hover:shadow-teren-primary/5'}"
>
  <div class="p-5 flex flex-col h-full">
    <div class="flex items-start justify-between gap-4 mb-1">
      <h2
        class="text-xl font-bold text-teren-text-main line-clamp-1 leading-tight"
      >
        <a
          href={resolve(`/trips/${trip.id}`)}
          class="before:absolute before:inset-0 before:z-10 focus:outline-none"
        >
          {trip.name?.startsWith('inspiration.') ? $t(trip.name as any) : trip.name}
        </a>
      </h2>

      {#if trip.is_public_demo}
        <span
          class="px-2 py-0.5 rounded text-[10px] font-bold tracking-wider bg-orange-50 text-orange-500 border border-orange-100 uppercase"
        >
          {$t("dashboard.inspiration_badge")}
        </span>
      {/if}
    </div>

    <!-- 2. FECHAS (Contexto temporal, permite wrap) -->
    <div class="text-xs font-medium text-teren-text-muted tabular-nums mb-3">
      {formatDate(trip.start_date)} — {formatDate(trip.end_date)}
    </div>

    <!-- 3. DESCRIPCIÓN (Máx 2 líneas, itálica sutil) -->
    {#if trip.description}
      <p
        class="text-sm text-teren-text-muted italic line-clamp-2 leading-relaxed flex-grow"
      >
        {trip.description?.startsWith('inspiration.') ? $t(trip.description as any) : trip.description}
      </p>
    {:else}
      <div class="flex-grow"></div>
      <!-- Espaciador para alinear footers -->
    {/if}

    <!-- 4. FOOTER: Destinations (Left) + Price/Delete (Right) -->
    <div class="flex items-center mt-4 pt-3 border-t border-teren-border/50">
      <!-- Destinos (Icono + Número) - Left aligned -->
      {#if trip.place_count > 0}
        <div class="flex items-center gap-1.5 text-teren-text-muted text-xs">
          <svg
            class="w-3.5 h-3.5 opacity-60"
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
          </svg>
          <span class="tabular-nums font-medium">
            {trip.place_count}
            {$t(
              trip.place_count === 1
                ? "dashboard.destination"
                : "dashboard.destinations",
            )}
          </span>
        </div>
      {/if}

      <!-- Grouping Price and Delete to the right -->
      <div class="flex items-center gap-2 ml-auto relative z-20">
        <!-- Badge de Precio -->
        <span
          class="px-2.5 py-1 rounded-full bg-teren-primary-subtle text-teren-primary text-xs font-bold whitespace-nowrap tabular-nums"
        >
          € {trip.total_spent?.toFixed(2) || "0.00"}
        </span>

        <!-- Botón Borrar (Siempre visible en móvil, Hover en Desktop) - Oculto en Demos -->
        {#if !trip.is_public_demo}
          <button
            onclick={(e) => onDeleteClick(trip, e)}
            class="p-1.5 rounded-md text-teren-text-muted hover:text-error-base hover:bg-error-subtle transition-all flex md:flex md:opacity-0 md:group-hover:opacity-100 md:focus:opacity-100"
            aria-label="Eliminar viaje"
          >
            <svg
              class="w-4 h-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
              />
            </svg>
          </button>
        {/if}
      </div>
    </div>
  </div>
</article>
