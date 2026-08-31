import type { PageServerLoad } from './$types';
import { resolveLegalLocale, loadDoc } from '$lib/legal/docs';
import type { LegalLocale } from '$lib/legal/manifest';

export const prerender = true;

export const load: PageServerLoad = async ({ params, cookies }) => {
	const entryLang = (params as { lang?: string }).lang;
	const cookieLang = cookies.get('itinera_legal_lang');
	const locale: LegalLocale = resolveLegalLocale(
		entryLang ?? cookieLang,
		null,
	);

	const doc = await loadDoc('privacy', locale);

	return {
		doc,
		localeShown: locale,
	};
};