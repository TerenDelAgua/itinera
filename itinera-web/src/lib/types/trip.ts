export interface Trip {
    id: string;
    user_id?: string;
    session_id?: string;
    name: string;
    note?: string;
    start_date: string;
    end_date: string;
    base_currency: string;
    created_at: string;
    updated_at?: string;
}