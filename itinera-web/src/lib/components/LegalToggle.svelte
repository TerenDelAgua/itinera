<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import type { LegalLocale } from '$lib/legal/manifest';
	// page is used to read the current pathname for the URL rewrite
	// in pickLocale().

	let { currentLocale }: { currentLocale: LegalLocale } = $props();

	/**
	 * Switches the legal locale by setting the `itinera_legal_lang`
	 * cookie (1 year, SameSite=Lax) and navigating to the same URL
	 * with `?lang=<new>`. The page server load already reads the
	 * `?lang=` query and returns the right variant.
	 *
	 * Accessibility: `aria-pressed` reflects the current selection,
	 * and both buttons are `type="button"` so they never submit any
	 * surrounding form.
	 */
	function pickLocale(next: LegalLocale) {
		if (next === currentLocale) return;
		// Persist preference (1 year, SameSite=Lax, no third-party).
		document.cookie = `itinera_legal_lang=${next}; max-age=31536000; path=/; samesite=lax`;
		// Build the localised URL: `/terms/es`, `/privacy/en`, etc.
		// We keep the current path's first segment (terms|privacy) and
		// append `/<locale>` so the prerendered static HTML variant
		// (see +page.ts#entries) is served. Any other query params
		// (utm, etc.) are preserved.
		const path = page.url.pathname.replace(/\/(es|en)$/, '');
		void goto(`${path}/${next}`, { keepFocus: true });
	}
</script>

<!--
  Group of two buttons styled as a segmented toggle. Each one carries
  its own `aria-pressed` so screen readers can announce which option
  is currently selected without relying on visual state.
-->
<div
	class="inline-flex rounded-lg border border-teren-border overflow-hidden"
	role="group"
	aria-label={currentLocale === 'es' ? 'Cambiar idioma del documento legal' : 'Change legal-document language'}
>
	<button
		type="button"
		aria-pressed={currentLocale === 'es'}
		class="px-3 py-1 text-xs font-medium transition-colors {currentLocale === 'es'
			? 'bg-teren-primary text-white'
			: 'bg-transparent text-teren-text-muted hover:bg-teren-interactive-hover hover:text-teren-text-main'}"
		onclick={() => pickLocale('es')}
	>
		ES
	</button>
	<button
		type="button"
		aria-pressed={currentLocale === 'en'}
		class="px-3 py-1 text-xs font-medium transition-colors border-l border-teren-border {currentLocale === 'en'
			? 'bg-teren-primary text-white'
			: 'bg-transparent text-teren-text-muted hover:bg-teren-interactive-hover hover:text-teren-text-main'}"
		onclick={() => pickLocale('en')}
	>
		EN
	</button>
</div>