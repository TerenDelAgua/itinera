<script lang="ts">
  import { t } from "$lib/i18n/store";

  let { 
    isCreatingTrip = false, 
    onCreateClick 
  }: { 
    isCreatingTrip?: boolean; 
    onCreateClick: () => void; 
  } = $props();
</script>

<section
  class="flex flex-col sm:flex-row sm:justify-between sm:items-center mb-10 gap-4"
>
  <div>
    <h1
      class="text-3xl sm:text-4xl font-bold text-teren-text-main tracking-tight"
    >
      {$t("dashboard.title")}
    </h1>
    <p class="text-teren-text-muted mt-2 text-base">
      {$t("dashboard.subtitle")}
    </p>
  </div>

  <button
    onclick={onCreateClick}
    disabled={isCreatingTrip}
    class="bg-teren-primary hover:bg-teren-primary-hover text-white font-medium py-2.5 px-6 rounded-lg shadow-sm hover:shadow-md active:scale-95 transition-all duration-200 flex items-center gap-2 disabled:opacity-50 w-fit"
  >
    <span class="text-xl leading-none">+</span>
    <span>{$t("dashboard.button_new")}</span>
  </button>
</section>
