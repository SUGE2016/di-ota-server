//go:build !e2e

package integration_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestDeviceReportStatus_HappyPath_FirstReport(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Auth.DeviceAPIAuthEnabled = false
	_, mock, r := newTestRouter(t, cfg)

	mock.ExpectQuery(`SELECT idem_key`).WithArgs("AMS000001:task-1:Downloading").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`SELECT status`).WithArgs("AMS000001", "task-1").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`INSERT INTO t_upgrade_record`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "device_id", "task_id", "status", "created_at", "source_version", "target_version", "error_code",
		}).AddRow(int64(1), "AMS000001", "task-1", "Downloading", time.Now(), "v2.3", "v2.4", ""))
	mock.ExpectQuery(`INSERT INTO t_idempotency`).
		WillReturnRows(sqlmock.NewRows([]string{"idem_key", "response", "created_at"}).
			AddRow("AMS000001:task-1:Downloading", []byte(`{"code":0}`), time.Now()))

	body := `{"device_id":"AMS000001","task_id":"task-1","status":"downloading","source_version":"v2.3","target_version":"v2.4"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/device/v1/report-status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if int(resp["code"].(float64)) != 0 {
		t.Fatalf("code = %v, want 0", resp["code"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestDeviceReportStatus_InvalidTransition(t *testing.T) {
	cfg := defaultTestConfig()
	_, mock, r := newTestRouter(t, cfg)

	mock.ExpectQuery(`SELECT idem_key`).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`SELECT status`).WithArgs("AMS000001", "task-1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("Success"))

	body := `{"device_id":"AMS000001","task_id":"task-1","status":"downloading"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/device/v1/report-status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if int(resp["code"].(float64)) != 2005 {
		t.Fatalf("code = %v, want 2005", resp["code"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
