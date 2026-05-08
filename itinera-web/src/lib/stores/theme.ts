import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export type Theme = 'light' | 'dark' | 'system';

function createThemeStore() {
    // Start with a safe default for SSR
    const { subscribe, set } = writable<Theme>('system');

    return {
        subscribe,
        init: () => {
            if (!browser) return;
            const saved = localStorage.getItem('teren-theme') as Theme | null;
            if (saved) {
                set(saved);
                applyTheme(saved);
            } else {
                applyTheme('system');
            }
        },
        toggle: () => {
            if (!browser) return;
            const html = document.documentElement;
            const currentlyDark = html.classList.contains('dark');
            const next: Theme = currentlyDark ? 'light' : 'dark';
            
            set(next);
            localStorage.setItem('teren-theme', next);
            applyTheme(next);
        },
        setTheme: (theme: Theme) => {
            if (!browser) return;
            set(theme);
            localStorage.setItem('teren-theme', theme);
            applyTheme(theme);
        }
    };
}

function applyTheme(t: Theme) {
    if (!browser) return;
    const html = document.documentElement;
    
    if (t === 'dark') {
        html.classList.add('dark');
        html.classList.remove('light');
    } else if (t === 'light') {
        html.classList.add('light');
        html.classList.remove('dark');
    } else {
        // System
        const isDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
        html.classList.toggle('dark', isDark);
        html.classList.toggle('light', !isDark);
        localStorage.removeItem('teren-theme');
    }
}

export const themeStore = createThemeStore();
