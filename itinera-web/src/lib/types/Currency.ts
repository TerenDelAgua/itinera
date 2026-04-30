export interface Currency {
    code: string;
    symbol: string;
    name: string;
}

export const COMMON_CURRENCIES: Currency[] = [
    { code: 'EUR', symbol: '\u20AC', name: 'Euro' },
    { code: 'USD', symbol: '$', name: 'US Dollar' },
    { code: 'GBP', symbol: '\u00A3', name: 'British Pound' },
    { code: 'JPY', symbol: '\u00A5', name: 'Japanese Yen' },
    { code: 'CHF', symbol: 'Fr', name: 'Swiss Franc' },
    { code: 'CAD', symbol: 'C$', name: 'Canadian Dollar' },
    { code: 'AUD', symbol: 'A$', name: 'Australian Dollar' },
];
