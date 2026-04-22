export interface Expense {
    id: string;
    user_id?: string;
    session_id?: string;
    trip_id?: string;
    amount: number;
    currency: string;
    category_id?: string;
    created_at: string;
    updated_at?: string;
}

export interface CategorySummary {
    category_id: string;
    total: number;
}