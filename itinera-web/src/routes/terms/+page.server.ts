import type { PageServerLoad } from './$types';
import { resolveLegalLocale, loadDoc } from '$lib/legal/docs';
import type { LegalLocale } from '$lib/legal/manifest';

/**
 * Legal page load: resolves which locale to render and returns the
 * rendered HTML + TOC.
 *
 * With `prerender = true` we cannot read `url.searchParams`
 * directly. Instead we use the `lang` entry-param declared in
 * `+page.ts#entries`: SvelteKit passes each generated entry's
 * params here. For canonical `/terms` (no entry) the param is
 * `undefined`, which `resolveLegalLocale` falls through to ES.
 */
export const prerender = true;

export const load: PageServerLoad = async ({ params, cookies }) => {
	const entryLang = (params as { lang?: string }).lang;
	const cookieLang = cookies.get('itinera_legal_lang');
	const locale: LegalLocale = resolveLegalLocale(
		entryLang ?? cookieLang,
		null,
	);

	const doc = await loadDoc('terms', locale);

	return {
		doc,
		localeShown: locale,
	};
};