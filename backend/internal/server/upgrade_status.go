package server

import "strings"

// upgradeStatusAction is the server decision for a report-status transition.
type upgradeStatusAction string

const (
	upgradeStatusAllow  upgradeStatusAction = "allow"
	upgradeStatusReject upgradeStatusAction = "reject"
	upgradeStatusIgnore upgradeStatusAction = "ignore" // keep previous; HTTP 200
)

func decideUpgradeStatusAction(prev, next, mode string) upgradeStatusAction {
	prev = strings.TrimSpace(prev)
	next = strings.TrimSpace(next)
	if next == "" {
		return upgradeStatusReject
	}
	if prev == "" || prev == next {
		return upgradeStatusAllow
	}

	// Terminal Success: only re-report Success; intermediate states ignored (both modes).
	if prev == "Success" {
		if next == "Success" {
			return upgradeStatusAllow
		}
		return upgradeStatusIgnore
	}

	if strings.EqualFold(strings.TrimSpace(mode), "strict") {
		if canTransitUpgradeStatusStrict(prev, next) {
			return upgradeStatusAllow
		}
		return upgradeStatusReject
	}

	// relaxed: allow reset / Failed→Success / forward jumps
	return upgradeStatusAllow
}

// canTransitUpgradeStatus keeps historical name for tests; implements strict rules.
func canTransitUpgradeStatus(prev, next string) bool {
	return canTransitUpgradeStatusStrict(prev, next)
}

func canTransitUpgradeStatusStrict(prev, next string) bool {
	if next == "" {
		return false
	}
	if strings.TrimSpace(prev) == "" {
		return true
	}
	if prev == next {
		return true
	}

	if prev == "Success" || prev == "RolledBack" || prev == "RollbackFailed" {
		return false
	}
	if next == "Failed" {
		return true
	}
	if prev == "Failed" {
		return next == "Rollbacking" || next == "RolledBack" || next == "RollbackFailed"
	}
	if prev == "Rollbacking" {
		return next == "RolledBack" || next == "RollbackFailed"
	}

	order := map[string]int{
		"Pending":     0,
		"Downloading": 1,
		"Downloaded":  2,
		"Verifying":   3,
		"Upgrading":   4,
		"Success":     5,
	}
	p, okP := order[prev]
	n, okN := order[next]
	if !okP || !okN {
		return false
	}
	return n >= p
}
