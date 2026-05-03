import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { getCurrencySymbol, getCategoryEmoji, getCategoryName, getRelativeDateLabel } from './utils';

describe('utils.ts', () => {
  describe('getCurrencySymbol', () => {
    it('should return correct symbol for EUR', () => {
      expect(getCurrencySymbol('EUR')).toBe('€');
    });

    it('should return correct symbol for USD', () => {
      expect(getCurrencySymbol('USD')).toBe('$');
    });

    it('should return the code if no symbol is found', () => {
      expect(getCurrencySymbol('PHP')).toBe('PHP');
    });
  });

  describe('getCategoryEmoji', () => {
    it('should return food emoji for "food" slug', () => {
      expect(getCategoryEmoji('food')).toBe('🍔');
    });

    it('should return default emoji for unknown slug', () => {
      expect(getCategoryEmoji('unknown')).toBe('📦');
    });

    it('should return default emoji for undefined slug', () => {
      expect(getCategoryEmoji(undefined)).toBe('📦');
    });
  });

  describe('getCategoryName', () => {
    it('should capitalize the slug', () => {
      expect(getCategoryName('food')).toBe('Food');
    });

    it('should return "Other" for undefined slug', () => {
      expect(getCategoryName(undefined)).toBe('Other');
    });
  });

  describe('getRelativeDateLabel', () => {
    beforeEach(() => {
      // Mock "Today" to 2026-05-03
      vi.useFakeTimers();
      vi.setSystemTime(new Date('2026-05-03'));
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it('should return "Today" for current date', () => {
      expect(getRelativeDateLabel('2026-05-03')).toBe('Today');
    });

    it('should return "Tomorrow" for next day', () => {
      expect(getRelativeDateLabel('2026-05-04')).toBe('Tomorrow');
    });

    it('should return "Yesterday" for previous day', () => {
      expect(getRelativeDateLabel('2026-05-02')).toBe('Yesterday');
    });

    it('should return formatted date for other dates', () => {
      // Note: testing locale-dependent output can be tricky, but let's assume default locale
      const result = getRelativeDateLabel('2026-05-10', 'en-US');
      expect(result).toBe('May 10');
    });
  });
});
