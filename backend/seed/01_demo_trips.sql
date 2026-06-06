
-- Función helper para evitar duplicados
DO $$
DECLARE
    v_roma_id UUID;
    v_japan_id UUID;
    v_asia_id UUID;
BEGIN

-- ============================================
-- DEMO 1: Roma 3 días (City Break)
-- ============================================
INSERT INTO public.trips (id, name, description, start_date, end_date, base_currency, is_public_demo, session_id)
VALUES (
    '550e8400-e29b-41d4-a716-446655440001'::UUID,
    'inspiration.roma.title',
    'inspiration.roma.desc',
    '2026-06-15',
    '2026-06-17',
    'EUR',
    TRUE,
    'DEMO'
)
ON CONFLICT (id) DO UPDATE SET 
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    is_public_demo = TRUE;

v_roma_id := '550e8400-e29b-41d4-a716-446655440001'::UUID;

-- Lugares Roma
INSERT INTO public.places (id, trip_id, name, notes, lat, lon, start_date, end_date, default_expense_currency)
VALUES 
    ('550e8400-e29b-41d4-a716-446655440101'::UUID, v_roma_id, 'Coliseo y Foro Romano', 'Llegar antes de las 8:30 para evitar colas. Entrada combinada con Foro.', 41.8902, 12.4922, '2026-06-15', '2026-06-15', 'EUR'),
    ('550e8400-e29b-41d4-a716-446655440102'::UUID, v_roma_id, 'Vaticano y Museos', 'Reservar entrada online con antelación. Código de vestimenta estricto.', 41.9029, 12.4534, '2026-06-16', '2026-06-16', 'EUR'),
    ('550e8400-e29b-41d4-a716-446655440103'::UUID, v_roma_id, 'Trastevere', 'Barrio bohemio para cena. Pizzeria Da Enzo es obligatoria.', 41.8895, 12.4668, '2026-06-16', '2026-06-17', 'EUR')
ON CONFLICT (id) DO NOTHING;

-- Actividades Roma
INSERT INTO public.activities (id, trip_id, place_id, title, date, time, notes)
VALUES
    ('550e8400-e29b-41d4-a716-446655440201'::UUID, v_roma_id, '550e8400-e29b-41d4-a716-446655440101'::UUID, 'Visita guiada Coliseo', '2026-06-15', '08:30:00', 'Duración 3h. Incluye acceso a arena y underground.'),
    ('550e8400-e29b-41d4-a716-446655440202'::UUID, v_roma_id, '550e8400-e29b-41d4-a716-446655440101'::UUID, 'Paseo Foro Romano', '2026-06-15', '12:00:00', 'Almuerzo ligero en Mercati di Traiano.'),
    ('550e8400-e29b-41d4-a716-446655440203'::UUID, v_roma_id, '550e8400-e29b-41d4-a716-446655440102'::UUID, 'Museos Vaticanos', '2026-06-16', '09:00:00', 'Capilla Sixtina primero, luego Galerías. 4h mínimo.'),
    ('550e8400-e29b-41d4-a716-446655440204'::UUID, v_roma_id, '550e8400-e29b-41d4-a716-446655440103'::UUID, 'Cena en Trastevere', '2026-06-16', '20:00:00', 'Reservar mesa. Cacio e pepe y supplì.')
ON CONFLICT (id) DO NOTHING;

-- ============================================
-- DEMO 2: Japón 15 días (Tour de país)
-- ============================================
INSERT INTO public.trips (id, name, description, start_date, end_date, base_currency, is_public_demo, session_id)
VALUES (
    '550e8400-e29b-41d4-a716-446655440002'::UUID,
    'inspiration.japan.title',
    'inspiration.japan.desc',
    '2026-10-01',
    '2026-10-15',
    'JPY',
    TRUE,
    'DEMO'
)
ON CONFLICT (id) DO UPDATE SET 
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    is_public_demo = TRUE;

v_japan_id := '550e8400-e29b-41d4-a716-446655440002'::UUID;

-- Lugares Japón
INSERT INTO public.places (id, trip_id, name, notes, lat, lon, start_date, end_date, default_expense_currency)
VALUES 
    ('550e8400-e29b-41d4-a716-446655440111'::UUID, v_japan_id, 'Tokio (Shinjuku/Shibuya)', 'Base inicial. Tarjeta Suica obligatoria. Shibuya Crossing al atardecer.', 35.6895, 139.6917, '2026-10-01', '2026-10-04', 'JPY'),
    ('550e8400-e29b-41d4-a716-446655440112'::UUID, v_japan_id, 'Kioto (Gion/Fushimi)', 'Templos y geishas. Fushimi Inari al amanecer para evitar multitudes.', 35.0116, 135.7681, '2026-10-05', '2026-10-08', 'JPY'),
    ('550e8400-e29b-41d4-a716-446655440113'::UUID, v_japan_id, 'Osaka (Dotonbori)', 'Gastronomía y nightlife. Takoyaki en callejón principal.', 34.6937, 135.5023, '2026-10-09', '2026-10-10', 'JPY'),
    ('550e8400-e29b-41d4-a716-446655440114'::UUID, v_japan_id, 'Hiroshima y Miyajima', 'Memorial de la Paz y santuario flotante. Ferry incluido en JR Pass.', 34.3853, 132.4553, '2026-10-11', '2026-10-13', 'JPY'),
    ('550e8400-e29b-41d4-a716-446655440115'::UUID, v_japan_id, 'Fukuoka (Hakata)', 'Final en Kyushu. Ramen en yatai, mercado de pescado.', 33.5902, 130.4017, '2026-10-14', '2026-10-15', 'JPY')
