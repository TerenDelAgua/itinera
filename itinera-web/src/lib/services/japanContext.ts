import dataset from '$lib/data/japan_context.json';
import { matchContext } from './contextMatcher';
import { getClimate, shouldShowClimate, type ClimateDisplay } from './climateService';

export interface RuleDisplay {
    id: string;
    priority: string;
    icon: string;
    severity: string;
}

export interface PhraseDisplay {
    id: string;
    romaji: string;
    ja: string;
    priority: string;
}

const etiquetteRules = dataset.etiquette_rules as any[];
const phrasesData = dataset.phrases as any[];

export function getRulesForActivity(name: string, notes: string | undefined, locale: string): RuleDisplay[] {
    const matchedTerms = matchContext(name, notes, locale);
    const results: RuleDisplay[] = [];
    const addedIds = new Set<string>();

    for (const termObj of matchedTerms) {
        for (const rule of etiquetteRules) {
            // ONLY match specific contexts for individual activities to avoid saturating them with general tips
            if (rule.contexts.includes(termObj.canonical_term)) {
                if (!addedIds.has(rule.id)) {
                    results.push({
                        id: rule.id,
                        priority: rule.priority,
                        icon: rule.icon,
                        severity: rule.severity
                    });
                    addedIds.add(rule.id);
                }
            }
        }
    }

    // Sort by severity critical -> high -> medium -> low
    const severityOrder: Record<string, number> = { critical: 0, high: 1, medium: 2, low: 3 };
    results.sort((a, b) => (severityOrder[a.severity] || 99) - (severityOrder[b.severity] || 99));

    return results;
}

export function getPlaceLevelRules(placeName: string, placeCity: string, locale: string): RuleDisplay[] {
    // Para Place header, pasamos el nombre del lugar y ciudad
    const rules = getRulesForActivity(`${placeName} ${placeCity}`, '', locale);
    
    // Append general rules (any / any_public) so general etiquette tips show up here
    const addedIds = new Set(rules.map(r => r.id));
    for (const rule of etiquetteRules) {
        if (rule.contexts.includes('any') || rule.contexts.includes('any_public')) {
            if (!addedIds.has(rule.id)) {
                rules.push({
                    id: rule.id,
                    priority: rule.priority,
                    icon: rule.icon,
                    severity: rule.severity
                });
                addedIds.add(rule.id);
            }
        }
    }
    return rules;
}

export function getPhrasesForContext(context: string, locale: string): PhraseDisplay[] {
    const filtered = phrasesData.filter(p => 
        (context === 'all' && (p.contexts.includes('general') || p.contexts.includes('any'))) ||
        (context !== 'all' && (p.contexts.includes(context) || p.contexts.includes('any')))
    );
    
    // Sort by priority
    const priorityOrder: Record<string, number> = { critical: 0, high: 1, medium: 2, low: 3 };
    filtered.sort((a, b) => (priorityOrder[a.priority] || 99) - (priorityOrder[b.priority] || 99));

    return filtered.map(p => ({
        id: p.id,
        romaji: p.romaji,
        ja: p.ja,
        priority: p.priority
    }));
}

export { getClimate, shouldShowClimate, type ClimateDisplay };
