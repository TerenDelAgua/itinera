export const currencySymbols: Record<string, string> = {
  EUR: '€',
  USD: '$',
  GBP: '£',
  JPY: '¥'
};

export function getCurrencySymbol(code: string): string {
  return currencySymbols[code] || code;
}
