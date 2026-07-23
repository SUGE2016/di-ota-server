package server

import "testing"

func TestNormalizeUpgradeStatus_Aliases(t *testing.T) {
	cases := map[string]string{
		"downloading":      "Downloading",
		"download_success": "Downloaded",
		"upgrade_success":  "Success",
		"rolling_back":     "Rollbacking",
		"rollback_success": "RolledBack",
		"PENDING":          "Pending",
	}
	for in, want := range cases {
		got, ok := normalizeUpgradeStatus(in)
		if !ok {
			t.Fatalf("normalizeUpgradeStatus(%q) ok=false", in)
		}
		if got != want {
			t.Fatalf("normalizeUpgradeStatus(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeUpgradeStatus_Invalid(t *testing.T) {
	if _, ok := normalizeUpgradeStatus("not-a-state"); ok {
		t.Fatal("normalizeUpgradeStatus(invalid) ok=true, want false")
	}
}

func TestCanTransitUpgradeStatus_FirstReport(t *testing.T) {
	if !canTransitUpgradeStatus("", "Pending") {
		t.Fatal("empty prev should allow first report")
	}
}

func TestCanTransitUpgradeStatus_Idempotent(t *testing.T) {
	if !canTransitUpgradeStatus("Downloading", "Downloading") {
		t.Fatal("same status should be allowed")
	}
}

func TestCanTransitUpgradeStatus_Forward(t *testing.T) {
	if !canTransitUpgradeStatus("Downloading", "Downloaded") {
		t.Fatal("forward transition should be allowed")
	}
}

func TestCanTransitUpgradeStatus_Backward(t *testing.T) {
	if canTransitUpgradeStatus("Upgrading", "Downloading") {
		t.Fatal("backward transition should be rejected")
	}
}

func TestCanTransitUpgradeStatus_TerminalSuccess(t *testing.T) {
	if canTransitUpgradeStatus("Success", "Failed") {
		t.Fatal("Success is terminal under strict canTransit")
	}
}

func TestCanTransitUpgradeStatus_FailedToRollback(t *testing.T) {
	if !canTransitUpgradeStatus("Failed", "Rollbacking") {
		t.Fatal("Failed -> Rollbacking should be allowed")
	}
}

func TestCanTransitUpgradeStatus_AnyToFailed(t *testing.T) {
	if !canTransitUpgradeStatus("Verifying", "Failed") {
		t.Fatal("any stage can transition to Failed")
	}
}

func TestCanTransitUpgradeStatus_Rollbacking(t *testing.T) {
	if !canTransitUpgradeStatus("Rollbacking", "RolledBack") {
		t.Fatal("Rollbacking -> RolledBack should be allowed")
	}
	if canTransitUpgradeStatus("Rollbacking", "Downloading") {
		t.Fatal("Rollbacking -> Downloading should be rejected")
	}
}

func TestDecideUpgradeStatusAction_RelaxedResetAndFailedToSuccess(t *testing.T) {
	if decideUpgradeStatusAction("Upgrading", "Pending", "relaxed") != upgradeStatusAllow {
		t.Fatal("relaxed should allow reset")
	}
	if decideUpgradeStatusAction("Failed", "Success", "relaxed") != upgradeStatusAllow {
		t.Fatal("relaxed should allow Failed -> Success")
	}
}

func TestDecideUpgradeStatusAction_StrictRejectsReset(t *testing.T) {
	if decideUpgradeStatusAction("Upgrading", "Pending", "strict") != upgradeStatusReject {
		t.Fatal("strict should reject reset")
	}
	if decideUpgradeStatusAction("Failed", "Success", "strict") != upgradeStatusReject {
		t.Fatal("strict should reject Failed -> Success")
	}
}

func TestDecideUpgradeStatusAction_SuccessThenIntermediateIgnored(t *testing.T) {
	if decideUpgradeStatusAction("Success", "Downloading", "relaxed") != upgradeStatusIgnore {
		t.Fatal("Success -> Downloading should be ignored")
	}
	if decideUpgradeStatusAction("Success", "Downloading", "strict") != upgradeStatusIgnore {
		t.Fatal("Success -> Downloading should be ignored in strict too")
	}
	if decideUpgradeStatusAction("Success", "Success", "relaxed") != upgradeStatusAllow {
		t.Fatal("Success -> Success should allow")
	}
}
