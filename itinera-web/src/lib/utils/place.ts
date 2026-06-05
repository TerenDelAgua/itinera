// src/lib/utils/place.ts

// Lista completa: 19 ciudades principales + 4 estaciones adicionales + variantes en español
const JAPAN_CITIES = [
  // Ciudades principales (19)
  'tokyo', 'tokio', 'kyoto', 'kioto', 'osaka', 'nagano', 'fukuoka', 'hiroshima', 
  'sapporo', 'yokohama', 'kobe', 'kumamoto', 'nara', 'kanazawa', 
  'nagasaki', 'hakodate', 'matsumoto', 'miyajima', 'himeji', 
  'sendai', 'nikko',
  // Estaciones adicionales con Station Guides (4)
  'shinjuku', 'shibuya', 'hakata', 'nagoya'
];

/**
 * Determina si un place está en Japón.
 * Prioridad:
 *  1. country_code explícito en DB (fuente de verdad)
 *  2. Fallback por nombre de ciudad conocida (datos pre-migración)
 * 
 * @param place - Objeto con country_code y/o city
 * @returns true si el place está en Japón
 */
export function isJapanPlace(place: { 
  country_code?: string | null; 
  city?: string;
  name?: string;
} | null | undefined): boolean {
  if (!place) return false;

  // 1. Prioridad: campo explícito en DB
  if (place.country_code === 'JP') return true;

  // 2. Si es otro país explícito, no es Japón
  if (place.country_code && place.country_code !== 'JP') return false;

  // 3. Fallback: lista de ciudades conocidas en la prop "city"
  const city = place.city?.toLowerCase().trim() || '';
  if (JAPAN_CITIES.includes(city)) return true;

  // 4. Fallback extremo: buscar coincidencias en el nombre general
  const name = place.name?.toLowerCase().trim() || '';
  if (name && JAPAN_CITIES.some(c => name.includes(c))) return true;

  return false;
}
