-- Migration 012: Add country_code to places table
ALTER TABLE public.places 
ADD COLUMN IF NOT EXISTS country_code CHAR(2);

COMMENT ON COLUMN public.places.country_code IS 
'ISO 3166-1 alpha-2. Populated via backfill script.';

-- Backfill: populate_japan_cities.sql
-- 23 ciudades: 19 principales + 4 estaciones adicionales
UPDATE public.places
SET country_code = 'JP'
WHERE country_code IS NULL
  AND LOWER(city) IN (
    'tokyo', 'kyoto', 'osaka', 'nagano', 'fukuoka', 'hiroshima',
    'sapporo', 'yokohama', 'kobe', 'kumamoto', 'nara', 'kanazawa',
    'nagasaki', 'hakodate', 'matsumoto', 'miyajima', 'himeji',
    'sendai', 'nikko', 'shinjuku', 'shibuya', 'hakata', 'nagoya'
  );
