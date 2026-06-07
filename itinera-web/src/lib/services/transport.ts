import routesDataRaw from '../data/jr_routes.json';
import type { Place } from '../types/Place';

// Define types that match the JSON structure
export interface RouteData {
  id: string;
  from: string;
  to: string;
  fare: number;
  time_min: number;
  shinkansen: boolean;
  jr_covered: boolean;
  line: string;
  notes?: string;
  alternatives?: {
    operator: string;
    line?: string;
    fare: number;
    time_min: number;
    jr_covered: boolean;
  }[];
}

export interface RoutesData {
  version: string;
  currency_base: string;
  last_updated: string;
  passes: {
    '7_day': { price: number; name: string; daily_cost: number };
    '14_day': { price: number; name: string; daily_cost: number };
    '21_day': { price: number; name: string; daily_cost: number };
  };
  routes: RouteData[];
}

export const routesData = routesDataRaw as RoutesData;

export interface RouteMatch {
  from: string;
  to: string;
  fare: number;
  timeMin: number;
  line: string;
  shinkansen: boolean;
  noteKey?: string;
  noteParams?: Record<string, string>;
  notes?: string; // from the static data
}

export interface TransportResult {
  totalFare: number;
  routesFound: RouteMatch[];
  routesMissing: string[];
  recommendedPass: '7_day' | '14_day' | '21_day' | null;
  passPrice: number;
  savings: number;
  breakEvenDays: number | null;
  honestMessageKey: string;
  honestMessageParams?: Record<string, string>;
  algorithmUsed: 'date-based' | 'sequential-fallback';
}

/**
 * Normalizes city names for comparison
 */
function normalizeName(name: string): string {
  if (!name) return '';
  return name.toLowerCase().trim()
    .replace(/airport/g, '')
    .replace(/ station/g, '')
    .replace(/tokio/g, 'tokyo')
    .replace(/kioto/g, 'kyoto')
    .replace(/fukuoka/g, 'hakata')
    .trim();
}

/**
 * Finds a route between two cities, regardless of direction
 */
function findRoute(from: string, to: string): RouteData | undefined {
  const normFrom = normalizeName(from);
  const normTo = normalizeName(to);

  return routesData.routes.find(r => {
    const rFrom = normalizeName(r.from);
    const rTo = normalizeName(r.to);

    // El dataset tiene nombres limpios (ej. "tokyo"). 
    // El usuario puede poner "tokyo (shinjuku)", así que usamos includes.
    const matchForward = normFrom.includes(rFrom) && normTo.includes(rTo);
    const matchBackward = normFrom.includes(rTo) && normTo.includes(rFrom);

    return matchForward || matchBackward;
  });
}

/**
 * Para un place destino, encuentra el place de origen real
 * mirando hacia atrás en el itinerario por fechas.
 *
 * CAPA 1: Match exacto (termina el mismo día que empieza el destino)
 * CAPA 2: Match más cercano anterior (termina antes, más cerca posible)
 * CAPA 3: Fallback secuencial (anterior en la lista)
 */
export function findOriginPlace(
  targetPlace: Place,
  previousPlaces: Place[]
): { place: Place | null; method: 'date-exact' | 'date-closest' | 'sequential-fallback' } {
  // CAPA 3 (fallback): Si el destino no tiene fecha, no podemos hacer nada inteligente
  if (!targetPlace.start_date) {
    const fallback = previousPlaces[previousPlaces.length - 1] || null;
    return { place: fallback, method: 'sequential-fallback' };
  }

  const targetStart = new Date(targetPlace.start_date);
  targetStart.setHours(0, 0, 0, 0);

  // CAPA 1: Match exacto
  const exactMatch = previousPlaces.find(p => {
    if (!p.end_date) return false;
    const end = new Date(p.end_date);
    end.setHours(0, 0, 0, 0);
    return end.getTime() === targetStart.getTime();
  });

  if (exactMatch) {
    return { place: exactMatch, method: 'date-exact' };
  }

  // CAPA 2: Match más cercano anterior
  const candidates = previousPlaces.filter(p => {
    if (!p.end_date) return false;
    const end = new Date(p.end_date);
    end.setHours(0, 0, 0, 0);
    return end < targetStart;
  });

  if (candidates.length > 0) {
    const closest = candidates.sort((a, b) => {
      const diffA = Math.abs(new Date(a.end_date!).getTime() - targetStart.getTime());
      const diffB = Math.abs(new Date(b.end_date!).getTime() - targetStart.getTime());
      return diffA - diffB;
    })[0];
    return { place: closest, method: 'date-closest' };
  }

  // CAPA 3: Fallback secuencial
  const fallback = previousPlaces[previousPlaces.length - 1] || null;
  return { place: fallback, method: 'sequential-fallback' };
}

