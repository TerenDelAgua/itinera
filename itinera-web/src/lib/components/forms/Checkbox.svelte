<script lang="ts">
  /**
   *
   * Two ways to provide the label:
   *
   *   1. Plain string: `<Checkbox label="I accept the terms" />`
   *      Svelte escapes the string automatically — safe for any
   *      caller-supplied content (user input, server strings, etc.).
   *
   *   2. Rich markup via a snippet: pass children and Svelte renders
   *      them as compiled template, NOT raw HTML. This means the
   *      caller can include inline `<a>` tags and other elements
   *      without ever opening an XSS window.
   */
  import type { Snippet } from "svelte";

  interface Props {
    /** Plain-text label. Used when no `children` snippet is provided. */
    label?: string;
    /**
     * Rich content for the label — pass Svelte markup directly.
     * Example:
     *   <Checkbox bind:checked>
     *     {#snippet children()}I accept the <a href="/legal/terms">terms</a>{/snippet}
     *   </Checkbox>
     */
    children?: Snippet;
    checked: boolean;
    error?: string | null;
    id?: string;
  }

  let {
    label,
    children,
    checked = $bindable(),
    error = null,
    id,
  }: Props = $props();

  const inputId = $derived(
    id ?? `checkbox-${Math.random().toString(36).slice(2, 9)}`,
  );
  const errorId = $derived(`${inputId}-error`);
</script>

<div class="flex flex-col gap-1">
  <label
    for={inputId}
    class="flex items-start gap-2.5 text-sm text-teren-text-main cursor-pointer select-none"
  >
    <input
      id={inputId}
      type="checkbox"
      bind:checked
      aria-invalid={error ? "true" : undefined}
      aria-describedby={error ? errorId : undefined}
      class="mt-0.5 h-4 w-4 rounded border-teren-border text-teren-primary
                   focus:ring-2 focus:ring-teren-primary/30 focus:ring-offset-0
                   checked:bg-teren-primary checked:border-teren-primary
                   transition-colors cursor-pointer"
    />
    <span class="leading-snug">
      {#if children}
        {@render children()}
      {:else}
        {label}
      {/if}
    </span>
  </label>
  {#if error}
    <p
      id={errorId}
      class="text-xs text-error-base font-medium ml-6.5"
      role="alert"
    >
      {error}
    </p>
  {/if}
</div>
