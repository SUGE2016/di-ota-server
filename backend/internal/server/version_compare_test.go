package server

import "testing"

func TestPolicyReportedVersion_NoDowngradeFromRequest(t *testing.T) {
	got := policyReportedVersion("v2.4", "v2.3", "v2.3.0")
	if got != "v2.4" {
		t.Fatalf("expected stored v2.4, got %q", got)
	}
}

func TestPolicyReportedVersion_RequestAhead(t *testing.T) {
	got := policyReportedVersion("v2.3", "v2.3", "v2.5")
	if got != "v2.5" {
		t.Fatalf("expected request v2.5, got %q", got)
	}
}

func TestPolicyReportedVersion_FallbackCatalog(t *testing.T) {
	got := policyReportedVersion("", "v2.3", "")
	if got != "v2.3" {
		t.Fatalf("expected catalog v2.3, got %q", got)
	}
}
