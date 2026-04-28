export interface Activity {
    id: string;
    trip_id: string;
    place_id?: string; //if not exists is global activity
    title: string;
    date: string; //YYY-MM-DD
    time?: string; //HH:MM
    notes?: string;
    created_at?: string
}