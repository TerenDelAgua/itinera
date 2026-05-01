import { get } from 'svelte/store';
import { t } from '$lib/i18n/store';

// Dictionary mapping backend string snippets to frontend i18n keys
const ERROR_MAPPINGS: Record<string, string> = {
    "within trip dates": "errors.date_out_of_bounds",
    "Title and Date are required": "errors.required_fields",
    "invalid place ID": "errors.invalid_place",
    "Place does not belong to this trip": "errors.invalid_place",
    "Trip not found": "errors.trip_not_found"
    // Future backend errors should be mapped here
};

/**
 * Extracts a user-friendly, localized error message from an unknown error object.
 * It uses partial string matching against known backend responses.
 */
export function getFriendlyErrorMessage(error: unknown, params?: Record<string, string | number>): string {
    const tFunction = get(t);
    const genericError = tFunction("errors.generic") || "Something went wrong. Please try again.";

    if (error instanceof Error) {
        const msg = error.message;

        // Try to find a specific mapping
        for (const [backendMsg, i18nKey] of Object.entries(ERROR_MAPPINGS)) {
            if (msg.includes(backendMsg)) {
                return tFunction(i18nKey as any, params) || genericError;
            }
        }

        // If no specific mapping is found, but it's an API error, return generic
        if (msg.startsWith("API Error:")) {
            return genericError;
        }

        // Fallback to the raw message if it's completely unhandled (useful for debugging)
        // In production, you might want to return genericError here too depending on policy.
        return msg;
    }

    // Completely unknown error type
    return genericError;
}
