//go:build e2e

package integration_test

import "testing"

// E2E 占位：需 docker compose 环境，运行方式：
//   cd backend && go test -tags=e2e ./tests/integration/ -run TestE2E -v
func TestE2E_ReleaseFlow_Placeholder(t *testing.T) {
	t.Skip("requires docker compose stack; see tests/specs/release_task_flow.md")
}
