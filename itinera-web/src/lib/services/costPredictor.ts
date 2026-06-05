import { isJapanPlace } from '../utils/place';
import { calculateTransport } from './transport';
import { apiFetch } from '$lib/api';
import type { Trip } from '../types/Trip';
import type { Place } from '../types/Place';
import type { Activity } from '../types/Activity';
import type { Expense } from '../types/Expense';
import type { Expense_Category } from '../types/Category';
import costModelsRaw from '../data/cost_models.json';


export interface CostModel {
  accommodation: { budget: number; mid: number; luxury: number; unit: string };
  food: { budget: number; mid: number; luxury: number; unit: string };
  transport: { intercity_base: number; shinkansen_per_km: number; local_per_day: number; unit: string };
  activities: { temple: number; museum: number; onsen: number; theme_park: number; default: number; unit: string };
  misc: { daily_buffer: number; unit: string };
}

export interface CityMultiplier {
  accommodation: number;
  food: number;
  transport: number;
  activities: number;
}

export interface CostModels {
  version: string;
  last_updated: string;
  cost_models: {
    japan: CostModel;
  };
  city_multipliers: Record<string, CityMultiplier>;
  intercity_distances_km: Record<string, number>;
}

export const costModels = costModelsRaw as CostModels;

export interface CostEstimate {
  total: {
    estimated: number;      // En trip.currency, sin vuelo
    actual: number;         // Suma de expenses en trip.currency (convertidos)
    currency: string;       // trip.currency para display
  };
  byCategory: {
    accommodation: { estimated: number; actual: number };
    food: { estimated: number; actual: number };
    transport: { estimated: number; actual: number };
    activities: { estimated: number; actual: number };
    misc: { estimated: number; actual: number };
  };
  status: 'on_track' | 'halfway' | 'over_budget';
  excludesFlights: true;    // Siempre true en v1

  context: {
    phase: 'pre_trip' | 'active' | 'completed';
    daysUntilStart: number;   // días hasta inicio (0+ en pre_trip, 0 en active/completed)
    daysElapsedDisplay: number;      // días desde inicio (1+ en active, totalDays en completed)
    totalDays: number;
    expectedRatio: number;    // daysElapsed / totalDays (0 en pre_trip)
  };
}

export interface ExchangeService {
  getRate(from: string, to: string): Promise<number>;
}

export class BackendExchangeService implements ExchangeService {
  constructor(private tripId: string) {}

  async getRate(from: string, to: string): Promise<number> {
    if (from === to) return 1;
    try {
      const data = await apiFetch<{rate: number}>(`/trips/${this.tripId}/rates?from=${from}&to=${to}`);
      return data.rate || 1;
    } catch (e) {
      console.warn(`Failed to fetch rate from ${from} to ${to}:`, e);
      // Fallback rates if offline/failed
      const fallbacks: Record<string, number> = {
        'EUR->JPY': 170,
        'JPY->EUR': 1 / 170,
        'USD->JPY': 155,
        'JPY->USD': 1 / 155,
        'GBP->JPY': 200,
        'JPY->GBP': 1 / 200,
        // Cross rates
        'JPY->JPY': 1,
      };
      const key = `${from}->${to}`;
      if (fallbacks[key] !== undefined) return fallbacks[key];

      // If we don't have it directly, try converting through JPY as intermediary if possible, or return 1
      if (from !== 'JPY' && to !== 'JPY') {
        const toJpy = fallbacks[`${from}->JPY`] || 1;
        const fromJpy = fallbacks[`JPY->${to}`] || 1;
        return toJpy * fromJpy;
      }
      return 1;
    }
  }
}

export interface CostPredictor {
  estimate(
    trip: Trip,
    places: Place[],
    activities: Activity[],
    expenses: Expense[],
    categories: Expense_Category[]
  ): Promise<CostEstimate | null>;
}

export class HeuristicPredictor implements CostPredictor {
  private costModels: CostModels;
  private exchangeService: ExchangeService;
  private rateCache = new Map<string, { rate: number; timestamp: number }>();
  private readonly CACHE_TTL = 5 * 60 * 1000; // 5 minutos

  constructor(costModels: CostModels, exchangeService: ExchangeService) {
    this.costModels = costModels;
    this.exchangeService = exchangeService;
  }

  // Filtro geográfico (solo Japón)
  static shouldShow(places: Place[]): boolean {
    return places.some(p => isJapanPlace(p));
  }

