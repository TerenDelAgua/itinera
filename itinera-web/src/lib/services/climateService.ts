import dataset from '$lib/data/japan_context.json';

export interface ClimateDisplay {
    temp_min: number;
    temp_max: number;
    temp_current?: number;
    rain_mm?: number;
    icon: string;
    notes: string;
    is_historic: boolean;
}

interface OpenMeteoResponse {
    current: {
        temperature_2m: number;
        weather_code: number;
    };
    daily: {
        temperature_2m_max: number[];
        temperature_2m_min: number[];
        weather_code: number[];
    };
}

const climateCache = new Map<string, { data: ClimateDisplay, timestamp: number }>();
const CACHE_TTL = 4 * 60 * 60 * 1000; // 4 hours

// Convert WMO weather codes to our icons
function weatherCodeToIcon(code: number): string {
    if (code === 0 || code === 1) return 'clear';
    if (code === 2 || code === 3) return 'partly-cloudy';
    if (code >= 45 && code <= 48) return 'fog';
    if ((code >= 51 && code <= 67) || (code >= 80 && code <= 82)) return 'rainy';
    if ((code >= 71 && code <= 77) || (code >= 85 && code <= 86)) return 'snow';
    if (code >= 95) return 'storm';
    return 'partly-cloudy';
}

export function shouldShowClimate(activityDate: string, tripStart?: string, tripEnd?: string): boolean {
    if (!tripStart || !tripEnd) return false;
    
    const today = new Date().toISOString().split('T')[0];
    const isDuringTrip = today >= tripStart && today <= tripEnd;
    
    if (!isDuringTrip) return false; // Traveler is not currently on the trip

    const actDateObj = new Date(activityDate);
    const todayObj = new Date(today);
    const diffTime = actDateObj.getTime() - todayObj.getTime();
    const diffDays = diffTime / (1000 * 3600 * 24);

    return diffDays >= 0 && diffDays <= 2;
}

export async function getClimate(city: string, date: string, placeLat?: number, placeLon?: number): Promise<ClimateDisplay | null> {
    const cityData = (dataset.city_metadata as any)[city.toLowerCase()];
    if (!cityData && !placeLat) return null;

    const lat = placeLat || cityData?.lat;
    const lon = placeLon || cityData?.lon;
    const region = cityData?.region;
    
    const cacheKey = `${lat},${lon},${date}`;
    const cached = climateCache.get(cacheKey);
    if (cached && Date.now() - cached.timestamp < CACHE_TTL) {
        return cached.data;
    }

    const month = new Date(date).getMonth() + 1;
    let historicData = null;
    if (region && (dataset.climate as any)[region]) {
        historicData = (dataset.climate as any)[region].months[month.toString()];
    }

    if (lat && lon) {
        try {
            const controller = new AbortController();
            const timeoutId = setTimeout(() => controller.abort(), 3000);
            
            const url = `https://api.open-meteo.com/v1/forecast?latitude=${lat}&longitude=${lon}&current=temperature_2m,weather_code&daily=weather_code,temperature_2m_max,temperature_2m_min&timezone=auto&forecast_days=4`;
            const response = await fetch(url, { signal: controller.signal });
            clearTimeout(timeoutId);
            
            if (response.ok) {
                const data: OpenMeteoResponse = await response.json();
                
                // Find index for the requested date
                let dayIndex = 0;
                // For simplicity, we just use the first day (current forecast) or if it's up to 3 days.
                // A better approach would parse the dates array in daily.time
                
                const display: ClimateDisplay = {
                    temp_current: Math.round(data.current.temperature_2m),
                    temp_min: Math.round(data.daily.temperature_2m_min[dayIndex]),
                    temp_max: Math.round(data.daily.temperature_2m_max[dayIndex]),
                    icon: weatherCodeToIcon(data.daily.weather_code[dayIndex]),
                    notes: historicData ? historicData.notes : '',
                    is_historic: false
                };
                
                climateCache.set(cacheKey, { data: display, timestamp: Date.now() });
                return display;
            }
        } catch (e) {
            console.warn('OpenMeteo API failed, falling back to historic data', e);
        }
    }

    if (historicData) {
        const display: ClimateDisplay = {
            temp_min: historicData.temp_c_min,
            temp_max: historicData.temp_c_max,
            rain_mm: historicData.rain_mm,
            icon: historicData.icon,
            notes: historicData.notes,
            is_historic: true
        };
        climateCache.set(cacheKey, { data: display, timestamp: Date.now() });
        return display;
    }

    return null;
}
