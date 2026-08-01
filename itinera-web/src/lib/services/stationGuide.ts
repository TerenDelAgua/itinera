import stationData from '$lib/data/station_guides.json';
import { isJapanPlace } from '$lib/utils/place';

// ============================================================
// INTERFACES
// ============================================================

export interface StationGuide {
  id: string;
  name: string;
  name_ja: string;
  name_ja_romaji?: string;
  city: string;
  region: string;
  complexity: 'low' | 'medium' | 'high' | 'extreme';
  daily_passengers_millions?: number;
  operators: Operator[];
  exits: Record<string, Exit>;
  key_tips: Tip[];
  svg_map: string;
  svg_map_simplified: string;
}

export interface Operator {
  name: string;
  name_ja: string;
  lines: Line[];
}

export interface Line {
  name: string;
  name_ja: string;
  color: string;
  platforms: string[];
  directions: string[];
}

export interface Exit {
  name: string;
  name_es?: string;
  name_ja: string;
  landmarks: string[];
  best_for: string[];
  time_to_surface: string;
  tips: string[];
  tips_es?: string[];
}

export interface Tip {
  category: 'navigation' | 'practical' | 'safety' | 'cultural';
  priority: 'critical' | 'high' | 'medium' | 'low';
  text: string;
  text_ja?: string;
}

export interface Place {
  id: string;
  name: string;
  city?: string;
  visitDate?: string;
  endDate?: string;
  notes?: string;
  country_code?: string | null;
}

export interface GuideDecision {
  show: boolean;
  city: string | null;
}

// ============================================================
// DATASET
// ============================================================

const stations: StationGuide[] = stationData.stations as unknown as StationGuide[];

// Mapa de normalización para búsqueda O(1)
const cityToStation = new Map<string, StationGuide>();
stations.forEach(s => {
  const cityLower = s.city.toLowerCase();
  cityToStation.set(cityLower, s);

  // Soporte robusto de traducciones/variantes de ciudades del español (Tokio/Kioto)
  if (cityLower === 'tokyo') {
    cityToStation.set('tokio', s);
  } else if (cityLower === 'kyoto') {
    cityToStation.set('kioto', s);
  }

  cityToStation.set(s.id.toLowerCase(), s);
  cityToStation.set(s.name.toLowerCase(), s);
  // Also index the Japanese name to allow robust fallbacks
  cityToStation.set(s.name_ja.toLowerCase(), s);
});

// ============================================================
// BÚSQUEDA
// ============================================================

/**
 * Encuentra un Station Guide por nombre de ciudad, id o nombre de estación.
 * Normaliza el input (minúsculas, trim) para matching flexible.
 */
export function findStationGuide(cityName: string): StationGuide | null {
  if (!cityName) return null;
  const normalized = cityName.toLowerCase().trim();
  return cityToStation.get(normalized) || null;
}

/**
 * Encuentra un Station Guide buscando primero coincidencia en el nombre del lugar,
 * y luego en el nombre de la ciudad como fallback.
 * Esto permite dar cobertura a múltiples estaciones dentro de la misma ciudad (ej. Tokyo Station, Shinjuku Station, Shibuya Station en Tokyo).
 */
export function matchStationGuide(place: { name: string; city?: string; country_code?: string | null } | null | undefined): StationGuide | null {
  if (!place || !isJapanPlace(place)) {
    return null;
  }

  const placeName = place.name;
  const cityName = place.city;
  const normalizedName = placeName ? placeName.toLowerCase().trim() : "";
  const normalizedCity = cityName ? cityName.toLowerCase().trim() : "";

  // Normalización de variantes español/inglés para búsquedas (Tokio -> Tokyo, Kioto -> Kyoto)
  const cleanName = normalizedName.replace("tokio", "tokyo").replace("kioto", "kyoto");
  const cleanCity = normalizedCity.replace("tokio", "tokyo").replace("kioto", "kyoto");

  // 1. Intentar buscar por coincidencia en el nombre del lugar (ej. "Shinjuku Station" o "Shibuya")
  for (const s of stations) {
    const sId = s.id.toLowerCase();
    if (cleanName.includes(sId)) {
      return s;
    }
  }

  // 1.5. Si la ciudad especificada coincide exactamente con el ID de alguna estación (ej. "shibuya", "shinjuku", "kyoto", "hakata"),
  // le damos prioridad antes de caer en coincidencias parciales o difusas del nombre del lugar.
  if (cleanCity) {
    const directGuide = stations.find(s => s.id.toLowerCase() === cleanCity);
    if (directGuide) {
      return directGuide;
    }
  }

  // 2. Coincidencia parcial en el nombre del lugar (ej. "Estación de Tokio" -> "tokyo_station")
  for (const s of stations) {
    const sName = s.name.toLowerCase().replace(" station", ""); // remove station suffix for flexible match
    const normStationName = sName.replace("tokio", "tokyo").replace("kioto", "kyoto");
    if (sName.length > 2 && (cleanName.includes(normStationName) || normStationName.includes(cleanName))) {
      return s;
    }
  }

  // 3. Fallback a la ciudad
  if (cleanCity) {
    // Si la ciudad es Tokyo, preferimos Tokyo Station como fallback por defecto
    if (cleanCity === 'tokyo') {
      return findStationGuide('tokyo_station');
    }
    return findStationGuide(cleanCity);
  }

  return null;
}

