import { describe, it, expect } from 'vitest';
import { HeuristicPredictor, costModels } from './costPredictor';
import type { Trip } from '../types/Trip';
import type { Place } from '../types/Place';
import type { Activity } from '../types/Activity';
import type { Expense } from '../types/Expense';
import type { Expense_Category } from '../types/Category';

// Mock Exchange Service
class MockExchangeService {
  async getRate(from: string, to: string): Promise<number> {
    if (from === 'JPY' && to === 'EUR') return 1 / 170;
    if (from === 'EUR' && to === 'JPY') return 170;
    if (from === 'USD' && to === 'JPY') return 155;
    return 1;
  }
}

describe('HeuristicPredictor shouldShow', () => {
  it('returns true if any place is in Japan', () => {
    const places: Place[] = [
      { id: '1', trip_id: 'trip1', name: 'Paris', city: 'Paris', country_code: 'FR', notes: '', lat: null, lon: null },
      { id: '2', trip_id: 'trip1', name: 'Tokyo', city: 'Tokyo', country_code: 'JP', notes: '', lat: null, lon: null }
    ];
    expect(HeuristicPredictor.shouldShow(places)).toBe(true);
  });

  it('returns false if no places are in Japan', () => {
    const places: Place[] = [
      { id: '1', trip_id: 'trip1', name: 'Paris', city: 'Paris', country_code: 'FR', notes: '', lat: null, lon: null },
      { id: '2', trip_id: 'trip1', name: 'Rome', city: 'Rome', country_code: 'IT', notes: '', lat: null, lon: null }
    ];
    expect(HeuristicPredictor.shouldShow(places)).toBe(false);
  });
});

describe('HeuristicPredictor estimate', () => {
  const exchangeService = new MockExchangeService();
  const predictor = new HeuristicPredictor(costModels, exchangeService);

  const categories: Expense_Category[] = [
    { id: 'cat_accomm', name: 'Accommodation', icon: '🏨', slug: 'accommodation' },
    { id: 'cat_food', name: 'Food', icon: '🍔', slug: 'food' },
    { id: 'cat_transport', name: 'Transport', icon: '🚆', slug: 'transport' },
    { id: 'cat_activities', name: 'Activities', icon: '🎟️', slug: 'activities' },
    { id: 'cat_misc', name: 'Misc', icon: '📦', slug: 'misc' }
  ];

  it('returns null if places are not in Japan', async () => {
    const trip: Trip = {
      id: 'trip1',
      name: 'Europe Trip',
      start_date: '2026-06-01',
      end_date: '2026-06-10',
      base_currency: 'EUR',
      default_expense_currency: 'EUR',
      is_public_demo: false,
      created_at: new Date().toISOString(),
      place_count: 1,
      total_spent: 0
    };
    const places: Place[] = [{ id: 'p1', trip_id: 'trip1', name: 'Paris', city: 'Paris', country_code: 'FR', notes: '', lat: null, lon: null }];

    const result = await predictor.estimate(trip, places, [], [], categories);
    expect(result).toBeNull();
  });

  it('returns cost estimate for Japanese trip - Pre-trip phase', async () => {
    const today = new Date();
    const start_date = new Date(today.getTime() + 10 * 24 * 60 * 60 * 1000).toISOString().split('T')[0]; // 10 days in future
    const end_date = new Date(today.getTime() + 15 * 24 * 60 * 60 * 1000).toISOString().split('T')[0]; // 15 days in future

    const trip: Trip = {
      id: 'trip_jp',
      name: 'Japan Trip',
      start_date,
      end_date,
      base_currency: 'EUR',
      default_expense_currency: 'EUR',
      is_public_demo: false,
      created_at: today.toISOString(),
      place_count: 1,
      total_spent: 0
    };

    const places: Place[] = [{ id: 'p1', trip_id: 'trip_jp', name: 'Tokyo Hotel', city: 'Tokyo', country_code: 'JP', notes: '', lat: null, lon: null }];
    const activities: Activity[] = [];
    const expenses: Expense[] = [
      {
        id: 'exp1',
        amount: 85,
        currency: 'EUR',
        category_id: 'cat_accomm',
        notes: 'Deposit',
        date: today.toISOString(),
        created_at: today.toISOString()
      }
    ];

    const result = await predictor.estimate(trip, places, activities, expenses, categories);
    expect(result).not.toBeNull();
    expect(result!.context.phase).toBe('pre_trip');
    expect(result!.context.daysUntilStart).toBeGreaterThan(0);
    expect(result!.total.estimated).toBeGreaterThan(0);
    // 85 EUR converted to JPY: 85 * 170 = 14450 JPY. Then back to EUR: 85 EUR.
    expect(result!.total.actual).toBe(85);
  });
});
