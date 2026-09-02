import type { PageServerLoad } from './$types';
import { resolveLegalLocale, loadDoc } from '$lib/legal/docs';
import { SUPPORTED_LOCALES, type LegalLocale } from '$lib/legal/manifest';

export const prerender = true;

export const load: PageServerLoad = async ({ params, cookies }) => {
	const entryLang = params.lang;
	const cookieLang = cookies.get('itinera_legal_lang');
	const locale: LegalLocale = resolveLegalLocale(
		entryLang ?? cookieLang,
		null,
	);
	const safeLocale: LegalLocale = (SUPPORTED_LOCALES as readonly string[]).includes(
		locale,
	)
		? locale
		: 'es';

	const doc = await loadDoc('privacy', safeLocale);

	return {
		doc,
		localeShown: safeLocale,
	};
};