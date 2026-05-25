-- Migration 011: Add city to places table
ALTER TABLE public.places 
ADD COLUMN IF NOT EXISTS city VARCHAR;

COMMENT ON COLUMN public.places.city IS 
'Ciudad del lugar. Usado por Station Guide y JR Optimizer para matching preciso. Población manual en F1, IA en F2.';
