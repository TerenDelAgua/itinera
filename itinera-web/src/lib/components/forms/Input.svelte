<script lang="ts">
  /**
   * Text input wrapped with label + help + error states.
   *
   * The input is fully controlled (`bind:value` from the parent).
   * The parent is responsible for trimming / validation; this
   * component only renders state.
   */
  import type { HTMLInputAttributes } from "svelte/elements";

  type $$Props = HTMLInputAttributes & {
    label: string;
    value: string;
    error?: string | null;
    help?: string;
    autocomplete?: string;
  };

  let {
    label,
    value = $bindable(),
    error = null,
    help,
    id,
    type = "text",
    ...rest
  }: $$Props = $props();

  const inputId = $derived(
    id ?? `input-${Math.random().toString(36).slice(2, 9)}`,
  );
  const errorId = $derived(`${inputId}-error`);
  const helpId = $derived(`${inputId}-help`);
</script>

<div class="flex flex-col gap-1.5">
  <label
    for={inputId}
    class="text-xs font-semibold uppercase tracking-wide text-teren-text-muted"
  >
    {label}
  </label>
  <div
    class="flex items-center bg-teren-surface border rounded-lg transition-colors
               {error
      ? 'border-error-base/60 bg-error-subtle/30'
      : 'border-teren-border focus-within:border-teren-primary focus-within:bg-teren-card'}"
  >
    <input
      id={inputId}
      {type}
      bind:value
      aria-invalid={error ? "true" : undefined}
      aria-describedby={error ? errorId : help ? helpId : undefined}
      class="w-full h-11 px-3 bg-transparent text-sm text-teren-text-main placeholder:text-teren-text-muted/40 focus:outline-none"
      {...rest}
    />
  </div>
  {#if error}
    <p id={errorId} class="text-xs text-error-base font-medium" role="alert">
      {error}
    </p>
  {:else if help}
    <p id={helpId} class="text-xs text-teren-text-muted">
      {help}
    </p>
  {/if}
</div>
