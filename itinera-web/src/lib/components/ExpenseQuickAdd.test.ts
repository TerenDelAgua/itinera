import { render, screen, fireEvent, cleanup } from '@testing-library/svelte';
import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import ExpenseQuickAdd from './ExpenseQuickAdd.svelte';

// Mock apiFetch
vi.mock('$lib/api', () => ({
  apiFetch: vi.fn()
}));

// Mock transitions to avoid issues during tests
vi.mock('svelte/transition', () => ({
  fly: vi.fn(() => ({ duration: 0 })),
  slide: vi.fn(() => ({ duration: 0 }))
}));

// Mock i18n store
vi.mock('$lib/i18n/store', () => {
  const { writable, derived } = require('svelte/store');
  const locale = writable('en');
  const t = derived(locale, () => (key: string) => key);
  return { t, locale };
});

describe('ExpenseQuickAdd.svelte', () => {
  const mockCategories = [
    { id: '1', slug: 'food', name: 'Food' },
    { id: '2', slug: 'transport', name: 'Transport' }
  ];

  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  afterEach(() => {
    cleanup();
  });

  it('should use default baseCurrency if nothing in localStorage', () => {
    render(ExpenseQuickAdd, {
      tripId: 'trip-1',
      categories: mockCategories,
      onSuccess: vi.fn(),
      baseCurrency: 'EUR'
    });

    // EUR symbol is €
    const currencySelect = screen.getByDisplayValue('€');
    expect(currencySelect).toBeInTheDocument();
  });

  it('should retrieve currency from localStorage if available', () => {
    localStorage.setItem('last_expense_currency', 'USD');

    render(ExpenseQuickAdd, {
      tripId: 'trip-1',
      categories: mockCategories,
      onSuccess: vi.fn(),
      baseCurrency: 'EUR'
    });

    // USD symbol is $
    const currencySelect = screen.getByDisplayValue('$');
    expect(currencySelect).toBeInTheDocument();
  });

  it('should save selected currency to localStorage', async () => {
    render(ExpenseQuickAdd, {
      tripId: 'trip-1',
      categories: mockCategories,
      onSuccess: vi.fn(),
      baseCurrency: 'EUR'
    });

    // Find the currency select. It has the symbol as text.
    const currencySelect = screen.getByDisplayValue('€');
    
    // Change currency to USD
    await fireEvent.change(currencySelect, { target: { value: 'USD' } });

    expect(localStorage.getItem('last_expense_currency')).toBe('USD');
  });

  it('should sort categories placing "food" first', () => {
    const cats = [
      { id: '2', slug: 'transport', name: 'Transport' },
      { id: '1', slug: 'food', name: 'Food' },
      { id: '3', slug: 'accommodation', name: 'Accommodation' }
    ];

    render(ExpenseQuickAdd, {
      tripId: 'trip-1',
      categories: cats,
      onSuccess: vi.fn()
    });

    const options = screen.getAllByRole('option');
    // First option should be the emoji for food (🍔)
    expect(options[0].textContent).toBe('🍔');
  });
});
