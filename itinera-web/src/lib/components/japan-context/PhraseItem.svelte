<script lang="ts">
  import { t } from '$lib/i18n/store';
  import type { PhraseDisplay } from '$lib/services/japanContext';

  let { phrase } = $props<{ phrase: PhraseDisplay }>();

  let showToast = $state(false);
  let showHint = $state(false);

  function copyToClipboard(e: Event) {
    e.stopPropagation();
    navigator.clipboard.writeText(phrase.ja).then(() => {
      showToast = true;
      setTimeout(() => showToast = false, 2000);
    });
  }

  function toggleHint(e: Event) {
    e.stopPropagation();
    showHint = !showHint;
  }
</script>

<button
  class="w-full text-left bg-teren-surface border border-teren-border hover:border-teren-primary/50 hover:bg-teren-background rounded-xl p-4 shadow-sm transition-all relative group cursor-pointer active:scale-[0.99] focus:outline-none focus:ring-2 focus:ring-teren-primary/30"
  onclick={copyToClipboard}
  oncontextmenu={(e) => { e.preventDefault(); toggleHint(e); }}
>
  <div class="flex flex-col gap-1 pr-8">
    <div class="text-xl font-bold text-teren-text-main font-sans tracking-wide">{phrase.ja}</div>
    <div class="text-sm font-medium text-teren-primary">{phrase.romaji}</div>
    <div class="text-xs text-teren-text-muted mt-1 leading-snug">"{$t(`japan_context.phrases.${phrase.id}.translation` as any)}"</div>
    
    {#if showHint}
      <div class="mt-2 text-[11px] text-teren-text-muted/80 bg-teren-background p-2 rounded flex gap-1.5 items-start">
        <span class="shrink-0">💡</span> {$t(`japan_context.phrases.${phrase.id}.usage` as any)}
      </div>
    {/if}
  </div>

  <div class="absolute right-4 top-1/2 -translate-y-1/2 opacity-0 group-hover:opacity-100 transition-opacity text-teren-primary">
    <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
    </svg>
  </div>

  {#if showToast}
    <div class="absolute inset-0 bg-teren-surface/90 backdrop-blur-[2px] rounded-xl flex items-center justify-center text-teren-primary font-bold z-10">
      Copied!
    </div>
  {/if}
</button>
