import dataset from '$lib/data/japan_context.json';

export interface KeywordRule {
    keywords: { es: string[], en: string[], ja: string[] };
    canonical_term: string;
    priority: number;
}

const rules: KeywordRule[] = dataset.keyword_rules as KeywordRule[];

export function matchContext(activityName: string, activityNotes: string | undefined, locale: 'es' | 'en' | 'ja' | string): { canonical_term: string, priority: number }[] {
    const text = `${activityName} ${activityNotes || ''}`.toLowerCase();
    const matched = new Map<string, number>();

    for (const rule of rules) {
        // Fallback a 'en' si el locale no es uno de los soportados
        const safeLocale = (locale === 'es' || locale === 'en' || locale === 'ja') ? locale : 'en';
        const keywords = rule.keywords[safeLocale] || rule.keywords['en'];
        
        for (const kw of keywords) {
            if (text.includes(kw.toLowerCase())) {
                if (!matched.has(rule.canonical_term) || matched.get(rule.canonical_term)! > rule.priority) {
                    matched.set(rule.canonical_term, rule.priority);
                }
                break; // Found a match for this rule, no need to check other keywords of the same rule
            }
        }
    }

    const result = Array.from(matched.entries()).map(([term, priority]) => ({ canonical_term: term, priority }));
    result.sort((a, b) => a.priority - b.priority);
    return result;
}
