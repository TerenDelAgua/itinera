<script lang="ts">
  import { activityApi } from "$lib/api/activity";
  import type { Activity } from "$lib/types/Activity";

  let { tripId, onSuccess, onCancel, defaultDate, activity } = $props<{
    tripId: string;
    onSuccess: (activity: Activity) => void;
    onCancel?: () => void;
    defaultDate?: string;
    activity?: Activity;
  }>();

  let title = $state(activity?.title || "");
  let date = $state(
    activity?.date || defaultDate || new Date().toISOString().split("T")[0],
  );
  let time = $state(activity?.time || "");
  let notes = $state(activity?.notes || "");
  let isLoading = $state(false);

  async function handleSubmit() {
    if (!title.trim()) return;
    isLoading = true;
    try {
      const data = {
        title,
        date,
        time: time || undefined,
        notes: notes || undefined,
      };

      let result: Activity;
      if (activity) {
        result = await activityApi.update(tripId, activity.id, data);
      } else {
        result = await activityApi.create(tripId, data);
      }

      if (!activity) {
        title = "";
        time = "";
        notes = "";
      }
      onSuccess(result);
    } finally {
      isLoading = false;
    }
  }
</script>

<div
  class="bg-teren-background p-4 rounded-xl border border-teren-border space-y-3"
>
  <div>
    <input
      type="text"
      bind:value={title}
      placeholder="Activity title"
      class="w-full px-3 py-2 bg-white border border-teren-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-teren-primary/30"
    />
  </div>
  <div class="flex gap-3">
    <div class="flex-1 relative flex items-center group">
      <span class="absolute left-3 text-teren-text-muted group-focus-within:text-teren-primary transition-colors pointer-events-none">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5" />
        </svg>
      </span>
      <input
        type="date"
        bind:value={date}
        onclick={(e) => e.currentTarget.showPicker()}
        class="w-full pl-9 pr-3 py-2 bg-white cursor-pointer border border-teren-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-teren-primary/30 min-w-0 transition-all"
      />
    </div>
    <div class="w-32 relative flex items-center group">
      <span class="absolute left-3 text-teren-text-muted group-focus-within:text-teren-primary transition-colors pointer-events-none">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      </span>
      <input
        type="time"
        bind:value={time}
        onclick={(e) => e.currentTarget.showPicker()}
        class="w-full pl-9 pr-3 py-2 bg-white cursor-pointer border border-teren-border rounded-lg text-sm tabular-nums focus:outline-none focus:ring-2 focus:ring-teren-primary/30 transition-all"
      />
    </div>
  </div>
  <div class="flex gap-3">
    <input
      type="text"
      bind:value={notes}
      placeholder="Notes (optional)"
      class="flex-1 px-3 py-2 bg-white border border-teren-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-teren-primary/30 min-w-0"
    />
    {#if onCancel}
      <button
        onclick={onCancel}
        class="px-3 py-2 text-sm font-medium text-teren-text-muted hover:text-teren-text-main transition active:scale-95"
      >
        Cancel
      </button>
    {/if}
    <button
      onclick={handleSubmit}
      disabled={isLoading || !title.trim()}
      aria-label={activity ? "Save activity" : "Add activity"}
      class="w-10 h-10 shrink-0 bg-teren-primary hover:bg-teren-primary-hover text-white font-semibold rounded-lg shadow-sm active:scale-95 transition-all disabled:opacity-50 flex items-center justify-center"
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
