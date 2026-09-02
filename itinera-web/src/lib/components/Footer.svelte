<script lang="ts">
	import { CONTACT } from '$lib/legal/manifest';

	// Dynamic year via `new Date()` so the footer never gets stale.
	const year = new Date().getFullYear();
</script>

<footer class="border-t border-teren-border py-8 mt-auto">
	<div
		class="max-w-6xl mx-auto px-4 flex flex-col sm:flex-row justify-between items-center gap-3 text-sm text-teren-text-muted"
	>
		<p>© {year} TEREN</p>
		<nav
			class="flex flex-wrap justify-center gap-x-6 gap-y-2"
			aria-label="Legal"
		>
			<a href="/terms" class="hover:text-teren-text-main transition">Terms</a>
			<a href="/privacy" class="hover:text-teren-text-main transition">Privacy</a>
			<a
				href="mailto:{CONTACT.general}"
				class="hover:text-teren-text-main transition"
			>
				{CONTACT.general}
			</a>
		</nav>
	</div>
</footer>