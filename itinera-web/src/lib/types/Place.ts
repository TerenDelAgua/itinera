export interface Place {
    id: string;
    trip_id: string;
    name: string;
    notes: string;
    start_date?: string;
    end_date?: string;
    lat: number;
    lon: number;
    default_expense_currency?: string;
    total_expenses?: number;
}
