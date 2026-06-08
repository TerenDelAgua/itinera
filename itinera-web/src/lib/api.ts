const API_URL = import.meta.env.VITE_API_URL;

if (!API_URL) {
    throw new Error("VITE_API_URL is not defined in the environment variables");
}

// Authentication is handled exclusively by HttpOnly cookies:
//   - `session_id` is issued by SessionMiddleware for every visitor (guest or
//     registered). The browser sends it on every request via credentials:
//     'include' below.
//   - `auth_token` is issued by Register/Login on top of the session cookie
//     once the guest upgrades. It carries the JWT and is HttpOnly, so it is
//     not readable from JavaScript and is therefore safe against XSS
//     exfiltration. AuthMiddleware reads it server-side.
//
// We previously mirrored the session id and JWT in localStorage to send them
// as X-Session-Id / Authorization headers. That path was XSS-readable and
// redundant once cookies are present, so it has been removed.

export async function apiFetch<T>(
    endpoint: string,
    options: RequestInit = {}
): Promise<T> {
    const headers: HeadersInit = {
        'Content-Type': 'application/json',
        'ngrok-skip-browser-warning': 'true',
        ...options.headers,
    };

    const response = await fetch(`${API_URL}${endpoint}`, {
        ...options,
        headers,
        credentials: 'include',
    });

    if (!response.ok) {
        let errorMessage = response.statusText;
        try {
            const bodyText = await response.text();
            if (bodyText) {
                // Remove trailing newlines from backend responses like `http.Error`
                errorMessage = bodyText.trim();
            }
        } catch (e) {}
        throw new Error(errorMessage);
    }

    const text = await response.text();
    return text ? JSON.parse(text) : {} as T;
}