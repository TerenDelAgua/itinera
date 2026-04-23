export interface CategorySummary {
    category_id: string;
    total: number;
}

export interface PlaceSummary {
    place_id: string;
    place_name: string;
    total: number;
}

export interface TripExpenseSummary {
    global_total: number;
    places_total: number;
    grand_total: number;
    by_category: CategorySummary[];
    by_place: PlaceSummary[];
}