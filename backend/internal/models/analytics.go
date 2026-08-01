package models

import "time"

// AnalyticsOverview is the response shape of GET /api/analytics/overview.
// It aggregates the high-level metrics every product question starts with.
type AnalyticsOverview struct {
	RealUsers     int       `json:"real_users"`
	RealTrips     int       `json:"real_trips"`
	DAU           int       `json:"dau"`
	WAU           int       `json:"wau"`
	MAU           int       `json:"mau"`
	TotalExpenses float64   `json:"total_expenses"`
	AsOf          time.Time `json:"as_of"`
}

// AnalyticsActivation measures the % of landing sessions that produced a
// trip within 7 days of their first landing_viewed event.
type AnalyticsActivation struct {
	Landed              int        `json:"landed"`
	Activated           int        `json:"activated"`
	ActivationRatePct   float64    `json:"activation_rate_pct"`
	AvgHoursToActivate  *float64   `json:"avg_hours_to_activate"`
	AsOf                time.Time  `json:"as_of"`
}

// AnalyticsEngagement surfaces depth-of-use: avg/median expenses per trip,
// power-user counts, etc.
type AnalyticsEngagement struct {
	AvgExpensesPerTrip    float64   `json:"avg_expenses_per_trip"`
	MedianExpensesPerTrip float64   `json:"median_expenses_per_trip"`
	MaxExpensesPerTrip    float64   `json:"max_expenses_per_trip"`
	AvgPlacesPerTrip      float64   `json:"avg_places_per_trip"`
	PowerUsers            int       `json:"power_users"`
	TotalUsers            int       `json:"total_users"`
	PowerUserRatePct      float64   `json:"power_user_rate_pct"`
	AsOf                  time.Time `json:"as_of"`
}

// AnalyticsRetention is the raw counts behind the retention percentage; the
// handler computes the percentage.
type AnalyticsRetention struct {
	WindowDays    int       `json:"window_days"`
	TotalUsers    int       `json:"total_users"`
	RetainedUsers int       `json:"retained_users"`
	RetentionPct  float64   `json:"retention_pct"`
	AsOf          time.Time `json:"as_of"`
}

// AnalyticsFunnel captures the three counters of the Ghost Mode funnel.
// Percentages are computed in the handler (safePct helper).
type AnalyticsFunnel struct {
	Viewed          int       `json:"viewed"`
	Forked          int       `json:"forked"`
	Reused          int       `json:"reused"`
	ViewToForkPct   float64   `json:"view_to_fork_pct"`
	ForkToReusePct  float64   `json:"fork_to_reuse_pct"`
	AsOf            time.Time `json:"as_of"`
}

// AnalyticsPowerUser is one row of the top-engaged-users list.
type AnalyticsPowerUser struct {
	SessionID    string `json:"session_id"`
	TripCount    int    `json:"trip_count"`
	ExpenseCount int    `json:"expense_count"`
	PlaceCount   int    `json:"place_count"`
	IsPowerUser  bool   `json:"is_power_user"`
}

// AnalyticsPowerUsers wraps the list with metadata so callers can tell
// whether the limit was hit.
type AnalyticsPowerUsers struct {
	Users []AnalyticsPowerUser `json:"users"`
	Count int                  `json:"count"`
	AsOf  time.Time            `json:"as_of"`
}
