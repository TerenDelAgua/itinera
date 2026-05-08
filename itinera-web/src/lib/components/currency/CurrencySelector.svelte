<script lang="ts">
  import { COMMON_CURRENCIES } from "$lib/types/Currency";

  let {
    value,
    fallbackLabel = "",
    allowInherit = false,
    onSave,
  }: {
    value: string;
    fallbackLabel: string;
    allowInherit?: boolean;
    onSave: (code?: string) => void;
  } = $props();

  function handleChange(e: Event) {
    onSave((e.target as HTMLSelectElement).value);
  }
</script>

<!-- Contenedor compacto que se adapta al contenido -->
<div class="relative group w-fit">
  <select
    {value}
    data-testid="currency-selector"
    onchange={handleChange}
    data-testid="currency-selector"
    class="appearance-none bg-teren-card border border-teren-border hover:border-teren-primary/50 text-teren-text-main font-bold px-2.5 py-0.5 rounded-lg text-xs cursor-pointer focus:outline-none focus:ring-2 focus:ring-teren-primary/30 transition-colors w-auto min-w-[3.5rem] text-center"
  >
    {#if allowInherit}
      <option value="" class="text-teren-text-muted">{fallbackLabel}</option>
      <option disabled class="text-teren-border">──────────</option>
    {/if}

    {#each COMMON_CURRENCIES as c}
      <option value={c.code}>{c.code}</option>
    {/each}
  </select>
</div>
