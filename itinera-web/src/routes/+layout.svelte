<script lang="ts">
	let {children: childrenProp} = $props();
</script>

<div class="flex flex-col min-h-screen">
	<!-- Header -->
	<header class="bg-white shadow-sm sticky top-0 z-50">
		<div class="max-w-5xl mx-auto px-4 py-4 flex justify-between items-center">
			<!-- Logo / Marca -->
			<a href="/" class="text-2xl font-bold text-teren-primary tracking-tight">
				Itinera
			</a>
			
			<!-- Menú Simple -->
			<nav>
				<button class="text-sm font-medium text-teren-textMuted hover:text-teren-primary transition">
					My Trips
				</button>
			</nav>
		</div>
	</header>

	<!-- Contenido Principal (Aquí entrará la página) -->
	<main class="flex-1 max-w-5xl mx-auto w-full px-4 py-8">
		 {@render childrenProp?.()}
	</main>

	<!-- Footer -->
	<footer class="py-6 text-center text-teren-textMuted text-xs">
		Built with TEREN Design System
	</footer>
</div>