<script lang="ts">
  import { t } from '$lib/i18n/store';
</script>

<footer class="bg-teren-background border-t border-teren-border py-16 px-6 transition-colors duration-300">
  <div class="max-w-6xl mx-auto flex flex-col items-center text-center">
    <div class="flex items-center gap-2 mb-4">
      <span class="text-orange-500 text-xl">🧡</span>
      <span class="text-xl font-bold tracking-tight text-teren-text-main uppercase">TEREN</span>
    </div>
    
    <p class="text-teren-text-muted font-medium mb-8 max-w-md italic">
      "Built with intention. Designed for flow."
    </p>

    <div class="flex flex-wrap justify-center gap-x-6 gap-y-3 text-sm text-teren-text-muted mb-10 font-medium">
      <span class="flex items-center gap-1.5">Fukuoka, Japan 📍</span>
      <a href="mailto:hello@teren.dev" class="hover:text-teren-primary transition-colors">hello@teren.dev</a>
      <a href="https://github.com/teren-dev" target="_blank" rel="noopener" class="hover:text-teren-primary transition-colors">github.com/teren-dev</a>
    </div>

    <div class="mb-12">
      <a 
        href="https://forms.gle/M8RuESQufXRmcgz37" 
        target="_blank" 
        rel="noopener"
        class="inline-flex items-center gap-2 px-6 py-2.5 rounded-xl text-sm font-bold text-teren-text-muted hover:text-teren-primary hover:bg-teren-primary-subtle transition-all duration-300 border border-transparent hover:border-teren-primary/10"
      >
        {$t('landing.give_feedback')}
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
        </svg>
      </a>
    </div>

    <p class="text-[11px] font-bold text-teren-text-muted/60 uppercase tracking-widest">
      &copy; 2026 TEREN. Privacy-first. No tracking.
    </p>
  </div>
</footer>