  async estimate(
    trip: Trip,
    places: Place[],
    activities: Activity[],
    expenses: Expense[],
    categories: Expense_Category[]
  ): Promise<CostEstimate | null> {
    if (!HeuristicPredictor.shouldShow(places)) return null;

    const totalDays = this.calculateDurationDays(trip.start_date, trip.end_date);
    const now = Date.now();
    const start = new Date(trip.start_date).getTime();
    const end = new Date(trip.end_date).getTime();

    const daysUntilStart = Math.max(0, Math.ceil((start - now) / (1000 * 60 * 60 * 24)));
    
    // Gasto temporal ratio
    const daysElapsedRaw = Math.max(0, Math.ceil((now - start) / (1000 * 60 * 60 * 24)));
    const daysElapsedDisplay = Math.max(1, daysElapsedRaw);

    // Fase temporal
    let phase: 'pre_trip' | 'active' | 'completed';
    if (now < start) {
      phase = 'pre_trip';
    } else if (now > end) {
      phase = 'completed';
    } else {
      phase = 'active';
    }

    // Regla de archivo: Si el viaje acabó hace > 30 días, no mostrar badge
    if (phase === 'completed') {
      const daysSinceEnd = Math.ceil((now - end) / (1000 * 60 * 60 * 24));
      if (daysSinceEnd > 30) return null;
    }

    const expectedRatio = totalDays > 0 ? (phase === 'pre_trip' ? 0 : daysElapsedRaw / totalDays) : 0;

    const estimatedJPY = this.calculateEstimateJPY(trip, places, activities);
    
    const targetCurrency = trip.base_currency || 'EUR';
    const rate = await this.getRate('JPY', targetCurrency);
    const estimatedDisplay = Math.round(estimatedJPY.total * rate);

    const actualDisplayUnrounded = await this.sumExpensesInCurrency(expenses, targetCurrency);
    const actualDisplay = Math.round(actualDisplayUnrounded);

    // Status solo en fase active
    let status: CostEstimate['status'] = 'on_track';
    if (phase === 'active') {
      const actualRatio = actualDisplay / (estimatedDisplay || 1);
      if (actualRatio <= expectedRatio * 1.2) {
        status = 'on_track';
      } else if (actualRatio <= expectedRatio * 1.5) {
        status = 'halfway';
      } else {
        status = 'over_budget';
      }
    }

    const byCategory = await this.convertCategoriesToDisplay(estimatedJPY.byCategory, expenses, categories, rate, targetCurrency);

    return {
      total: { estimated: estimatedDisplay, actual: actualDisplay, currency: trip.base_currency || 'EUR' },
      byCategory,
      status,
      excludesFlights: true,
      context: { phase, daysUntilStart, daysElapsedDisplay, totalDays, expectedRatio }
    };
  }

  private calculateEstimateJPY(
    trip: Trip,
    places: Place[],
    activities: Activity[]
  ): { total: number; byCategory: Record<string, number> } {
    const days = this.calculateDurationDays(trip.start_date, trip.end_date);
    const cities = places.map(p => p.city?.toLowerCase() || '');
    const model = this.costModels.cost_models.japan;
    const multiplier = this.getAverageMultiplier(cities);

    const accommodation = days * model.accommodation.mid * multiplier.accommodation;
    const food = days * model.food.mid * multiplier.food;
    const transport = this.calculateTransportCost(cities, places, model, multiplier.transport);
    const activitiesCost = this.estimateActivities(activities, days, model, multiplier.activities);
    const misc = days * model.misc.daily_buffer;

    return {
      total: accommodation + food + transport + activitiesCost + misc,
      byCategory: {
        accommodation: Math.round(accommodation),
        food: Math.round(food),
        transport: Math.round(transport),
        activities: Math.round(activitiesCost),
        misc: Math.round(misc)
      }
    };
  }

  // TRANSPORTE HÍBRIDO: datos reales → heurística fallback
  private calculateTransportCost(
    cities: string[], 
    places: Place[], 
    model: CostModel, 
    multiplier: number
  ): number {
    const validPlaces = places.filter(p => p.city);

    // Si hay ≥2 lugares con city, intentar cálculo real
    if (validPlaces.length >= 2) {
      try {
        const realCalc = calculateTransport(validPlaces, cities.length);
        if (realCalc && typeof realCalc.totalFare === 'number') {
          return realCalc.totalFare;
        }
      } catch (e) {
        console.warn('Transport calculation failed, using heuristic fallback', e);
      }
    }

    // FALLBACK HEURÍSTICO
    if (cities.length <= 1) {
      return cities.length * 7 * model.transport.local_per_day * multiplier;
    }

    let intercity = 0;
    for (let i = 0; i < cities.length - 1; i++) {
      const pair = [cities[i], cities[i + 1]].sort().join('_');
      const distance = this.costModels.intercity_distances_km[pair] || 200;
      intercity += model.transport.intercity_base + (distance * model.transport.shinkansen_per_km);
    }
    const local = cities.length * 3 * model.transport.local_per_day * multiplier;
    return intercity + local;
  }

