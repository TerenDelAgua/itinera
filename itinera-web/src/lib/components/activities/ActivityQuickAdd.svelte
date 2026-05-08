<script lang="ts">
  import { activityApi } from "$lib/api/activity";
  import type { Activity } from "$lib/types/Activity";
  import { t } from "$lib/i18n/store";
  import { fly } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import { getFriendlyErrorMessage } from "$lib/utils/errorMapper";
  import { locale } from "$lib/i18n/store";
  import { tick } from "svelte";
  import { Events } from "$lib/services/tracking";

  let {
    tripId,
    tripStart,
    tripEnd,
    onSuccess,
    onCancel,
    defaultDate,
    activity,
    placeId,
  } = $props<{
    tripId: string;
    tripStart?: string;
    tripEnd?: string;
    onSuccess: (activity: Activity) => void;
    onCancel?: () => void;
    defaultDate?: string;
    activity?: Activity;
    placeId?: string;
  }>();

  let title = $state(activity?.title || "");
  let date = $state(
    activity?.date || defaultDate || new Date().toISOString().split("T")[0],
  );
  let time = $state(activity?.time || "");
  let notes = $state(activity?.notes || "");
  let isLoading = $state(false);
  let errorMessage = $state<string | null>(null);
  let titleInput: HTMLInputElement;

  async function handleSubmit() {
    if (!title.trim()) return;
    isLoading = true;
    try {
      const data = {
        title,
        date,
        time: time || undefined,
        notes: notes || undefined,
        place_id: placeId || activity?.place_id
      };

      let result: Activity;
      if (activity) {
        result = await activityApi.update(tripId, activity.id, data);
      } else {
        result = await activityApi.create(tripId, data);
        Events.activityCreated(tripId, result.id, result.title);
      }

      if (!activity) {
        title = "";
        time = "";
        notes = "";
        tick().then(() => titleInput?.focus());
      }
      errorMessage = null;
      onSuccess(result);
    } catch (error) {
      const formatOptions: Intl.DateTimeFormatOptions = {
        month: "short",
        day: "numeric",
        timeZone: "UTC",
      };
      const parseYMD = (s: string) => {
        const [y, m, d] = s.split("-").map(Number);
        return new Date(Date.UTC(y, m - 1, d));
      };

      const startFormatted = tripStart
        ? new Intl.DateTimeFormat($locale, formatOptions).format(
            parseYMD(tripStart),
          )
        : "...";
      const endFormatted = tripEnd
        ? new Intl.DateTimeFormat($locale, formatOptions).format(
            parseYMD(tripEnd),
          )
        : "...";

      errorMessage = getFriendlyErrorMessage(error, {
        tripStart: startFormatted,
        tripEnd: endFormatted,
      });
    } finally {
      isLoading = false;
    }
  }
</script>

<div class="flex flex-col gap-2">
  {#if errorMessage}
    <div
      class="bg-error-subtle border border-error-base/20 rounded-xl p-3 flex items-start gap-3 shadow-sm"
      transition:fly={{ y: -5, duration: 200, easing: cubicOut }}
    >
      <div
        class="flex-shrink-0 w-5 h-5 flex items-center justify-center rounded-full bg-error-base/10 text-error-base mt-0.5"
      >
        <svg
          class="w-4 h-4"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
        >
          <circle cx="12" cy="12" r="10" />
          <line x1="12" y1="8" x2="12" y2="12" />
          <line x1="12" y1="16" x2="12.01" y2="16" />
        </svg>
      </div>
      <div class="flex-1 min-w-0">
        <p class="text-sm font-semibold text-error-base leading-tight">
          {errorMessage}
        </p>
      </div>
      <button
        onclick={() => (errorMessage = null)}
        class="text-error-base/40 hover:text-error-base transition-colors p-1 -mr-1"
        aria-label="Dismiss error"
      >
        <svg
          class="w-4 h-4"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
        >
          <path d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
  {/if}

  <div
    class="bg-teren-surface rounded-xl border border-teren-border shadow-sm overflow-hidden transition-all duration-300 focus-within:ring-2 focus-within:ring-teren-primary/30 focus-within:border-teren-primary focus-within:shadow-md"
  >
    <!-- FILA 1: Título -->
    <div class="border-b border-teren-border/50">
      <input
        bind:this={titleInput}
        data-testid="activity-title-input"
        type="text"
        bind:value={title}
        placeholder={$t("itinerary.title_placeholder") || "Activity title"}
        onkeydown={(e) => e.key === "Enter" && handleSubmit()}
        class="w-full h-12 px-4 bg-transparent font-bold text-base text-teren-text-main placeholder:text-teren-text-muted/30 focus:outline-none"
      />
    </div>

    <!-- FILA 2: Fecha y Hora -->
    <div
      class="flex items-stretch divide-x divide-teren-border border-b border-teren-border/50"
    >
      <div class="flex-1 relative flex items-center group min-w-0">
        <span
          class="absolute left-2.5 text-teren-text-muted group-focus-within:text-teren-primary transition-colors pointer-events-none"
        >
          <svg
            class="w-4 h-4"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="1.5"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5"
            />
          </svg>
        </span>
        <input
          type="date"
          bind:value={date}
          onclick={(e) => e.currentTarget.showPicker()}
          onkeydown={(e) => e.key === "Enter" && handleSubmit()}
          class="w-full h-11 pl-8 pr-3 bg-transparent cursor-pointer text-sm text-teren-text-main focus:outline-none min-w-0"
        />
      </div>
      <div class="w-24 relative flex items-center group">
        <span
          class="absolute left-2.5 text-teren-text-muted group-focus-within:text-teren-primary transition-colors pointer-events-none"
        >
          <svg
            class="w-4 h-4"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="1.5"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        </span>
        <input
          type="time"
          bind:value={time}
          onkeydown={(e) => e.key === "Enter" && handleSubmit()}
          class="w-full h-11 pl-8 pr-3 bg-transparent text-sm text-teren-text-main tabular-nums focus:outline-none"
        />
      </div>
    </div>

    <!-- FILA 3: Notas y Botón -->
    <div class="flex items-stretch divide-x divide-teren-border">
      <div class="flex-1 min-w-0">
        <input
          type="text"
          bind:value={notes}
          placeholder={$t("itinerary.notes_placeholder") || "Notes (optional)"}
          onkeydown={(e) => e.key === "Enter" && handleSubmit()}
          class="w-full h-12 px-4 bg-transparent text-sm text-teren-text-muted placeholder:text-teren-text-muted/40 focus:outline-none"
        />
      </div>
      <div class="flex-shrink-0 flex items-center gap-2 px-3">
        {#if onCancel}
          <button
            onclick={onCancel}
            class="px-2 py-1 text-xs font-medium text-teren-text-muted hover:text-teren-text-main transition"
          >
            {$t("common.cancel")}
          </button>
        {/if}
        <button
          onclick={handleSubmit}
          disabled={isLoading || !title.trim()}
          class="w-10 h-10 bg-teren-primary hover:bg-teren-primary-hover text-white rounded-lg shadow-sm active:scale-95 transition-all disabled:opacity-50 flex items-center justify-center"
        >
          {#if isLoading}
            <div
              class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"
            ></div>
          {:else if activity}
            <svg
              class="h-5 w-5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2.5"
                d="M5 13l4 4L19 7"
              />
            </svg>
          {:else}
            <svg
              class="h-5 w-5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2.5"
                d="M12 4v16m8-8H4"
              />
            </svg>
          {/if}
        </button>
      </div>
    </div>
  </div>
</div>
