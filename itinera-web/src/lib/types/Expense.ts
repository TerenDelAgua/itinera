export interface Expense {
    id: string;
    trip_id?: string;
    place_id?: string; //undefined = global expense

    // Financial
    amount: number;
    original_amount?: number;
    original_currency?: string;
    exchange_rate?: number;
    conversion_date?: string;

    currency: string;
    category_id?: string;
    notes: string;
    date: string;
    created_at: string;
    updated_at?: string;
}

export interface CategorySummary {
    category_id: string;
    total: number;
}