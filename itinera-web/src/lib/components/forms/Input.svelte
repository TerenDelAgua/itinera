<script lang="ts">
  import type { HTMLInputAttributes } from "svelte/elements";
  import type { Snippet } from "svelte";

  type $$Props = HTMLInputAttributes & {
    label: string;
    value: string;
    error?: string | null;
    help?: string;
    autocomplete?: string;
    trailingIcon?: Snippet;
  };

  let {
    label,
    value = $bindable(),
    error = null,
    help,
    id,
    type = "text",
    trailingIcon,
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
    class="flex items-center bg-input border rounded-lg transition-colors
               {error
                   ? 'border-error-base bg-error-subtle/30'
                   : 'border-teren-border hover:border-teren-primary/40 focus-within:border-teren-primary focus-within:bg-input'}"
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
    {#if trailingIcon}
      <!-- 44×44 tap target with the icon centered. The container
                 reserves this slot unconditionally so the layout never
                 shifts, even when no icon is provided. -->
      <div class="shrink-0 w-11 h-11 flex items-center justify-center">
        {@render trailingIcon()}
      </div>
    {/if}
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
