package tracking

import "strings"

var AllowedMetadataKeys = map[string]bool{
	"demo_name": true, "original_demo": true, "trigger": true,
	"hours_since_first_fork": true, "time_to_fork_ms": true,
	"name_length": true, "has_dates": true, "currency": true,
	"days_duration": true, "days_old": true, "fields_changed": true,
	"had_places": true, "had_expenses": true,
	"has_coords": true, "has_time": true, "has_notes": true,
	"has_place": true, "has_category": true, "amount": true,
	"landing_source": true, "trips_before_register": true,
	"days_as_guest": true, "trips_transferred": true,
	"forks_transferred": true,
	"error_message": true, "stack_trace": true, "url": true,
	"legacy_source": true, "migrated_at": true,
}

func SanitizeMetadata(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return make(map[string]interface{})
	}
	result := make(map[string]interface{})
	for k, v := range m {
		if strings.HasPrefix(k, "client.") || strings.HasPrefix(k, "server.") {
			result[k] = v
			continue
		}
		if AllowedMetadataKeys[k] {
			result[k] = v
		}
	}
	return result
}