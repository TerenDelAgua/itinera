<script lang="ts">
  import Card from '$lib/components/Card.svelte';

  // Mock data - luego vendrá de Go API
  let trips = [
    { 
      id: '1', 
      name: 'Japan 2026', 
      start: '2026-10-01', 
      end: '2026-10-15', 
      places: ['Tokyo', 'Kyoto', 'Nara'],
      progress: 45 
    },
    { 
      id: '2', 
      name: 'Italy Roadtrip', 
      start: '2026-05-10', 
      end: '2026-05-20', 
      places: ['Rome', 'Florence', 'Venice'],
      progress: 10 
    },
    { 
      id: '3', 
      name: 'Japan - Fukuoka Exploration', 
      start: '2027-03-15', 
      end: '2027-03-25', 
      places: ['Fukuoka', 'Nagasaki'],
      progress: 0 
    }
  ];

  function handleCreateTrip() {
    alert('🚀 Próximamente: Formulario de creación de viajes');
  }
</script>

<div class="max-w-5xl mx-auto px-4 py-12 space-y-10">
  
  <!-- Header Section -->
  <section class="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-4">
    <div>
      <h1 class="text-4xl font-extrabold text-teren-text-main tracking-tight">
        Your Trips
      </h1>
      <p class="text-teren-text-muted mt-2 text-lg">
        Plan your adventures with elegance and simplicity.
      </p>
    </div>
    
    <button 
      onclick={handleCreateTrip}
      class="bg-teren-primary hover:bg-teren-primary-hover text-white font-bold py-3 px-6 rounded-lg shadow-md transition-all duration-200 active:scale-95 inline-flex items-center gap-2">
      <span class="text-xl leading-none">+</span> 
      <span>New Trip</span>
    </button>
  </section>

  <!-- Trips Grid -->
  {#if trips.length > 0}
    <div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
      {#each trips as trip}
        <Card href={`/trip/${trip.id}`}
          title={trip.name} 
          subtitle={`${trip.start} - ${trip.end}`}
          clickable={true}
        >
          <!-- Progress Bar -->
          <div class="mb-4">
            <div class="flex justify-between text-sm mb-1">
              <span class="text-teren-text-muted">Planning progress</span>
              <span class="font-semibold text-teren-primary">{trip.progress}%</span>
            </div>
            <div class="w-full bg-teren-surface rounded-full h-2 overflow-hidden">
              <div 
                class="bg-teren-primary h-full rounded-full transition-all duration-500"
                style="width: {trip.progress}%"
              ></div>
            </div>
          </div>

          <!-- Places Tags -->
          <div class="flex flex-wrap gap-2">
            {#each trip.places as place}
              <span class="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-teren-surface text-teren-primary border border-teren-primary/20">
                {place}
              </span>
            {/each}
          </div>
        </Card>
      {/each}
    </div>
  {:else}
    <!-- Empty State -->
    <div class="text-center py-20">
      <div class="text-6xl mb-4">🗾</div>
      <h3 class="text-xl font-bold text-teren-text-main mb-2">No trips yet</h3>
      <p class="text-teren-text-muted mb-6">Start planning your first adventure!</p>
      <button 
        onclick={handleCreateTrip}
        class="bg-teren-primary hover:bg-teren-primary-hover text-white font-bold py-3 px-8 rounded-lg shadow-md transition-all">
        Create Your First Trip
      </button>
    </div>
  {/if}
</div>