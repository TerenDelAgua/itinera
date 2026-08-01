package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// testToken is long enough to satisfy any future length validation.
const testInternalToken = "test-token-must-be-long-enough-32-chars-abc"

func TestIsInternalSession(t *testing.T) {
	// Set the env once for the whole suite so each case uses the same
	// reference value, then clean up at the end.
	if err := os.Setenv("ITINERA_INTERNAL_TOKEN", testInternalToken); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	defer os.Unsetenv("ITINERA_INTERNAL_TOKEN")

	cases := []struct {
		name    string
		headers map[string]string
		cookies []*http.Cookie
		want    bool
	}{
		{
			name: "no header, no cookie -> false",
			want: false,
		},
		{
			name:    "valid header -> true",
			headers: map[string]string{"X-Internal-Session": testInternalToken},
			want:    true,
		},
		{
			name:    "invalid header -> false",
			headers: map[string]string{"X-Internal-Session": "wrong-token"},
			want:    false,
		},
		{
			name:    "valid cookie -> true",
			cookies: []*http.Cookie{{Name: "itinera_internal", Value: testInternalToken}},
			want:    true,
		},
		{
			name:    "invalid cookie -> false",
			cookies: []*http.Cookie{{Name: "itinera_internal", Value: "wrong"}},
			want:    false,
		},
		{
			name:    "header takes precedence over cookie",
			headers: map[string]string{"X-Internal-Session": testInternalToken},
			cookies: []*http.Cookie{{Name: "itinera_internal", Value: "wrong"}},
			want:    true,
		},
		{
			name:    "empty header value falls back to cookie",
			headers: map[string]string{"X-Internal-Session": ""},
			cookies: []*http.Cookie{{Name: "itinera_internal", Value: testInternalToken}},
			want:    true,
		},
		{
			name:    "empty cookie value falls back to false",
			cookies: []*http.Cookie{{Name: "itinera_internal", Value: ""}},
			want:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			for k, v := range tc.headers {
				r.Header.Set(k, v)
			}
			for _, c := range tc.cookies {
				r.AddCookie(c)
			}

			got := IsInternalSession(r)
			if got != tc.want {
				t.Fatalf("IsInternalSession() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestIsInternalSession_EmptyEnv covers the fail-closed contract:
// if the operator forgets to set ITINERA_INTERNAL_TOKEN, the helper
// must return false for every request, regardless of header/cookie.
func TestIsInternalSession_EmptyEnv(t *testing.T) {
	os.Unsetenv("ITINERA_INTERNAL_TOKEN")

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Internal-Session", testInternalToken)
	r.AddCookie(&http.Cookie{Name: "itinera_internal", Value: testInternalToken})

	if IsInternalSession(r) {
		t.Fatal("expected false when env is unset, got true")
	}
}
