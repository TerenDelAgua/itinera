import { apiFetch } from '$lib/api';
import type { Activity } from '$lib/types/Activity';

export const activityApi = {
    list: (tripId: string) => apiFetch<Activity[]>(`/trips/${tripId}/activities`),

    create: (tripId: string, data: Omit<Activity, 'id' | 'trip_id' | 'created_at'>) =>
        apiFetch<Activity>(`/trips/${tripId}/activities`, { method: 'POST', body: JSON.stringify(data) }),

    update: (tripId: string, id: string, data: Partial<Activity>) =>
        apiFetch<Activity>(`/trips/${tripId}/activities/${id}`, { method: 'PUT', body: JSON.stringify(data) }),

    delete: (tripId: string, id: string) =>
        apiFetch(`/trips/${tripId}/activities/${id}`, { method: 'DELETE' }),
};