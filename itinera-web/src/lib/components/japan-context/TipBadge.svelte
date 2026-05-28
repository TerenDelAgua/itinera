<script lang="ts">
  import { slide } from "svelte/transition";
  import { t } from "$lib/i18n/store";
  import type { RuleDisplay } from "$lib/services/japanContext";

  let { rules } = $props<{ rules: RuleDisplay[] }>();

  let isExpanded = $state(false);
  let showAll = $state(false);

  let displayedRules = $derived(showAll ? rules : rules.slice(0, 3));
  let remainingCount = $derived(
    rules.length > 3 && !showAll ? rules.length - 3 : 0,
  );

  function toggleExpand(e: Event) {
    e.stopPropagation();
    isExpanded = !isExpanded;
    if (!isExpanded) showAll = false;
  }

  function handleShowAll(e: Event) {
    e.stopPropagation();
    showAll = true;
  }
</script>

{#if rules.length > 0}
  <div
    class="mt-2 text-sm flex flex-col gap-2 relative z-10 w-full"
    role="presentation"
    onclick={(e) => e.stopPropagation()}
  >
    <!-- Badge Header -->
    <div class="pl-[4.5rem] flex w-full">
      <button
        onclick={toggleExpand}
        class="inline-flex items-center gap-1.5 px-2 py-1 rounded-md transition-colors bg-teren-primary-subtle border border-teren-primary/30 text-teren-primary-hover self-start cursor-pointer hover:bg-teren-primary/10 active:scale-95"
      >
        <span class="text-xs">
          {#if rules.length === 1}
            {$t("japan_context.ui.tip_singular" as any)}
          {:else}
            {$t("japan_context.ui.tips_plural" as any, {
              count: rules.length.toString(),
            })}
          {/if}
        </span>
        <svg
          class="w-3.5 h-3.5 transition-transform duration-200 {isExpanded
            ? 'rotate-180'
            : ''}"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M19 9l-7 7-7-7"
          />
        </svg>
      </button>
    </div>

    <!-- Expanded Tips List -->
    {#if isExpanded}
      <div
        class="flex flex-col gap-2 mt-1.5 w-full"
        transition:slide={{ duration: 200 }}
      >
        {#each displayedRules as rule (rule.id)}
          <div
            class="bg-teren-surface border border-teren-border rounded-lg p-3 text-xs flex gap-3 shadow-sm"
          >
            <div
              class="text-base leading-none shrink-0 pt-0.5 select-none"
              aria-hidden="true"
            >
              {rule.icon}
            </div>
            <div class="flex flex-col gap-1 text-teren-text-main leading-snug">
              <span class="font-bold text-sm"
                >{$t(`japan_context.rules.${rule.id}.title` as any)}</span
              >
              <span class="text-teren-text-muted mt-0.5"
                >{$t(`japan_context.rules.${rule.id}.body` as any)}</span
              >
              <div class="flex flex-col gap-0.5 mt-1">
                <span class="text-error-base/90 italic flex gap-1 items-start">
                  <span class="shrink-0 text-xs">⚠️</span>
                  {$t(`japan_context.rules.${rule.id}.consequence` as any)}
                </span>
                <span
                  class="text-success-base/90 italic flex gap-1 items-start"
                >
                  <span class="shrink-0 text-xs">💡</span>
                  {$t(`japan_context.rules.${rule.id}.tip` as any)}
                </span>
              </div>
            </div>
          </div>
        {/each}

        {#if remainingCount > 0}
          <button
            onclick={handleShowAll}
            class="text-xs font-semibold text-teren-text-muted hover:text-teren-primary transition-colors self-center py-2 cursor-pointer"
          >
            {$t("japan_context.ui.more_tips" as any, {
              count: remainingCount.toString(),
            })}
          </button>
        {/if}
      </div>
    {/if}
  </div>
{/if}
