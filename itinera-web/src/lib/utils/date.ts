export function formatDate(dateStr: string | undefined | null, locale: string = 'en-US'): string {
    if (!dateStr) return '';

    // Check if it's today or tomorrow
    const today = new Date();
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);

    const targetDate = new Date(dateStr);

    // Compare YYYY-MM-DD to avoid timezone issues with exact equality
    const formatYMD = (d: Date) => d.toISOString().split('T')[0];
    const targetYMD = targetDate.toISOString().split('T')[0];

    if (targetYMD === formatYMD(today)) return 'Today';
    if (targetYMD === formatYMD(tomorrow)) return 'Tomorrow';

    return new Intl.DateTimeFormat(locale, {
        day: "2-digit",
        month: "short",
        year: "numeric",
    }).format(targetDate);
}

export function formatDisplayDate(dateStr: string | undefined | null, todayLabel: string, tomorrowLabel: string, locale: string = 'en-US'): string {
    if (!dateStr) return '';

    const today = new Date();
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);

    const formatYMD = (d: Date) => {
        // Need local date string, not ISO to avoid timezone shifts
        const offset = d.getTimezoneOffset()
        d = new Date(d.getTime() - (offset * 60 * 1000))
        return d.toISOString().split('T')[0]
    };

    // The dateStr from DB is usually YYYY-MM-DD
    const targetYMD = dateStr.includes('T') ? dateStr.split('T')[0] : dateStr;

    if (targetYMD === formatYMD(today)) return todayLabel.toUpperCase();
    if (targetYMD === formatYMD(tomorrow)) return tomorrowLabel.toUpperCase();

    // Note: ensure we treat the date string as local date to avoid TZ shifts
    const [year, month, day] = targetYMD.split("-").map(Number);
    const date = new Date(year, month - 1, day);

    return new Intl.DateTimeFormat(locale, {
        day: "2-digit",
        month: "short",
        year: "numeric",
    }).format(date);
}
