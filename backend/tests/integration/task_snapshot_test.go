//go:build !e2e

package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestReleaseTaskResume_BuildsSnapshot(t *testing.T) {
	cfg := defaultTestConfig()
	_, mock, r := newTestRouter(t, cfg)
	auth := loginBearer(t, r)

	taskID := "task-resume-1"
	now := time.Now()

	mock.ExpectQuery(`FROM t_release_task\s+WHERE task_id`).
		WithArgs(taskID).
		WillReturnRows(sqlmock.NewRows([]string{
			"task_id", "package_id", "target_group", "product_model", "hardware_version",
			"failure_threshold", "state", "created_at",
		}).AddRow(taskID, "pkg-1", "org-1001", "V9", "1.0", "0.05", "Paused", now))

	mock.ExpectQuery(`UPDATE t_release_task`).
		WithArgs(taskID, "Running").
		WillReturnRows(sqlmock.NewRows([]string{
			"task_id", "package_id", "target_group", "product_model", "hardware_version",
			"failure_threshold", "state", "created_at",
		}).AddRow(taskID, "pkg-1", "org-1001", "V9", "1.0", "0.05", "Running", now))

	mock.ExpectQuery(`SELECT device_id FROM t_device`).
		WithArgs("org-1001", "V9", "1.0").
		WillReturnRows(sqlmock.NewRows([]string{"device_id"}).
			AddRow("AMS000001").
			AddRow("AMS000002"))

	mock.ExpectExec(`INSERT INTO t_task_target`).
		WithArgs(taskID, "AMS000001").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO t_task_target`).
		WithArgs(taskID, "AMS000002").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(`INSERT INTO t_audit_log`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "trace_id", "operator", "operation_type", "resource_id", "before_state", "after_state", "created_at",
		}).AddRow(int64(1), "trace-1", "admin", "RESUME", taskID, []byte(`{}`), []byte(`{}`), now))

	mock.ExpectQuery(`FROM t_release_task\s+WHERE task_id = \$1`).
		WithArgs(taskID).
		WillReturnRows(sqlmock.NewRows([]string{
			"task_id", "package_id", "target_group", "product_model", "hardware_version",
			"failure_threshold", "state", "created_at", "canary_percent", "schedule_time", "force_upgrade",
		}).AddRow(taskID, "pkg-1", "org-1001", "V9", "1.0", "0.05", "Running", now, int32(100), nil, false))

	body := `{"action":"resume","reason":"integration test"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/release-tasks/"+taskID+"/actions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", auth)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json: %v", err)
	}
	if int(resp["code"].(float64)) != 0 {
		t.Fatalf("code = %v, want 0", resp["code"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