  // ACTIVIDADES: reales + buffer
  private estimateActivities(
    activities: Activity[], 
    tripDays: number, 
    model: CostModel, 
    multiplier: number
  ): number {
    const paidActivities = activities.filter(a => {
      const titleLower = a.title?.toLowerCase() || '';
      return !['walk', 'free', 'sightseeing', 'photo'].some(k => titleLower.includes(k));
    });

    if (paidActivities.length > 0) {
      const categorizedCost = paidActivities.reduce(
        (sum, a) => sum + this.categorizeActivity(a.title), 
        0
      );
      const uncategorizedBuffer = Math.max(0, tripDays * 2 - paidActivities.length) * 2000;
      return (categorizedCost + uncategorizedBuffer) * multiplier;
    }

    return tripDays * 2 * 2000 * multiplier;
  }

  private categorizeActivity(name: string): number {
    const n = name.toLowerCase();
    const model = this.costModels.cost_models.japan.activities;
    if (n.includes('temple') || n.includes('shrine')) return model.temple;
    if (n.includes('museum') || n.includes('gallery')) return model.museum;
    if (n.includes('onsen') || n.includes('bath')) return model.onsen;
    if (n.includes('disney') || n.includes('universal') || n.includes('park')) return model.theme_park;
    return model.default;
  }

  // CONVERSIONES CON CACHE PERSISTENTE
  private async getRate(from: string, to: string): Promise<number> {
    if (from === to) return 1;

    const key = `${from}->${to}`;
    const cached = this.rateCache.get(key);

    if (cached && Date.now() - cached.timestamp < this.CACHE_TTL) {
      return cached.rate;
    }

    const rate = await this.exchangeService.getRate(from, to);
    this.rateCache.set(key, { rate, timestamp: Date.now() });
    return rate;
  }

  private async sumExpensesInCurrency(expenses: Expense[], targetCurrency: string): Promise<number> {
    let total = 0;
    for (const expense of expenses) {
      if (expense.currency === targetCurrency) {
        total += expense.amount;
      } else {
        const rate = await this.getRate(expense.currency, targetCurrency);
        total += expense.amount * rate;
      }
    }
    return total;
  }

  private calculateDurationDays(start: string, end: string): number {
    return Math.max(1, Math.ceil(
      (new Date(end).getTime() - new Date(start).getTime()) / (1000 * 60 * 60 * 24)
    ));
  }

  private getAverageMultiplier(cities: string[]): CityMultiplier {
    if (cities.length === 0) return { accommodation: 1, food: 1, transport: 1, activities: 1 };

    const sum = cities.reduce((acc, city) => {
      const m = this.costModels.city_multipliers[city] || { accommodation: 1, food: 1, transport: 1, activities: 1 };
      return {
        accommodation: acc.accommodation + m.accommodation,
        food: acc.food + m.food,
        transport: acc.transport + m.transport,
        activities: acc.activities + m.activities
      };
    }, { accommodation: 0, food: 0, transport: 0, activities: 0 });

    return {
      accommodation: sum.accommodation / cities.length,
      food: sum.food / cities.length,
      transport: sum.transport / cities.length,
      activities: sum.activities / cities.length
    };
  }

  private async convertCategoriesToDisplay(
    estimatedJPY: Record<string, number>,
    expenses: Expense[],
    categories: Expense_Category[],
    rate: number,
    targetCurrency: string
  ): Promise<CostEstimate['byCategory']> {
    const actualTarget: Record<string, number> = {
      accommodation: 0,
      food: 0,
      transport: 0,
      activities: 0,
      misc: 0
    };

    for (const expense of expenses) {
      const cat = categories.find(c => c.id === expense.category_id);
      const slug = cat ? cat.slug.toLowerCase() : 'misc';

      const expenseTarget = expense.currency === targetCurrency
        ? expense.amount
        : expense.amount * (await this.getRate(expense.currency, targetCurrency));

      if (slug === 'accommodation' || slug === 'lodging') {
        actualTarget.accommodation += expenseTarget;
      } else if (slug === 'food' || slug === 'dining') {
        actualTarget.food += expenseTarget;
      } else if (slug === 'transport' || slug === 'transit') {
        actualTarget.transport += expenseTarget;
      } else if (slug === 'activities' || slug === 'leisure' || slug === 'sightseeing') {
        actualTarget.activities += expenseTarget;
      } else {
        actualTarget.misc += expenseTarget;
      }
    }

    return {
      accommodation: {
        estimated: Math.round(estimatedJPY.accommodation * rate),
        actual: Math.round(actualTarget.accommodation)
      },
      food: {
        estimated: Math.round(estimatedJPY.food * rate),
        actual: Math.round(actualTarget.food)
      },
      transport: {
        estimated: Math.round(estimatedJPY.transport * rate),
        actual: Math.round(actualTarget.transport)
      },
      activities: {
        estimated: Math.round(estimatedJPY.activities * rate),
        actual: Math.round(actualTarget.activities)
      },
      misc: {
        estimated: Math.round(estimatedJPY.misc * rate),
        actual: Math.round(actualTarget.misc)
      }
    };
  }
}
