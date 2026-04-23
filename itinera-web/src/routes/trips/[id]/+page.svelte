<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { apiFetch } from "$lib/api";
  import { t, locale } from "$lib/i18n/store";
  import type { Trip } from "$lib/types/Trip";
  import { page } from "$app/stores";
  import ExpenseDrawer from '$lib/components/ExpenseDrawer.svelte';
  import type { Category, CategorySummary, Expense } from "$lib/types";
  import ExpenseQuickAdd from "$lib/components/ExpenseQuickAdd.svelte";
  import ExpenseSummaryPills from "$lib/components/ExpenseSummaryPills.svelte";

  let tripId = "";
  let trip = $state<Trip | null>(null);
  let isLoading = $state(true);

  // States for inline editing
  let isEditingName = $state(false);
  let isEditingCurrency = $state(false);
  let isEditingDescription = $state(false);

  // Form local state (runes)
  let name = $state("");
  let startDate = $state("");
  let endDate = $state("");
  let baseCurrency = $state("EUR");
  let description = $state("");

  // Expense states
  let categories = $state<Category[]>([]);
  let summary = $state<CategorySummary[]>([]);
  let expenses = $state<Expense[]>([]);
  let isDrawerOpen = $state(false);


  $effect(() => {
    if ($page.params.id) {
      tripId = $page.params.id;
      loadTrip();
    }

    if(tripId) {
      Promise.all([
        apiFetch<Category[]>(`/trips/${tripId}/expenses/categories`).then(data => categories = data),
        apiFetch<CategorySummary[]>(`/trips/${tripId}/expenses/summary`),
        apiFetch<Expense[]>(`/trips/${tripId}/expenses`)
    
      ]).then(([cats, sum, exps]) => {
        categories = cats;
        summary = sum;
        expenses = exps;
      })
    }
  });

  async function loadTrip() {
    isLoading = true;
    try {
      const data = await apiFetch<Trip>(`/trips/${tripId}`);
      trip = data;
      name = data.name;
      startDate = data.start_date;
      endDate = data.end_date;
      baseCurrency = data.base_currency;
      description = data.description || "";
    } catch {
      goto("/");
    } finally {
      isLoading = false;
    }
  }

  async function saveField(field: string, value: string) {
    if (!trip) return;

    if (trip[field as keyof Trip] === value) return;

    if (field === "name" && !value.trim()) return;

    const payload: Record<string, string> = { [field]: value };

    const previousTrip = { ...trip };
    // @ts-ignore - Dynamic update for optimistic UI
    trip[field as keyof Trip] = value;

    try {
      const updated = await apiFetch<Trip>(`/trips/${tripId}`, {
        method: "PUT",
        body: JSON.stringify(payload),
      });
      trip = updated;
    } catch {
      // Revert if error
      console.error("Error updating trip");
      trip = previousTrip;
      // Refresh data to be sure
      loadTrip();
    }
  }

  function formatDate(dateStr: string) {
    if (!dateStr) return "";
    const [year, month, day] = dateStr.split("-").map(Number);
    const date = new Date(year, month - 1, day);
    return new Intl.DateTimeFormat($locale, {
      day: "2-digit",
      month: "short",
      year: "numeric",
    }).format(date);
  }

  function loadSummary() {
    apiFetch<CategorySummary[]>(`/trips/${tripId}/expenses/summary`)
      .then((newSummary) => summary = newSummary);
  }

  function handleExpenseAdded(newExp: Expense) {
    expenses = [...expenses, newExp];
    loadSummary();
  }
</script>

<svelte:head>
  <title>{trip ? `${name} | Itinera` : $t("common.loading")}</title>
</svelte:head>

