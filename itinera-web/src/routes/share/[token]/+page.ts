import type { PageLoad } from "./$types";
import { APP_URL } from "$lib/config";

export const load: PageLoad = async ({ params, fetch }) => {
    const { token } = params;

    try {
        const res = await fetch(`${import.meta.env.VITE_API_URL}/share/${token}`, {
            credentials: "include",
        });

        if (!res.ok) {
            return {
                status: res.status,
                error: res.status === 404
                    ? "share.errors.not_found"
                    : "share.errors.generic",
                token,
                shareUrl: `${APP_URL}/share/${token}`,
                ogImage: `${APP_URL}/og-trip.svg`,
            };
        }

        const trip = await res.json();

        return {
            trip,
            token,
            shareUrl: `${APP_URL}/share/${token}`,
            ogImage: `${APP_URL}/og-trip.svg`,
        };
    } catch {
        return {
            status: 500,
            error: "share.errors.network",
            token,
            shareUrl: `${APP_URL}/share/${token}`,
            ogImage: `${APP_URL}/og-trip.svg`,
        };
    }
}