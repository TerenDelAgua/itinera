import type { PageServerLoad } from './$types';
import { resolveLegalLocale, loadDoc } from '$lib/legal/docs';
import { SUPPORTED_LOCALES, type LegalLocale } from '$lib/legal/manifest';

/**
 * Locale variant of /terms. The `[lang]` segment IS the locale —
 * we validate against SUPPORTED_LOCALES so a malformed URL like
 * `/terms/foo` falls back to ES instead of 404-ing.
 */
export const prerender = true;

export const load: PageServerLoad = async ({ params, cookies }) => {
	const entryLang = params.lang;
	const cookieLang = cookies.get('itinera_legal_lang');
	const locale: LegalLocale = resolveLegalLocale(
		entryLang ?? cookieLang,
		null,
	);
	// Defensive: if resolveLegalLocale somehow returned something
	// that isn't a supported locale (it never should), coerce to ES.
	const safeLocale: LegalLocale = (SUPPORTED_LOCALES as readonly string[]).includes(
		locale,
	)
		? locale
		: 'es';

	const doc = await loadDoc('terms', safeLocale);

	return {
		doc,
		localeShown: safeLocale,
	};
};