export function calculateTransport(
  places: Place[],
  tripDurationDays: number
): TransportResult | null {
  if (places.length < 2) return null;

  let totalFare = 0;
  const routesFound: RouteMatch[] = [];
  const routesMissing: string[] = [];
  let dateBasedRoutes = 0;
  let sequentialRoutes = 0;

  let firstRouteDate: number | null = null;
  let lastRouteDate: number | null = null;

  for (let i = 1; i < places.length; i++) {
    const targetPlace = places[i];
    const previousPlaces = places.slice(0, i);

    const { place: originPlace, method } = findOriginPlace(targetPlace, previousPlaces);

    if (method === 'date-exact' || method === 'date-closest') {
      dateBasedRoutes++;
    } else {
      sequentialRoutes++;
    }

    const originCity = originPlace
      ? originPlace.name
      : places[i - 1].name;
    const destCity = targetPlace.name;

    const route = findRoute(originCity, destCity);

    if (route) {
      totalFare += route.fare;

      const routeDateStr = targetPlace.start_date || originPlace?.end_date;
      if (routeDateStr) {
        const time = new Date(routeDateStr).getTime();
        if (firstRouteDate === null || time < firstRouteDate) firstRouteDate = time;
        if (lastRouteDate === null || time > lastRouteDate) lastRouteDate = time;
      }

      let noteKey: string | undefined;
      let noteParams: Record<string, string> | undefined;

      if (method === 'sequential-fallback') {
        noteKey = 'transport.fallback_note';
      } else if (method === 'date-closest') {
        noteKey = 'transport.closest_note';
        noteParams = { origin: originPlace?.name || originCity };
      }

      routesFound.push({
        // We want to show the direction the user is actually traveling
        from: originCity,
        to: destCity,
        fare: route.fare,
        timeMin: route.time_min,
        line: route.line,
        shinkansen: route.shinkansen,
        noteKey,
        noteParams,
        notes: route.notes
      });
    } else {
      routesMissing.push(`${originCity} → ${destCity}`);
    }
  }

  // Determinar pase recomendado según duración efectiva del transporte
  let effectiveTransportDays = tripDurationDays;
  if (firstRouteDate !== null && lastRouteDate !== null) {
    const diff = lastRouteDate - firstRouteDate;
    effectiveTransportDays = Math.ceil(diff / (1000 * 3600 * 24)) + 1;
  }

  let recommendedPass: TransportResult['recommendedPass'] = null;
  let passPrice = 0;

  if (effectiveTransportDays <= 7) {
    recommendedPass = '7_day';
    passPrice = routesData.passes['7_day'].price;
  } else if (effectiveTransportDays <= 14) {
    recommendedPass = '14_day';
    passPrice = routesData.passes['14_day'].price;
  } else {
    recommendedPass = '21_day';
    passPrice = routesData.passes['21_day'].price;
  }

  const savings = totalFare - passPrice;
  const breakEvenDays = totalFare > 0 && routesFound.length > 0
    ? Math.ceil(passPrice / (totalFare / routesFound.length))
    : null;

  let honestMessageKey = '';
  let honestMessageParams: Record<string, string> | undefined;

  if (routesFound.length === 0) {
    honestMessageKey = 'transport.no_routes';
  } else if (savings > 10000 && recommendedPass) {
    honestMessageKey = 'transport.significant_savings';
    honestMessageParams = { passName: routesData.passes[recommendedPass].name };
  } else if (savings > 0) {
    honestMessageKey = 'transport.modest_savings';
  } else {
    honestMessageKey = 'transport.cheaper_individual';
  }

  const algorithmUsed: TransportResult['algorithmUsed'] =
    dateBasedRoutes >= sequentialRoutes && places.some(p => p.start_date) ? 'date-based' : 'sequential-fallback';

  return {
    totalFare,
    routesFound,
    routesMissing,
    recommendedPass,
    passPrice,
    savings,
    breakEvenDays,
    honestMessageKey,
    honestMessageParams,
    algorithmUsed
  };
}
