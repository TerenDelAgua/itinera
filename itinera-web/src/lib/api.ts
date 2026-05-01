const API_URL = import.meta.env.VITE_API_URL;

if (!API_URL) {
    throw new Error("VITE_API_URL is not defined in the environment variables");
}

export async function apiFetch<T>(
    endpoint: string,
    options: RequestInit = {}
): Promise<T> {
    const token = typeof window !== 'undefined' ? localStorage.getItem('session_token') : null;
    const headers: HeadersInit = {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...options.headers,
    };

    const response = await fetch(`${API_URL}${endpoint}`, {
        ...options,
        headers,
        credentials: 'include' //To send the cookie automatically
    });

    if (!response.ok) {
        if (response.status === 401 || response.status === 403) {
            localStorage.removeItem('session_token');
        }
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