// ============================================================
// FILTRADO Y ORDENACIÓN
// ============================================================

const PRIORITY_ORDER: Record<Tip['priority'], number> = {
  critical: 0,
  high: 1,
  medium: 2,
  low: 3
};

/**
 * Ordena tips por prioridad (critical → high → medium → low)
 */
export function getTipsByPriority(guide: StationGuide): Tip[] {
  return [...guide.key_tips].sort(
    (a, b) => PRIORITY_ORDER[a.priority] - PRIORITY_ORDER[b.priority]
  );
}

/**
 * Filtra salidas relevantes al contexto del viajero.
 * Si no hay contexto, devuelve todas.
 */
export function getRelevantExits(
  guide: StationGuide,
  context?: string[]
): Exit[] {
  const allExits = Object.values(guide.exits);
  if (!context || context.length === 0) return allExits;
  return allExits.filter(exit =>
    exit.best_for.some(bf => context.includes(bf))
  );
}

// ============================================================
// I18N HELPER
// ============================================================

/**
 * Selecciona el campo apropiado según el locale actual.
 * Fallback a inglés si no hay traducción en JA.
 */
export function localize<T extends string>(
  obj: Record<string, T | undefined>,
  keyJa: string,
  keyEn: string,
  locale: string
): T | undefined {
  if (locale === 'ja' && obj[keyJa]) {
    return obj[keyJa] as T;
  }
  return obj[keyEn] as T;
}

// ============================================================
// DECISIÓN DE RENDERIZADO — EVITA BADGES DUPLICADOS
// ============================================================

/**
 * Computa qué places deben mostrar Station Guide.
 *
 * Reglas:
 * 1. El place debe tener campo `city` definido
 * 2. La ciudad debe tener un Station Guide en el dataset
 * 3. La complejidad del guide debe ser 'high' o 'extreme'
 * 4. Solo la PRIMERA aparición de cada ciudad muestra el badge
 *    (evita duplicados si el viajero vuelve a la misma ciudad)
 *
 * Uso en Svelte:
 *   let decisions = $derived(computeGuideDecisions(places));
 *   {#if decisions[index]?.show}
 *     <StationGuide cityName={decisions[index].city} />
 *   {/if}
 */
export function computeGuideDecisions(places: Place[]): GuideDecision[] {
  const seenCities = new Set<string>();

  return places.map(place => {
    if (!place.city || !isJapanPlace(place)) {
      return { show: false, city: null };
    }

    const city = place.city.toLowerCase().trim();
    const guide = findStationGuide(city);

    // Permitimos medium, high o extreme en F1 para dar cobertura a Hakata también!
    // Hakata es complexity medium pero está en las 3 del MVP.
    if (!guide || (guide.complexity !== 'medium' && guide.complexity !== 'high' && guide.complexity !== 'extreme')) {
      return { show: false, city: null };
    }

    if (seenCities.has(city)) {
      return { show: false, city: null };
    }

    seenCities.add(city);
    return { show: true, city: place.city };
  });
}

// ============================================================
// UTILIDADES
// ============================================================

/**
 * Lista todas las ciudades cubiertas por Station Guides.
 */
export function getCoveredCities(): string[] {
  return [...new Set(stations.map(s => s.city.toLowerCase()))];
}

/**
 * Verifica si una ciudad tiene Station Guide disponible.
 */
export function isCityCovered(cityName: string): boolean {
  return findStationGuide(cityName) !== null;
}

/**
 * Obtiene el color de complejidad para badges visuales.
 */
export function getComplexityColor(complexity: StationGuide['complexity']): string {
  const colors: Record<StationGuide['complexity'], string> = {
    low: '#22C55E',
    medium: '#EAB308',
    high: '#F97316',
    extreme: '#DC2626'
  };
  return colors[complexity];
}
