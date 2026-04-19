import { writable, derived } from 'svelte/store';
import en from './en.json';
import es from './es.json';

type Locale = 'en' | 'es';
type Messages = typeof en;

// TypeScript helper to extract deep paths (e.g., "dashboard.subtitle")
type TranslationKeys<T> = {
    [K in keyof T]: T[K] extends string
    ? K
    : T[K] extends Record<string, any>
    ? `${K & string}.${TranslationKeys<T[K]> & string}`
    : never;
}[keyof T];

export const locale = writable<Locale>('en');

const messages: Record<Locale, Messages> = { en, es };

export const t = derived(locale, ($locale) => {
    return (key: TranslationKeys<Messages>) => {
        const keys = (key as string).split('.');
        let text = messages[$locale] as any;

        for (const k of keys) {
            if (text && text[k] !== undefined) text = text[k];
            else return key as string;
        }
        return text as string;
    };
});