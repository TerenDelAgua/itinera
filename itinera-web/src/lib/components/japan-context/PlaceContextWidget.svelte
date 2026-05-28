<script lang="ts">
  import { slide } from "svelte/transition";
  import { t } from "$lib/i18n/store";
  import type { RuleDisplay } from "$lib/services/japanContext";

  let { rules = [] } = $props<{ rules: RuleDisplay[] }>();

  let isExpanded = $state(false);

  function toggleExpand(e: Event) {
    e.stopPropagation();
    isExpanded = !isExpanded;
  }
</script>

{#if rules && rules.length > 0}
  <div
    class="mt-2 text-sm flex flex-col gap-2 relative z-10 w-full animate-fade-in"
    role="presentation"
    onclick={(e) => e.stopPropagation()}
  >
    <!-- Collapsible Header Panel (Full Width) -->
    <button
      onclick={toggleExpand}
      class="flex w-full items-center justify-between px-5 py-3.5 rounded-xl transition-all bg-teren-primary-subtle border border-teren-primary text-teren-primary-hover cursor-pointer hover:bg-teren-primary/10 active:scale-[0.995] shadow-sm shadow-teren-primary/10"
    >
      <span class="text-sm font-semibold flex items-center gap-2.5">
        {#if rules.length === 1}
          {$t("japan_context.ui.cultural_tip" as any, { count: "1" })}
        {:else}
          {$t("japan_context.ui.cultural_tip" as any, {
            count: rules.length.toString(),
          })}
        {/if}
      </span>
      <svg
        class="w-5 h-5 text-teren-primary transition-transform duration-200 {isExpanded
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

    <!-- Expanded Content -->
    {#if isExpanded}
      <div
        class="flex flex-col gap-3 mt-1.5 w-full"
        transition:slide={{ duration: 250 }}
      >
        {#each rules as r (r.id)}
          <div
            class="bg-teren-surface border border-teren-border rounded-xl p-4 text-sm flex gap-3.5 shadow-md w-full hover:border-teren-primary/20 transition-all"
          >
            <div
              class="text-xl leading-none shrink-0 pt-0.5 select-none"
              aria-hidden="true"
            >
              {r.icon}
            </div>
            <div
              class="flex flex-col gap-1.5 text-teren-text-main leading-snug w-full"
            >
              <span class="font-bold text-base text-teren-text-main"
                >{$t(`japan_context.rules.${r.id}.title` as any)}</span
              >
              <span class="text-teren-text-muted text-[13px]"
                >{$t(`japan_context.rules.${r.id}.body` as any)}</span
              >
              <div
                class="flex flex-col gap-1 mt-2 bg-error-subtle/30 rounded-lg p-2 border border-error-base/10"
              >
                <span
                  class="text-error-base italic flex gap-2 items-start text-xs font-medium"
                >
                  <span class="shrink-0">⚠️</span>
                  {$t(`japan_context.rules.${r.id}.consequence` as any)}
                </span>
              </div>
              <div
                class="flex flex-col gap-1 mt-1.5 bg-success-subtle/30 rounded-lg p-2 border border-success-base/10"
              >
                <span
                  class="text-success-base italic flex gap-2 items-start text-xs font-medium"
                >
                  <span class="shrink-0">💡</span>
                  {$t(`japan_context.rules.${r.id}.tip` as any)}
                </span>
              </div>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
{/if}
