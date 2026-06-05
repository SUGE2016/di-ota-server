package server

import "testing"

func TestInCanaryRange_Boundaries(t *testing.T) {
	if inCanaryRange("dev-1", "task-1", 0) {
		t.Fatal("0% should not hit")
	}
	if !inCanaryRange("dev-1", "task-1", 100) {
		t.Fatal("100% should always hit")
	}
}

func TestInCanaryRange_Deterministic(t *testing.T) {
	first := inCanaryRange("AMS000001", "task-abc", 50)
	for i := 0; i < 10; i++ {
		if got := inCanaryRange("AMS000001", "task-abc", 50); got != first {
			t.Fatalf("inCanaryRange not deterministic: iter %d got %v want %v", i, got, first)
		}
	}
}

func TestInCanaryRange_DifferentDevices(t *testing.T) {
	hits := 0
	for i := 0; i < 100; i++ {
		id := "device-" + string(rune('A'+i%26)) + string(rune('0'+i/26))
		if inCanaryRange(id, "task-fixed", 50) {
			hits++
		}
	}
	if hits == 0 || hits == 100 {
		t.Fatalf("expected mixed canary hits at 50%%, got %d/100", hits)
	}
}
