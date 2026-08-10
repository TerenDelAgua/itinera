<script lang="ts">
  /**
   * Disabled state comes from BOTH the parent (`disabled` prop) and
   * the loading state — so a button mid-submission can't be tapped
   * again accidentally.
   */
  import type { HTMLButtonAttributes } from "svelte/elements";

  type $$Props = HTMLButtonAttributes & {
    loading?: boolean;
    fullWidth?: boolean;
    variant?: "primary" | "destructive";
  };

  let {
    loading = false,
    fullWidth = true,
    disabled = false,
    type = "button",
    variant = "primary",
    children,
    ...rest
  }: $$Props = $props();
</script>

<button
  {type}
  disabled={disabled || loading}
  class="inline-flex items-center justify-center gap-2 h-12 px-5
           font-semibold rounded-xl shadow-sm active:scale-[0.99]
           transition-all
           disabled:opacity-60 disabled:cursor-not-allowed disabled:active:scale-100
           {fullWidth ? 'w-full' : ''}
           {variant === 'destructive'
    ? 'bg-error-base hover:bg-error-hover text-white'
    : 'bg-teren-primary hover:bg-teren-primary-hover text-white'}"
  {...rest}
>
  {#if loading}
    <div
      class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"
    ></div>
    <span class="sr-only">Loading</span>
  {/if}
  {@render children?.()}
</button>
