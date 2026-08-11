<script lang="ts">
  import type { Snippet } from "svelte";

  interface Props {
    /** Plain-text label. Used when no `children` snippet is provided. */
    label?: string;
    /**
     * Rich content for the label — pass Svelte markup directly.
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
    <!-- Visually-hidden real input keeps accessibility & form
             participation; everything visible is the peer-styled
             overlay below. -->
    <input
      id={inputId}
      type="checkbox"
      bind:checked
      aria-invalid={error ? "true" : undefined}
      aria-describedby={error ? errorId : undefined}
      class="peer sr-only"
    />

    <span
      class="mt-0.5 inline-flex h-4.5 w-4.5 shrink-0 items-center justify-center
                   rounded border-2 transition-colors
                   border-teren-border bg-teren-card
                   peer-hover:border-teren-primary/40
                   peer-focus-visible:ring-2 peer-focus-visible:ring-teren-primary/30
                   peer-checked:bg-teren-primary peer-checked:border-teren-primary
                   peer-checked:[&>svg]:opacity-100 peer-checked:[&>svg]:scale-100
                   peer-disabled:opacity-50 peer-disabled:cursor-not-allowed
                   {error
        ? 'border-error-base peer-hover:border-error-base'
        : ''}"
      aria-hidden="true"
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="12"
        height="12"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="3.5"
        stroke-linecap="round"
        stroke-linejoin="round"
        class="text-white opacity-0 scale-50 transition-all duration-150"
      >
        <polyline points="20 6 9 17 4 12"></polyline>
      </svg>
    </span>

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