ON CONFLICT (id) DO NOTHING;

-- Actividades Japón (muestra, 5 de 15 días)
INSERT INTO public.activities (id, trip_id, place_id, title, date, time, notes)
VALUES
    ('550e8400-e29b-41d4-a716-446655440301'::UUID, v_japan_id, '550e8400-e29b-41d4-a716-446655440111'::UUID, 'Llegada Narita/Haneda → Shinjuku', '2026-10-01', '14:00:00', 'Narita Express o Limousine Bus. Check-in hotel después de 15:00.'),
    ('550e8400-e29b-41d4-a716-446655440302'::UUID, v_japan_id, '550e8400-e29b-41d4-a716-446655440111'::UUID, 'Shibuya Crossing + Hachiko', '2026-10-01', '17:30:00', 'Atardecer en Shibuya Sky (reservar con antelación).'),
    ('550e8400-e29b-41d4-a716-446655440303'::UUID, v_japan_id, '550e8400-e29b-41d4-a716-446655440112'::UUID, 'Fushimi Inari Taisha', '2026-10-05', '06:00:00', 'Sendero completo 2-3h. Llevar agua. Etiqueta: no tocar los torii.'),
    ('550e8400-e29b-41d4-a716-446655440304'::UUID, v_japan_id, '550e8400-e29b-41d4-a716-446655440113'::UUID, 'Dotonbori Night Walk', '2026-10-09', '20:00:00', 'Neón, comida callejera, ambiente. Cuidado con bicicletas.'),
    ('550e8400-e29b-41d4-a716-446655440305'::UUID, v_japan_id, '550e8400-e29b-41d4-a716-446655440115'::UUID, 'Ramen en Yatai de Nakasu', '2026-10-14', '21:00:00', 'Experiencia local auténtica. Llegar antes de las 22:00.')
ON CONFLICT (id) DO NOTHING;

-- ============================================
-- DEMO 3: Asia 2 meses (Mochilero/Nómada)
-- ============================================
INSERT INTO public.trips (id, name, description, start_date, end_date, base_currency, is_public_demo, session_id)
VALUES (
    '550e8400-e29b-41d4-a716-446655440003'::UUID,
    'inspiration.asia.title',
    'inspiration.asia.desc',
    '2026-11-01',
    '2026-12-30',
    'USD',
    TRUE,
    'DEMO'
)
ON CONFLICT (id) DO UPDATE SET 
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    is_public_demo = TRUE;

v_asia_id := '550e8400-e29b-41d4-a716-446655440003'::UUID;

-- Lugares Asia
INSERT INTO public.places (id, trip_id, name, notes, lat, lon, start_date, end_date, default_expense_currency)
VALUES 
    ('550e8400-e29b-41d4-a716-446655440121'::UUID, v_asia_id, 'Bangkok, Tailandia', 'Base inicial. Visa 60 días. Coworkings en Sukhumvit.', 13.7563, 100.5018, '2026-11-01', '2026-11-15', 'THB'),
    ('550e8400-e29b-41d4-a716-446655440122'::UUID, v_asia_id, 'Chiang Mai, Tailandia', 'Digital nomad hub. Café culture y montañas.', 18.7883, 98.9853, '2026-11-16', '2026-11-30', 'THB'),
    ('550e8400-e29b-41d4-a716-446655440123'::UUID, v_asia_id, 'Ho Chi Minh, Vietnam', 'Caos organizado. Phở para desayunar. Visado e-visa online.', 10.8231, 106.6297, '2026-12-01', '2026-12-15', 'VND'),
    ('550e8400-e29b-41d4-a716-446655440124'::UUID, v_asia_id, 'Bali (Canggu), Indonesia', 'Final en paraíso surfero. Coworkings en la playa.', -8.6478, 115.1385, '2026-12-16', '2026-12-30', 'IDR')
ON CONFLICT (id) DO NOTHING;

-- Actividades Asia
INSERT INTO public.activities (id, trip_id, place_id, title, date, time, notes)
VALUES
    ('550e8400-e29b-41d4-a716-446655440401'::UUID, v_asia_id, '550e8400-e29b-41d4-a716-446655440121'::UUID, 'Setup base: SIM + Coworking', '2026-11-01', '10:00:00', 'AIS o True SIM en aeropuerto. Hubba o Glowfish para coworking.'),
    ('550e8400-e29b-41d4-a716-446655440402'::UUID, v_asia_id, '550e8400-e29b-41d4-a716-446655440122'::UUID, 'Doi Suthep al amanecer', '2026-11-20', '05:30:00', 'Songthaew compartido desde Chiang Mai. 300 escalones. Vistas.'),
    ('550e8400-e29b-41d4-a716-446655440403'::UUID, v_asia_id, '550e8400-e29b-41d4-a716-446655440123'::UUID, 'Food tour motorbike District 1', '2026-12-03', '18:00:00', 'Bánh mì, phở, cà phê sữa đá. Conductor local recomendado.'),
    ('550e8400-e29b-41d4-a716-446655440404'::UUID, v_asia_id, '550e8400-e29b-41d4-a716-446655440124'::UUID, 'Surf en Batu Bolong', '2026-12-20', '06:00:00', 'Tabla 50k IDR/día. Olas para principiantes. Sunset en La Brisa.')
ON CONFLICT (id) DO NOTHING;

-- ============================================
-- Actualizar contador público inicial
-- ============================================
UPDATE public.public_stats 
SET total_trips_count = 3, last_updated = NOW()
WHERE id = 1;

END $$;