<div class="max-w-3xl mx-auto space-y-8 animate-fade-in">
  
  {#if isLoading}
    <!-- Skeleton loader compact -->
    <div class="flex items-center gap-4 animate-pulse">
      <div class="w-8 h-8 bg-gray-200 rounded-full"></div>
      <div class="h-10 bg-gray-200 rounded-lg flex-1"></div>
    </div>
    <div class="h-48 bg-white border border-teren-border rounded-2xl animate-pulse"></div>
  {:else if trip}
    <!-- Cabecera Compacta (Navegación + Título + Moneda) -->
    <div class="flex items-start gap-3 sm:gap-4">
      <!-- Botón Volver con Flecha Larga -->
      <button
        onclick={() => goto("/")}
        class="mt-1 p-2 -ml-2 text-teren-text-muted hover:text-teren-text-main transition-all duration-200 group rounded-full hover:bg-gray-100 active:scale-90"
        aria-label="Volver"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 transform group-hover:-translate-x-1 transition-transform" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
        </svg>
      </button>

      <!-- Título y Fechas -->
      <div class="flex-1 min-w-0 pt-0.5">
        {#if isEditingName}
          <input
            type="text"
            bind:value={name}
            onblur={() => {
              saveField("name", name);
              isEditingName = false;
            }}
            onkeydown={(e) => {
              if (e.key === "Enter") {
                saveField("name", name);
                isEditingName = false;
              }
            }}
            class="text-3xl font-bold text-teren-text-main bg-transparent border-b-2 border-teren-primary focus:outline-none w-full tracking-tight"
            autofocus
          />
        {:else}
          <h1
            onclick={() => (isEditingName = true)}
            class="text-3xl font-bold text-teren-text-main tracking-tight cursor-pointer hover:text-teren-primary transition-colors duration-200 truncate"
            title={name}
          >
            {name}
          </h1>
        {/if}
        
        <div class="flex items-center gap-2 text-teren-text-muted font-medium mt-1">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-3.5 w-3.5 opacity-70"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
            />
          </svg>
          <span class="text-xs sm:text-sm">
            {formatDate(startDate)} — {formatDate(endDate)}
          </span>
        </div>
      </div>

      <!-- Selector de Moneda -->
      <div class="pt-1">
        {#if isEditingCurrency}
          <select
            bind:value={baseCurrency}
            onchange={() => {
              saveField("base_currency", baseCurrency);
              isEditingCurrency = false;
            }}
            onblur={() => (isEditingCurrency = false)}
            class="bg-white border border-teren-border rounded-lg px-2 py-1 text-xs font-bold text-teren-primary focus:outline-none focus:ring-2 focus:ring-teren-primary/30"
            autofocus
          >
            <option value="EUR">EUR</option>
            <option value="USD">USD</option>
            <option value="JPY">JPY</option>
            <option value="GBP">GBP</option>
          </select>
        {:else}
          <button
            onclick={() => (isEditingCurrency = true)}
            class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-teren-primary-subtle text-teren-primary-hover rounded-full text-xs font-bold border border-teren-primary/20 hover:border-teren-primary/40 transition-all active:scale-95"
          >
            {baseCurrency}
          </button>
        {/if}
      </div>
    </div>

    <!-- Descripción -->
    <section class="group">
      {#if isEditingDescription}
        <textarea
          bind:value={description}
          onblur={() => {
            saveField("description", description);
            isEditingDescription = false;
          }}
          class="w-full bg-white border border-teren-border rounded-xl p-4 text-teren-text-main text-base leading-relaxed focus:outline-none focus:ring-2 focus:ring-teren-primary/30 transition-all resize-none"
          rows="3"
          autofocus
          placeholder={$t("trip_form.description_placeholder")}
        ></textarea>
      {:else}
        <div
          onclick={() => (isEditingDescription = true)}
          class="cursor-pointer p-4 -m-4 rounded-xl hover:bg-white hover:shadow-sm transition-all duration-200 border border-transparent hover:border-teren-border/50"
        >
          {#if description}
            <p class="text-teren-text-muted text-base leading-relaxed">
              {description}
            </p>
          {:else}
            <p class="text-teren-text-muted/50 italic text-sm">
              {$t("detail.no_description")}
            </p>
          {/if}
        </div>
      {/if}
    </section>

    <!-- Cuadrícula de Secciones -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
    
<section class="bg-teren-surface rounded-2xl border border-teren-border p-5 shadow-sm flex flex-col gap-4">
  <!-- Header: título + total -->
  <div class="flex items-center justify-between">
    <h2 class="text-base font-semibold text-teren-text-main tracking-tight">Gastos</h2>
    <span class="text-xl font-bold text-teren-primary">
      {(summary || []).reduce((a, b) => a + b.total, 0).toFixed(2)} {trip?.base_currency || 'EUR'}
    </span>
  </div>

  <!-- Pills de resumen por categoría -->
  <ExpenseSummaryPills {categories} {summary} currency={trip?.base_currency || 'EUR'} />

  <!-- Quick add inline -->
  {#if categories.length > 0}
    <ExpenseQuickAdd tripId={tripId} {categories} onSuccess={handleExpenseAdded} />
  {/if}

  <!-- Botón para abrir el drawer con la lista completa -->
  <button
    onclick={() => isDrawerOpen = true}
    class="w-full text-sm font-medium text-teren-text-muted hover:text-teren-primary transition-colors duration-200 py-1 text-center"
  >
    Ver todos los gastos →
  </button>
</section>

      <!-- Lugares -->
      <button
        class="bg-teren-surface p-8 rounded-2xl border border-teren-border hover:border-teren-primary/30 shadow-sm hover:shadow-xl hover:shadow-orange-900/5 transition-all duration-300 flex flex-col items-center text-center group active:scale-98"
      >
        <div
          class="w-14 h-14 mb-4 rounded-full bg-teren-background border border-teren-border flex items-center justify-center group-hover:bg-teren-primary-subtle transition-colors duration-300"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-7 w-7 text-teren-text-muted group-hover:text-teren-primary transition-colors duration-300"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"
            />
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"
            />
          </svg>
        </div>
        <h3 class="text-lg font-bold text-teren-text-main mb-1">
          {$t("detail.places")}
        </h3>
        <p class="text-xs text-teren-text-muted mt-2 font-medium opacity-70">
          {$t("detail.places_empty")}
        </p>
      </button>
    </div>
  {/if}
</div>

<!-- Drawer de Gastos — fuera del flujo, como overlay fijo -->
<ExpenseDrawer
  {tripId}
  {categories}
  isOpen={isDrawerOpen}
  onClose={() => isDrawerOpen = false}
  onRefreshSummary={loadSummary}
/>

<style>
  .animate-fade-in {
    animation: fadeIn 0.4s ease-out forwards;
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
      transform: translateY(10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
</style>