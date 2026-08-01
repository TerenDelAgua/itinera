export interface Trip {
    id: string;
    user_id?: string;
    session_id?: string;
    name: string;
    description?: string;
    start_date: string;
    end_date: string;
    base_currency: string;
    default_expense_currency: string;
    is_public_demo: boolean;
    forked_from?: string;
    created_at: string;
    updated_at?: string;
    place_count: number;
    total_spent: number;

    share_token?: string;
    share_enabled: boolean;
    share_expires_at?: string;
    share_created_at?: string;
}
