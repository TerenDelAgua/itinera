export const currencySymbols: Record<string, string> = {
  EUR: '\u20AC',
  USD: '$',
  GBP: '\u00A3',
  JPY: '\u00A5'
};

export function getCurrencySymbol(code: string): string {
  return currencySymbols[code] || code;
}

export const categoryEmojiMap: Record<string, string> = {
  accommodation: '\u{1F3E8}',
  transport: '\u{1F686}',
  food: '\u{1F354}',
  leisure: '\u{1F39F}\uFE0F',
  shopping: '\u{1F6CD}\uFE0F',
  others: '\u{1F4E6}'
};

export function getCategoryEmoji(slug: string | undefined): string {
  if (!slug) return '\u{1F4E6}';
  return categoryEmojiMap[slug] || '\u{1F4E6}';
}

export function getCategoryName(slug: string | undefined): string {
  if (!slug) return 'Other';
  return slug.charAt(0).toUpperCase() + slug.slice(1);
}
