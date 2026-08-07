<script lang="ts">
    /**
     * Checkbox with rich label content. Used by the registration form to
     * render the "I accept the terms" line with inline links.
     *
     * The label can be either a plain string or an HTML snippet — when
     * it is an HTML snippet the parent has already run it through
     * `$t()` and pre-substituted the links. We render it via `{@html}`
     * ONLY when the parent opts in via the `rich` prop, because we
     * never want to feed raw user input into @html. The default `label`
     * prop renders as plain text.
     */
    interface Props {
        label: string;
        checked: boolean;
        /** When true, the `label` is rendered as HTML. Use only with trusted content. */
        rich?: boolean;
        error?: string | null;
        id?: string;
    }

    let { label, checked = $bindable(), rich = false, error = null, id }: Props = $props();

    const inputId = $derived(id ?? `checkbox-${Math.random().toString(36).slice(2, 9)}`);
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
            aria-invalid={error ? 'true' : undefined}
            aria-describedby={error ? errorId : undefined}
            class="mt-0.5 h-4 w-4 rounded border-teren-border text-teren-primary
                   focus:ring-2 focus:ring-teren-primary/30 focus:ring-offset-0
                   checked:bg-teren-primary checked:border-teren-primary
                   transition-colors cursor-pointer"
        />
        <span class="leading-snug">
            {#if rich}
                {@html label}
            {:else}
                {label}
            {/if}
        </span>
    </label>
    {#if error}
        <p id={errorId} class="text-xs text-error-base font-medium ml-6.5" role="alert">
            {error}
        </p>
    {/if}
</div>
