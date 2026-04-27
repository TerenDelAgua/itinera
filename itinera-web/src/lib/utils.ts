export const currencySymbols: Record<string, string> = {
  EUR: '€',
  USD: '$',
  GBP: '£',
  JPY: '¥'
};

export function getCurrencySymbol(code: string): string {
  return currencySymbols[code] || code;
}
export const categoryEmojiMap: Record<string, string> = {
  accommodation: '🏨',
  transport: '🚆',
  food: '🍔',
  leisure: '🎟️',
  shopping: '🛍️',
  others: '📦'
};

export function getCategoryEmoji(slug: string | undefined): string {
  if (!slug) return '📦';
  return categoryEmojiMap[slug] || '📦';
}

export function getCategoryName(slug: string | undefined): string {
  if (!slug) return 'Other';
  return slug.charAt(0).toUpperCase() + slug.slice(1);
}
