<script lang="ts">
	import { fade, fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { apiFetch } from '$lib/api';
	import type { Expense, Category } from '$lib/types';
	import ConfirmModal from './ConfirmModal.svelte';
	import { getCurrencySymbol, getCategoryEmoji, getCategoryName } from '$lib/utils';
	import { t } from '$lib/i18n/store';

	let {
		tripId,
		categories,
		isOpen,
		onClose,
		onRefreshSummary,
		placeId
	}: {
		tripId: string;
		categories: Category[];
		isOpen: boolean;
		onClose: () => void;
		onRefreshSummary: () => void;
		placeId?: string;
	} = $props();

	let expenses = $state<Expense[]>([]);
	let loading = $state(true);
	let editingId = $state<string | null>(null);
	let deleteConfirmId = $state<string | null>(null);
	let draft = $state({ amount: '', date: '', notes: '', category_id: '' });

	$effect(() => {
		if (isOpen && tripId) loadExpenses();
	});

	async function loadExpenses() {
		loading = true;
		try {
			const endpoint = placeId
				? `/trips/${tripId}/places/${placeId}/expenses`
				: `/trips/${tripId}/expenses`;
			expenses = await apiFetch<Expense[]>(endpoint);
		} finally {
			loading = false;
		}
	}

	function startEdit(exp: Expense) {
		editingId = exp.id;
		draft = {
			amount: String(exp.amount),
			date: exp.date.split('T')[0],
			notes: exp.notes || '',
			category_id: exp.category_id || ''
		};
	}

	async function saveEdit(id: string) {
		if (!draft.amount || parseFloat(draft.amount) <= 0) return;
		const payload = {
			...draft,
			amount: parseFloat(draft.amount),
			date: new Date(draft.date).toISOString()
		};
		await apiFetch(`/trips/${tripId}/expenses/${id}`, {
			method: 'PUT',
			body: JSON.stringify(payload)
		});
		editingId = null;
		loadExpenses();
		onRefreshSummary();
	}

	function requestDelete(id: string) {
		deleteConfirmId = id;
	}

	async function confirmDeletion() {
		if (!deleteConfirmId) return;
		const id = deleteConfirmId;
		deleteConfirmId = null; // Cierra el modal inmediatamente

		await apiFetch(`/trips/${tripId}/expenses/${id}`, { method: 'DELETE' });
		if (editingId === id) editingId = null;
		loadExpenses();
		onRefreshSummary();
	}

	function cancelDeletion() {
		deleteConfirmId = null;
	}

	let grouped = $derived.by(() => {
		const groups = new Map<string, Expense[]>();
		for (const exp of expenses) {
			const cat = categories.find((c) => c.id === exp.category_id);
			const key = cat ? cat.slug : 'others';
			if (!groups.has(key)) groups.set(key, []);
			groups.get(key)!.push(exp);
		}
		return Array.from(groups.entries());
	});
</script>

{#if isOpen}
	<div
		class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/40 backdrop-blur-sm"
		transition:fade={{ duration: 200 }}
		onclick={(e) => e.target === e.currentTarget && onClose()}
	>
		<div
			class="bg-teren-background w-full sm:max-w-2xl h-[85vh] sm:h-[80vh] rounded-t-2xl sm:rounded-2xl shadow-2xl flex flex-col overflow-hidden"
			transition:fly={{ y: 40, duration: 250, easing: cubicOut }}
			onclick={(e) => e.stopPropagation()}
		>
			<header class="flex items-center justify-between px-6 py-4 border-b border-teren-border">
				<h2 class="text-lg font-semibold text-teren-text-main tracking-tight">
					{$t('detail.expenses')}
				</h2>
				<button
					onclick={onClose}
					class="text-teren-text-muted hover:text-teren-text-main p-2 rounded-lg hover:bg-gray-100 transition active:scale-95"
					aria-label="Close"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						class="h-5 w-5"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
						><path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M6 18L18 6M6 6l12 12"
						/></svg
					>
				</button>
			</header>

			<main class="flex-1 overflow-y-auto p-6 space-y-6">
				{#if loading}
					<div class="flex justify-center py-10">
						<div
							class="w-8 h-8 border-3 border-teren-primary/30 border-t-teren-primary rounded-full animate-spin"
						></div>
					</div>
				{:else if expenses.length === 0}
					<div class="text-center py-12 text-teren-text-muted">{$t('detail.expenses_empty')}</div>
				{:else}
					{#each grouped as [slug, items] (slug)}
						<section>
							<h3
								class="text-sm font-semibold text-teren-text-muted uppercase tracking-wider mb-3 flex items-center gap-2"
							>
								{getCategoryEmoji(slug)}
								{getCategoryName(slug)}
								<span class="ml-auto font-normal normal-case text-xs">({items.length})</span>
							</h3>
							<div class="space-y-3">
								{#each items as exp (exp.id)}
									{#if editingId === exp.id}
										<!-- Modo Edición Minimalista -->
										<div
											class="p-3 bg-teren-background border-2 border-teren-primary/30 rounded-lg space-y-2"
										>
											<div class="flex gap-2">
												<input
													type="number"
													step="0.01"
													bind:value={draft.amount}
													class="w-24 px-2 py-1.5 text-sm font-bold bg-white border border-teren-border rounded focus:ring-2 focus:ring-teren-primary/30 outline-none"
													autofocus
												/>
												<div class="relative group/date flex-1">
													<svg
														class="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-teren-text-muted pointer-events-none"
														fill="none"
														viewBox="0 0 24 24"
														stroke="currentColor"
														><path
															stroke-linecap="round"
															stroke-linejoin="round"
															stroke-width="1.5"
															d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
														/></svg
													>
													<input
														type="date"
														bind:value={draft.date}
														class="w-full pl-8 pr-2 py-1.5 text-xs bg-white border border-teren-border rounded focus:ring-2 focus:ring-teren-primary/30 outline-none appearance-none [&::-webkit-calendar-picker-indicator]:opacity-0"
													/>
												</div>
											</div>
											<input
												type="text"
												bind:value={draft.notes}
												placeholder={$t('detail.notes_optional')}
												class="w-full px-2 py-1.5 text-xs bg-white border border-teren-border rounded focus:ring-2 focus:ring-teren-primary/30 outline-none"
												onkeydown={(e) => e.key === 'Enter' && saveEdit(exp.id)}
											/>
											<div class="flex justify-end gap-2 pt-1">
												<button
													onclick={() => (editingId = null)}
													class="px-3 py-1.5 text-xs text-teren-text-muted hover:text-teren-text-main hover:bg-gray-100 rounded transition"
												>
													{$t('common.cancel')}
												</button>
												<button
													onclick={() => saveEdit(exp.id)}
													class="px-3 py-1.5 text-xs bg-teren-primary hover:bg-teren-primary-hover text-white font-medium rounded transition active:scale-95"
												>
													{$t('common.done')}
												</button>
											</div>
										</div>
									{:else}
										<!-- Vista Compacta (Fecha a la derecha) -->
										<div
											class="group p-4 bg-teren-surface border border-teren-border rounded-xl hover:border-teren-primary/20 transition cursor-pointer"
											onclick={() => startEdit(exp)}
										>
											<div class="flex justify-between items-start gap-3">
												<div class="flex gap-3 flex-1 min-w-0">
													<!-- Icono con Tooltip -->
													<span class="text-xl select-none relative group/icon flex-shrink-0">
														{getCategoryEmoji(slug)}
														<span
															class="absolute bottom-full left-1/2 -translate-x-1/2 mb-1 px-2 py-1 bg-teren-text-main text-white text-xs rounded opacity-0 group-hover/icon:opacity-100 transition pointer-events-none whitespace-nowrap z-10"
														>
															{getCategoryName(slug)}
														</span>
													</span>

													<!-- Contenido -->
													<div class="flex-1 min-w-0">
														<div class="flex justify-between items-baseline">
															<span class="font-bold text-teren-text-main text-base">
																{exp.amount.toFixed(2)}
																{getCurrencySymbol(exp.currency)}
															</span>
															<span class="text-xs text-teren-text-muted flex-shrink-0 ml-2">
																{new Date(exp.date).toLocaleDateString('en-US', {
																	month: 'short',
																	day: 'numeric'
																})}
															</span>
														</div>
														{#if exp.notes}
															<p
																class="text-sm text-teren-text-muted mt-0.5 line-clamp-1 italic opacity-80"
															>
																{exp.notes}
															</p>
														{/if}
													</div>
												</div>

												<!-- Botón Delete (solo hover) -->
												<button
													onclick={(e) => e.stopPropagation() || requestDelete(exp.id)}
													class="opacity-0 group-hover:opacity-100 text-red-400 hover:text-red-600 p-1.5 rounded-lg hover:bg-red-50 transition active:scale-95 flex-shrink-0"
													aria-label="Delete"
												>
													<svg
														xmlns="http://www.w3.org/2000/svg"
														class="h-5 w-5"
														fill="none"
														viewBox="0 0 24 24"
														stroke="currentColor"
													>
														<path
															stroke-linecap="round"
															stroke-linejoin="round"
															stroke-width="2"
															d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
														/>
													</svg>
												</button>
											</div>
										</div>
									{/if}
								{/each}
							</div>
						</section>
					{/each}
				{/if}
			</main>
		</div>
	</div>

	<ConfirmModal
		isOpen={deleteConfirmId !== null}
		title={$t('confirm.delete_expense_title')}
		message={$t('confirm.delete_expense_message')}
		confirmText={$t('common.delete')}
		cancelText={$t('common.cancel')}
		isDestructive={true}
		onConfirm={confirmDeletion}
		onCancel={cancelDeletion}
	/>
{/if}
