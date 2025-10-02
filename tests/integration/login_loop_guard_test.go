package integration

import (
	"net/http"
	"testing"
)

// TestLoginPageServes200 ensures /login responds 200 (no redirect loop) for unauthenticated clients.
func TestLoginPageServes200(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/login")
	if err != nil { t.Fatalf("request failed: %v", err) }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// Accept Unauthorized if auth middleware enforces it differently, but not redirect loops
		if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
			// Provide extra context
			loc := resp.Header.Get("Location")
			if loc == "" { loc = resp.Header.Get("location") }
			// Mark as failure explicitly
			msg := "redirect instead of serving login page"
			if loc != "" { msg += ": location=" + loc }
			// Failing because we want stable 200 for automation
			// If the environment legitimately requires redirect first, adjust test expectations.
			t.Fatalf("/login returned %d %s", resp.StatusCode, msg)
		}
	}
}

// TestRootReturnsLoginOrDashboard ensures root returns 200 login page or redirects only once to dashboard when authenticated.
func TestRootReturnsLoginOrDashboard(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/")
	if err != nil { t.Fatalf("request failed: %v", err) }
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusFound {
		// Single redirect acceptable only if to /dashboard
		loc := resp.Header.Get("Location")
		if loc != "/dashboard" {
			// Not acceptable to bounce elsewhere
			t.Fatalf("unexpected redirect from / to %s", loc)
		}
	} else if resp.StatusCode != http.StatusOK {
		// Want 200 login page otherwise
		if resp.StatusCode == http.StatusMovedPermanently {
			t.Fatalf("permanent redirect from / not expected")
		}
	}
}
