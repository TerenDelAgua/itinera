<script lang="ts">
  import { t } from "$lib/i18n/store";

  let { onCreateClick }: { onCreateClick: () => void } = $props();
</script>

<div class="flex flex-col items-center justify-center py-20 text-center">
  <div
    class="w-20 h-20 mb-6 rounded-full bg-white border border-teren-border flex items-center justify-center shadow-sm"
  >
    <svg
      xmlns="http://www.w3.org/2000/svg"
      class="h-10 w-10 text-teren-text-muted opacity-40"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
    >
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="1.5"
        d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0121 18.382V7.618a1 1 0 01-.553-.894L15 7m0 13V7"
      />
    </svg>
  </div>
  <h3 class="text-xl font-medium text-teren-text-main mb-2">
    {$t("empty_state.title")}
  </h3>
  <p class="text-teren-text-muted max-w-xs mb-8">
    {$t("empty_state.description")}
  </p>
  <button
    onclick={onCreateClick}
    class="text-teren-primary font-semibold hover:text-teren-primary-hover transition-colors"
  >
    {$t("empty_state.cta")} &rarr;
  </button>
</div>
