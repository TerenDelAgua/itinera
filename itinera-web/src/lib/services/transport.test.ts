import { calculateTransport } from './transport';
import type { Place } from '../types/Place';
import { describe, it, expect } from 'vitest';

describe('calculateTransport - Algoritmo híbrido', () => {

  it('detecta excursión: Tokyo → Kamakura → Nagoya', () => {
    const places: Place[] = [
      { id: '1', trip_id: '1', notes: '', lat: null, lon: null, name: 'Tokyo', start_date: '2026-06-01', end_date: '2026-06-03' },
      { id: '2', trip_id: '1', notes: '', lat: null, lon: null, name: 'Kamakura', start_date: '2026-06-02', end_date: '2026-06-02' },
      { id: '3', trip_id: '1', notes: '', lat: null, lon: null, name: 'Nagoya', start_date: '2026-06-03', end_date: '2026-06-05' }
    ];

    const result = calculateTransport(places, 5);

    expect(result).not.toBeNull();
    expect(result!.routesFound).toHaveLength(2);

    // Ruta 1: Tokyo → Kamakura
    expect(result!.routesFound[0].from).toBe('Tokyo');
    expect(result!.routesFound[0].to).toBe('Kamakura');

    // Ruta 2: Tokyo → Nagoya (NO Kamakura → Nagoya)
    expect(result!.routesFound[1].from).toBe('Tokyo');
    expect(result!.routesFound[1].to).toBe('Nagoya');

    expect(result!.algorithmUsed).toBe('date-based');
  });

  it('detecta excursión desde Kyoto a Nara', () => {
    const places: Place[] = [
      { id: '1', trip_id: '1', notes: '', lat: null, lon: null, name: 'Kyoto', start_date: '2026-06-01', end_date: '2026-06-04' },
      { id: '2', trip_id: '1', notes: '', lat: null, lon: null, name: 'Nara', start_date: '2026-06-03', end_date: '2026-06-03' },
      { id: '3', trip_id: '1', notes: '', lat: null, lon: null, name: 'Tokyo', start_date: '2026-06-04', end_date: '2026-06-06' }
    ];

    const result = calculateTransport(places, 6);

    // Nara → Tokyo debería ser Kyoto → Tokyo
    expect(result!.routesFound[1].from).toBe('Kyoto');
    expect(result!.routesFound[1].to).toBe('Tokyo');
  });

  it('fallback secuencial cuando no hay fechas', () => {
    const places: Place[] = [
      { id: '1', trip_id: '1', notes: '', lat: null, lon: null, name: 'Tokyo' },
      { id: '2', trip_id: '1', notes: '', lat: null, lon: null, name: 'Kyoto' }
    ];

    const result = calculateTransport(places, 5);

    expect(result!.algorithmUsed).toBe('sequential-fallback');
    expect(result!.routesFound[0].from).toBe('Tokyo');
    expect(result!.routesFound[0].to).toBe('Kyoto');
  });

  it('match más cercano cuando no hay exacto', () => {
    const places: Place[] = [
      { id: '1', trip_id: '1', notes: '', lat: null, lon: null, name: 'Tokyo', start_date: '2026-06-01', end_date: '2026-06-05' },
      { id: '2', trip_id: '1', notes: '', lat: null, lon: null, name: 'Kyoto', start_date: '2026-06-02', end_date: '2026-06-02' },
      { id: '3', trip_id: '1', notes: '', lat: null, lon: null, name: 'Osaka', start_date: '2026-06-06', end_date: '2026-06-08' }
    ];

    const result = calculateTransport(places, 8);

    // Osaka empieza 6, nadie termina 6 exacto
    // Tokyo termina 5 (más cercano anterior)
    expect(result!.routesFound[1].from).toBe('Tokyo');
    expect(result!.routesFound[1].to).toBe('Osaka');
    expect(result!.routesFound[1].noteKey).toContain('closest_note');
  });

  it('no sobreestima cuando la excursión está en medio', () => {
    const withExcursion = calculateTransport([
      { id: '1', trip_id: '1', notes: '', lat: null, lon: null, name: 'Tokyo', start_date: '2026-06-01', end_date: '2026-06-05' },
      { id: '2', trip_id: '1', notes: '', lat: null, lon: null, name: 'Kamakura', start_date: '2026-06-03', end_date: '2026-06-03' },
      { id: '3', trip_id: '1', notes: '', lat: null, lon: null, name: 'Nagoya', start_date: '2026-06-05', end_date: '2026-06-07' }
    ], 7);

    const withoutExcursion = calculateTransport([
      { id: '1', trip_id: '1', notes: '', lat: null, lon: null, name: 'Tokyo', start_date: '2026-06-01', end_date: '2026-06-05' },
      { id: '2', trip_id: '1', notes: '', lat: null, lon: null, name: 'Nagoya', start_date: '2026-06-05', end_date: '2026-06-07' }
    ], 7);

    expect(withExcursion!.totalFare).toBeGreaterThan(withoutExcursion!.totalFare);
    expect(withExcursion!.totalFare - withoutExcursion!.totalFare).toBeLessThan(3000);
  });

  it('honestidad: Osaka-Kobe no renta JR Pass', () => {
    const places: Place[] = [
      { id: '1', trip_id: '1', notes: '', lat: null, lon: null, name: 'Osaka', start_date: '2026-06-01', end_date: '2026-06-03' },
      { id: '2', trip_id: '1', notes: '', lat: null, lon: null, name: 'Kobe', start_date: '2026-06-03', end_date: '2026-06-03' }
    ];

    const result = calculateTransport(places, 3);

    expect(result!.savings).toBeLessThan(0);
    expect(result!.honestMessageKey).toContain('cheaper_individual');
  });

  it('detecta correctamente todas las rutas y calcula el JR span para recomendar pase de 14 días en un viaje de 17 días', () => {
    const places: Place[] = [
      { id: '1', trip_id: '1', notes: '', lat: null, lon: null, name: 'Tokio', start_date: '2026-10-01', end_date: '2026-10-04' },
      { id: '2', trip_id: '1', notes: '', lat: null, lon: null, name: 'Kioto', start_date: '2026-10-05', end_date: '2026-10-08' },
      { id: '3', trip_id: '1', notes: '', lat: null, lon: null, name: 'Osaka', start_date: '2026-10-09', end_date: '2026-10-10' },
      { id: '4', trip_id: '1', notes: '', lat: null, lon: null, name: 'Hiroshima', start_date: '2026-10-11', end_date: '2026-10-13' },
      { id: '5', trip_id: '1', notes: '', lat: null, lon: null, name: 'Fukuoka', start_date: '2026-10-14', end_date: '2026-10-15' },
      { id: '6', trip_id: '1', notes: '', lat: null, lon: null, name: 'Himeji', start_date: '2026-10-15', end_date: '2026-10-17' },
    ];

    const result = calculateTransport(places, 17);

    expect(result).not.toBeNull();
    // 5 trayectos: Tokio-Kioto, Kioto-Osaka, Osaka-Hiroshima, Hiroshima-Fukuoka, Fukuoka-Himeji
    expect(result!.routesFound).toHaveLength(5);
    expect(result!.routesFound[0].from).toBe('Tokio');
    expect(result!.routesFound[1].from).toBe('Kioto');
    expect(result!.routesFound[2].from).toBe('Osaka');
    expect(result!.routesFound[3].from).toBe('Hiroshima');
    expect(result!.routesFound[4].from).toBe('Fukuoka');
    
    // Transport span is from Oct 05 to Oct 15 = 11 days (which is <= 14).
    expect(result!.recommendedPass).toBe('14_day');
  });

  it('resuelve Fukuoka a Kyoto con normalización de fukuoka/hakata', () => {
    const places: Place[] = [
      { id: '1', trip_id: '1', notes: '', lat: null, lon: null, name: 'Fukuoka (Hakata)', start_date: '2026-10-10', end_date: '2026-10-11' },
      { id: '2', trip_id: '1', notes: '', lat: null, lon: null, name: 'Kyoto', start_date: '2026-10-11', end_date: '2026-10-12' }
    ];

    const result = calculateTransport(places, 2);

    expect(result).not.toBeNull();
    expect(result!.routesFound).toHaveLength(1);
    expect(result!.routesFound[0].from).toBe('Fukuoka (Hakata)');
    expect(result!.routesFound[0].to).toBe('Kyoto');
    expect(result!.routesFound[0].fare).toBe(16370);
  });
});
