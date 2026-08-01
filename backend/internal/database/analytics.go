package database

import (
	"context"
	"time"

	"backend/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AnalyticsRepository wraps the analytics queries. Each method is a
// one-to-one mapping with an endpoint in the analytics module.
type AnalyticsRepository struct {
	Pool *pgxpool.Pool
}

func NewAnalyticsRepository(pool *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{Pool: pool}
}

// GetAnalyticsOverview aggregates the high-level metrics: real users,
// real trips, DAU/WAU/MAU, and total expenses across all real trips.
func (r *AnalyticsRepository) GetAnalyticsOverview(ctx context.Context) (models.AnalyticsOverview, error) {
	var out models.AnalyticsOverview
	err := r.Pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM v_real_sessions)        AS real_users,
			(SELECT COUNT(*) FROM v_real_trips)           AS real_trips,
			(SELECT COUNT(DISTINCT session_id)
			   FROM events
			  WHERE created_at > NOW() - INTERVAL '1 day'
			    AND session_id IN (SELECT session_id FROM v_real_sessions)) AS dau,
			(SELECT COUNT(DISTINCT session_id)
			   FROM events
			  WHERE created_at > NOW() - INTERVAL '7 days'
			    AND session_id IN (SELECT session_id FROM v_real_sessions)) AS wau,
			(SELECT COUNT(DISTINCT session_id)
			   FROM events
			  WHERE created_at > NOW() - INTERVAL '30 days'
			    AND session_id IN (SELECT session_id FROM v_real_sessions)) AS mau,
			(SELECT COALESCE(SUM(amount), 0) FROM expenses
			  WHERE trip_id IN (SELECT id FROM v_real_trips)) AS total_expenses
	`).Scan(
		&out.RealUsers,
		&out.RealTrips,
		&out.DAU,
		&out.WAU,
		&out.MAU,
		&out.TotalExpenses,
	)
	if err != nil {
		return models.AnalyticsOverview{}, err
	}
	out.AsOf = time.Now().UTC()
	return out, nil
}

// GetAnalyticsActivation computes the % of landing sessions that
// produced a trip. AvgHoursToActivate is NULL when there are no
// activated sessions (no division-by-zero issue).
func (r *AnalyticsRepository) GetAnalyticsActivation(ctx context.Context) (models.AnalyticsActivation, error) {
	var out models.AnalyticsActivation
	var landed, activated int
	var avgHours *float64
	err := r.Pool.QueryRow(ctx, `
		WITH real_landings AS (
			SELECT session_id, MIN(created_at) AS landed_at
			FROM events
			WHERE type = 'landing_viewed'
			  AND session_id IN (SELECT session_id FROM v_real_sessions)
			GROUP BY session_id
		),
		first_trips AS (
			SELECT session_id, MIN(created_at) AS trip_at
			FROM v_real_trips
			GROUP BY session_id
		)
		SELECT
			COUNT(*) AS landed,
			COUNT(ft.trip_at) AS activated,
			AVG(EXTRACT(EPOCH FROM (ft.trip_at - rl.landed_at)) / 3600.0) AS avg_hours
		FROM real_landings rl
		LEFT JOIN first_trips ft ON ft.session_id = rl.session_id
	`).Scan(&landed, &activated, &avgHours)
	if err != nil {
		return models.AnalyticsActivation{}, err
	}

	out.Landed = landed
	out.Activated = activated
	out.AvgHoursToActivate = avgHours
	if landed > 0 {
		out.ActivationRatePct = 100.0 * float64(activated) / float64(landed)
	}
	out.AsOf = time.Now().UTC()
	return out, nil
}

// GetAnalyticsEngagement reports depth of use: avg/median expenses per
// trip, avg places per trip, and power-user counts.
func (r *AnalyticsRepository) GetAnalyticsEngagement(ctx context.Context) (models.AnalyticsEngagement, error) {
	var out models.AnalyticsEngagement
	var avgExpenses, avgPlaces, medianExpenses, maxExpenses float64
	var powerUsers, totalUsers int
	err := r.Pool.QueryRow(ctx, `
		SELECT
			AVG(expense_count)::float                                            AS avg_expenses,
			AVG(place_count)::float                                              AS avg_places,
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY expense_count)::float    AS median_expenses,
			MAX(expense_count)::float                                            AS max_expenses,
			(SELECT COUNT(*) FROM v_user_engagement WHERE is_power_user)         AS power_users,
			(SELECT COUNT(*) FROM v_user_engagement)                            AS total_users
		FROM v_real_trips
	`).Scan(&avgExpenses, &avgPlaces, &medianExpenses, &maxExpenses, &powerUsers, &totalUsers)
	if err != nil {
		return models.AnalyticsEngagement{}, err
	}
	out.AvgExpensesPerTrip = avgExpenses
	out.AvgPlacesPerTrip = avgPlaces
	out.MedianExpensesPerTrip = medianExpenses
	out.MaxExpensesPerTrip = maxExpenses
	out.PowerUsers = powerUsers
	out.TotalUsers = totalUsers
	if totalUsers > 0 {
		out.PowerUserRatePct = 100.0 * float64(powerUsers) / float64(totalUsers)
	}
	out.AsOf = time.Now().UTC()
	return out, nil
}

// GetAnalyticsRetention counts how many users returned N days after their
// first trip. The caller clamps `days` to a sane range (1..365).
func (r *AnalyticsRepository) GetAnalyticsRetention(ctx context.Context, days int) (models.AnalyticsRetention, error) {
	var total, retained int
	err := r.Pool.QueryRow(ctx, `
		WITH first_trips AS (
			SELECT session_id, MIN(created_at) AS first_at
			FROM v_real_trips
			GROUP BY session_id
		)
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE EXISTS (
				SELECT 1 FROM events e
				WHERE e.session_id = ft.session_id
				  AND e.created_at > ft.first_at + ($1 || ' days')::interval
			)) AS retained
		FROM first_trips ft
	`, days).Scan(&total, &retained)
	if err != nil {
		return models.AnalyticsRetention{}, err
	}

	out := models.AnalyticsRetention{
		WindowDays:    days,
		TotalUsers:    total,
		RetainedUsers: retained,
		AsOf:          time.Now().UTC(),
	}
	if total > 0 {
		out.RetentionPct = 100.0 * float64(retained) / float64(total)
	}
	return out, nil
}

// GetAnalyticsFunnel reports the three counters of the Ghost Mode funnel
// over the last 30 days, scoped to real sessions.
func (r *AnalyticsRepository) GetAnalyticsFunnel(ctx context.Context) (models.AnalyticsFunnel, error) {
	var out models.AnalyticsFunnel
	var viewed, forked, reused int
	err := r.Pool.QueryRow(ctx, `
		SELECT
			COUNT(DISTINCT session_id) FILTER (WHERE type = 'demo_viewed')     AS viewed,
			COUNT(DISTINCT session_id) FILTER (WHERE type = 'demo_deep_forked') AS forked,
			COUNT(DISTINCT session_id) FILTER (WHERE type = 'demo_fork_reused') AS reused
		FROM events
		WHERE created_at > NOW() - INTERVAL '30 days'
		  AND session_id IN (SELECT session_id FROM v_real_sessions)
	`).Scan(&viewed, &forked, &reused)
	if err != nil {
		return models.AnalyticsFunnel{}, err
	}

	out.Viewed = viewed
	out.Forked = forked
	out.Reused = reused
	if viewed > 0 {
		out.ViewToForkPct = 100.0 * float64(forked) / float64(viewed)
	}
	if forked > 0 {
		out.ForkToReusePct = 100.0 * float64(reused) / float64(forked)
	}
	out.AsOf = time.Now().UTC()
	return out, nil
}

// GetAnalyticsPowerUsers returns the top `limit` sessions by trip_count,
// then total_expenses. The caller must clamp `limit` to <= 100.
func (r *AnalyticsRepository) GetAnalyticsPowerUsers(ctx context.Context, limit int) (models.AnalyticsPowerUsers, error) {
	out := models.AnalyticsPowerUsers{
		Users: []models.AnalyticsPowerUser{},
		AsOf:  time.Now().UTC(),
	}
	rows, err := r.Pool.Query(ctx, `
		SELECT session_id, trip_count, total_expenses, total_places, is_power_user
		FROM v_user_engagement
		ORDER BY trip_count DESC, total_expenses DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return out, err
	}
	defer rows.Close()

	for rows.Next() {
		var u models.AnalyticsPowerUser
		if err := rows.Scan(&u.SessionID, &u.TripCount, &u.ExpenseCount, &u.PlaceCount, &u.IsPowerUser); err != nil {
			return out, err
		}
		out.Users = append(out.Users, u)
	}
	out.Count = len(out.Users)
	return out, nil
}

// GetAnalyticsSessions returns the raw internal-vs-real counts.
// The handler turns this into JSON. Kept on the same repo so all
// analytics reads share one home.
func (r *AnalyticsRepository) GetAnalyticsSessions(ctx context.Context) (int, int, int, int, error) {
	var internalTrips, realTrips, internalSessions, realSessions int
	err := r.Pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE is_internal)             AS internal_trips,
			COUNT(*) FILTER (WHERE NOT is_internal)         AS real_trips,
			COUNT(DISTINCT session_id) FILTER (WHERE is_internal)             AS internal_sessions,
			COUNT(DISTINCT session_id) FILTER (WHERE NOT is_internal)         AS real_sessions
		FROM trips
	`).Scan(&internalTrips, &realTrips, &internalSessions, &realSessions)
	return internalTrips, realTrips, internalSessions, realSessions, err
}
