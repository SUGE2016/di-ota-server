package server

import "testing"

func TestCompareVersion(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v2.3.0", "v2.4.0", -1},
		{"v2.4.0", "v2.3.0", 1},
		{"2.3", "2.3.0", 0},
		{"v2.4.0", "v2.4.0", 0},
	}
	for _, tc := range cases {
		got := CompareVersion(tc.a, tc.b)
		if got != tc.want {
			t.Fatalf("CompareVersion(%q,%q)=%d want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestVersionAtLeast(t *testing.T) {
	if !versionAtLeast("v2.4.0", "v2.4.0") {
		t.Fatal("expected true")
	}
	if versionAtLeast("v2.3.0", "v2.4.0") {
		t.Fatal("expected false")
	}
}
