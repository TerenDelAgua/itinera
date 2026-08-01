package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"backend/internal/http/middleware"
)

// TestTrackEvent_FiltersInternalConversionEvents locks in the spec
// contract: events from an internal session are dropped for
// conversion events (trip_created, demo_deep_forked, demo_fork_reused)
// and kept for behavioural events (landing_viewed, demo_viewed).
//
// We test the gating predicate directly because the handler has many
// dependencies (rate limiter, repo, JWT context) that are not under
// test here. The predicate is the contract the spec is concerned with.
func TestTrackEvent_FiltersInternalConversionEvents(t *testing.T) {
	const token = "test-internal-token-for-filter-32chars"

	conversionEvents := map[string]bool{
		"trip_created":     true,
		"demo_deep_forked": true,
		"demo_fork_reused": true,
	}

	cases := []struct {
		name      string
		setHeader bool
		eventType string
		want      bool // whether the filter triggers (skips insertion)
	}{
		{"internal+conversion -> dropped", true, "trip_created", true},
		{"internal+conversion(deep_fork) -> dropped", true, "demo_deep_forked", true},
		{"internal+conversion(reused) -> dropped", true, "demo_fork_reused", true},
		{"internal+behaviour -> kept", true, "landing_viewed", false},
		{"internal+behaviour(demo_viewed) -> kept", true, "demo_viewed", false},
		{"real+conversion -> kept", false, "trip_created", false},
		{"real+behaviour -> kept", false, "landing_viewed", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setHeader {
				os.Setenv("ITINERA_INTERNAL_TOKEN", token)
			} else {
				os.Unsetenv("ITINERA_INTERNAL_TOKEN")
			}
			defer os.Unsetenv("ITINERA_INTERNAL_TOKEN")

			body, _ := json.Marshal(EventRequest{Type: tc.eventType})
			r := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
			if tc.setHeader {
				r.Header.Set("X-Internal-Session", token)
			}

			// Same predicate the handler uses.
			skip := middleware.IsInternalSession(r) && conversionEvents[tc.eventType]
			if skip != tc.want {
				t.Fatalf("filter skip = %v, want %v", skip, tc.want)
			}
		})
	}
}

// TestEventRequest_DecodingOk ensures the EventRequest shape used by
// the handler validates the JSON contract that the frontend depends on.
func TestEventRequest_DecodingOk(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		wantTyp string
		wantErr bool
	}{
		{"valid minimal", `{"type":"trip_created"}`, "trip_created", false},
		{"valid with metadata", `{"type":"landing_viewed","metadata":{"foo":"bar"}}`, "landing_viewed", false},
		{"invalid json", `{not json`, "", true},
		{"empty type", `{"type":""}`, "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var req EventRequest
			err := json.Unmarshal([]byte(tc.body), &req)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tc.wantErr)
			}
			if !tc.wantErr && req.Type != tc.wantTyp {
				t.Errorf("type = %q, want %q", req.Type, tc.wantTyp)
			}
		})
	}
}

// TestTrackEvent_NoInternalTokenNeverFilters is a guard: if the env var
// is missing, the filter is a no-op. This prevents misconfiguration
// from silently dropping legitimate events.
func TestTrackEvent_NoInternalTokenNeverFilters(t *testing.T) {
	os.Unsetenv("ITINERA_INTERNAL_TOKEN")

	r := httptest.NewRequest(http.MethodPost, "/events", nil)
	r.Header.Set("X-Internal-Session", "anyone-can-claim-this")

	if middleware.IsInternalSession(r) {
		t.Fatal("without env token, IsInternalSession must be false")
	}
}
