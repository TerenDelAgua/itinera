import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch }) => {
    const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

    // Fetch stats and demo trips in parallel
    try {
        const tripsRes = await fetch(`${API_URL}/trips?demos_only=true`);
        if (!tripsRes.ok) {
            console.error(`Failed to fetch demos: ${tripsRes.status} ${tripsRes.statusText}`);
        }
        const tripsData = tripsRes.ok ? await tripsRes.json() : { trips: [] };
        
        const statsRes = await fetch(`${API_URL}/stats/public`);
        if (!statsRes.ok) {
            console.error(`Failed to fetch stats: ${statsRes.status} ${statsRes.statusText}`);
        }
        const stats = statsRes.ok ? await statsRes.json() : { total_trips: 0, total_places: 0, total_expenses: 0 };

        // Process and filter demos
        const demos = (tripsData.trips || [])
            .sort((a: any, b: any) => {
                // Pin "Japón" to the top
                const aIsJapan = a.name.toLowerCase().includes('japón') || a.name.toLowerCase().includes('japan');
                const bIsJapan = b.name.toLowerCase().includes('japón') || b.name.toLowerCase().includes('japan');
                if (aIsJapan && !bIsJapan) return -1;
                if (!aIsJapan && bIsJapan) return 1;
                return 0; // Keep original order (created_at DESC from backend)
            })
            .slice(0, 3)
            .map((t: any) => {
                const start = t.start_date ? new Date(t.start_date) : null;
                const end = t.end_date ? new Date(t.end_date) : null;
                const day_count = start && end 
                    ? Math.max(1, Math.ceil((end.getTime() - start.getTime()) / 86400000) + 1) 
                    : 1;

                return {
                    id: t.id,
                    name: t.name,
                    day_count,
                    destination_count: t.place_count || 0,
                    base_currency: t.base_currency || 'EUR',
                };
            });

        return {
            totalTrips: stats.total_trips || 0,
            demos
        };
    } catch (err) {
        console.error('Landing server load error:', err);
        return {
            totalTrips: 0,
            demos: []
        };
    }
};
