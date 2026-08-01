import { writable, derived } from 'svelte/store';
import en from './en.json';
import es from './es.json';
import ja from './ja.json';
import id from './id.json';

type Locale = 'en' | 'es' | 'ja' | 'id';
type Messages = typeof en;

type Leaves<T, P extends string = ""> =
    T extends string
        ? P
        : { [K in keyof T & string]: Leaves<T[K], `${P}${P extends "" ? "" : "."}${K}`> }[keyof T & string];

type TranslationKeys<T> = Leaves<T>;

export const locale = writable<Locale>('en');

const messages: Record<Locale, Messages> = { en, es, ja, id };

export const t = derived(locale, ($locale) => {
    return (key: TranslationKeys<Messages>, vars?: Record<string, string | number>) => {
        const keys = (key as string).split('.');
        let text: unknown = messages[$locale];

        for (const k of keys) {
            if (text && typeof text === 'object' && k in text) {
                text = (text as Record<string, unknown>)[k];
            } else {
                return key as string;
            }
        }

        if (typeof text !== 'string' || !vars) {
            return text as string;
        }

        return Object.entries(vars).reduce((result, [name, value]) => {
            return result.replaceAll(`{${name}}`, String(value));
        }, text);
    };
});