<script lang="ts">
  import { COMMON_CURRENCIES } from "$lib/types/Currency";

  let {
    value = $bindable(),
    fallbackLabel = "",
    allowInherit = false,
    onSave,
  }: {
    value: string | undefined;
    fallbackLabel: string;
    allowInherit?: boolean;
    onSave: (code?: string) => void;
  } = $props();

  function handleChange(e: Event) {
    onSave((e.target as HTMLSelectElement).value);
  }
</script>

<!-- Contenedor compacto que se adapta al contenido -->
<div class="relative flex items-center cursor-pointer group px-2 py-0.5 rounded-md bg-teren-primary-subtle hover:bg-teren-primary/15 border border-teren-primary/20 transition-all w-fit">
  <select
    value={value || ""}
    data-testid="currency-selector"
    onchange={handleChange}
    class="bg-transparent border-none p-0 focus:ring-0 text-xs font-bold text-teren-text-main cursor-pointer outline-none w-auto min-w-[2.5rem] text-center appearance-none"
  >
    {#if allowInherit}
      <option value="" class="text-teren-text-muted bg-teren-card">{fallbackLabel}</option>
      <option disabled class="text-teren-border bg-teren-card">──────────</option>
    {/if}

    {#each COMMON_CURRENCIES as c}
      <option value={c.code} class="bg-teren-card text-teren-text-main">{c.code}</option>
    {/each}
  </select>
</div>
