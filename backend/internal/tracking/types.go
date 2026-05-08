package tracking

type EventType string

const (
	// Navigation
	LandingViewed   EventType = "landing_viewed"
	TripsListViewed EventType = "trips_list_viewed"

	// Trip Core
	TripCreated EventType = "trip_created"
	TripViewed  EventType = "trip_viewed"
	TripUpdated EventType = "trip_updated"
	TripDeleted EventType = "trip_deleted"

	// Demo fork system
	DemoViewed     EventType = "demo_viewed"
	DemoDeepForked EventType = "demo_deep_forked"
	DemoForkReused EventType = "demo_fork_reused"

	// Resources
	PlaceCreated    EventType = "place_created"
	ActivityCreated EventType = "activity_created"
	ExpenseCreated  EventType = "expense_created"

	// Session & Auth
	SessionStarted EventType = "session_started"
	UserRegistered EventType = "user_registered"
	UserMigrated   EventType = "user_migrated"

	// System
	SystemError EventType = "system_error"
)

var ValidTypes = map[EventType]bool{
	LandingViewed:   true,
	TripsListViewed: true,
	TripCreated:     true,
	TripViewed:      true,
	TripUpdated:     true,
	TripDeleted:     true,
	DemoViewed:      true,
	DemoDeepForked:  true,
	DemoForkReused:  true,
	PlaceCreated:    true,
	ActivityCreated: true,
	ExpenseCreated:  true,
	SessionStarted:  true,
	UserRegistered:  true,
	UserMigrated:    true,
	SystemError:     true,
}

func IsValid(t string) bool {
	_, ok := ValidTypes[EventType(t)]
	return ok
}