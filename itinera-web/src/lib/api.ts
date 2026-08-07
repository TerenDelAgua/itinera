// Lightweight HTTP client for the Itinera backend.
//
// Auth: cookies only (`session_id`, `itinera_access`, `itinera_refresh`).
// Every request sets `credentials: 'include'` so the browser sends them.
//
// Error handling: backend returns `{ error: { code: string, message: string } }`
// for every non-2xx response (Spec 017 §9.2). `apiFetch` parses that body
// and throws an `ApiError` carrying the structured `code`, so callers can
// react to specific codes (e.g. EMAIL_ALREADY_EXISTS) without string-matching
// the human-readable message.
const API_URL = import.meta.env.VITE_API_URL;

if (!API_URL) {
    throw new Error('VITE_API_URL is not defined in the environment variables');
}

/**
 * Error thrown by `apiFetch` and helpers when the backend returns non-2xx.
 *
 * `code` is the stable machine identifier from the backend's error registry
 * (e.g. `EMAIL_ALREADY_EXISTS`). It is the only field a caller should branch
 * on — the `message` is localised on the server and may change.
 */
export class ApiError extends Error {
    readonly code: string;
    readonly status: number;
    readonly fields?: Record<string, string>;

    constructor(code: string, status: number, message: string, fields?: Record<string, string>) {
        super(message);
        this.name = 'ApiError';
        this.code = code;
        this.status = status;
        this.fields = fields;
    }
}

interface BackendErrorBody {
    error?: {
        code?: string;
        message?: string;
        fields?: Record<string, string>;
    };
}

export interface ApiFetchOptions extends Omit<RequestInit, 'body'> {
    body?: unknown;
}

/**
 * Low-level fetch wrapper. Use the typed helpers (`register`, `login`, ...)
 * for everything but custom endpoints. This function throws `ApiError` on
 * any non-2xx response with the parsed `code`; for 2xx it returns the
 * parsed JSON body (or `{}` for empty responses).
 */
export async function apiFetch<T = unknown>(
    endpoint: string,
    options: ApiFetchOptions = {}
): Promise<T> {
    const { body, headers, ...rest } = options;

    const finalHeaders: HeadersInit = {
        'Content-Type': 'application/json',
        'ngrok-skip-browser-warning': 'true',
        ...headers
    };

    const response = await fetch(`${API_URL}${endpoint}`, {
        ...rest,
        headers: finalHeaders,
        credentials: 'include',
        body: body === undefined ? undefined : JSON.stringify(body)
    });

    if (!response.ok) {
        const error = await parseErrorBody(response);
        throw new ApiError(error.code, response.status, error.message, error.fields);
    }

    const text = await response.text();
    return (text ? JSON.parse(text) : ({} as T)) as T;
}

async function parseErrorBody(
    response: Response
): Promise<{ code: string; message: string; fields?: Record<string, string> }> {
    // We always try JSON first because the v2 backend emits the structured
    // error shape. Legacy endpoints that still use `http.Error` (Spec §6.4
    // migration backlog) will fall through to a sensible default.
    try {
        const text = await response.text();
        if (text) {
            const parsed = JSON.parse(text) as BackendErrorBody;
            if (parsed?.error?.code) {
                return {
                    code: parsed.error.code,
                    message: parsed.error.message ?? response.statusText,
                    fields: parsed.error.fields
                };
            }
            // Legacy `http.Error` body: use it as a fallback message.
            return {
                code: response.status === 401 ? 'UNAUTHENTICATED' : 'INTERNAL_ERROR',
                message: text.trim() || response.statusText
            };
        }
    } catch {
        // Body was not JSON — fall through.
    }
    return {
        code: response.status === 401 ? 'UNAUTHENTICATED' : 'INTERNAL_ERROR',
        message: response.statusText
    };
}
