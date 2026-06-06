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

func TestDeviceReportStatus_IdempotentSuccess_StillUpdatesReportedVersion(t *testing.T) {
	cfg := defaultTestConfig()
	_, mock, r := newTestRouter(t, cfg)

	cached := `{"code":0,"message":"Status received","data":{"status":"Success"}}`
	mock.ExpectQuery(`SELECT idem_key`).
		WithArgs("AMS000001:task-1:Success").
		WillReturnRows(sqlmock.NewRows([]string{"idem_key", "response", "created_at"}).
			AddRow("AMS000001:task-1:Success", []byte(cached), time.Now()))
	mockDeviceRegistry(mock, "AMS000001", "org-1001", "V9", "1.0", "v2.3.0", "v2.3.0")
	mock.ExpectExec(`UPDATE t_device SET`).
		WithArgs("AMS000001", "v2.4.0").
		WillReturnResult(sqlmock.NewResult(0, 1))

	body := `{"device_id":"AMS000001","task_id":"task-1","status":"success","target_version":"v2.4.0"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/device/v1/report-status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
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

func TestDeviceReportStatus_IdempotentSuccess_SkipsUpdateWhenAlreadyLatest(t *testing.T) {
	cfg := defaultTestConfig()
	_, mock, r := newTestRouter(t, cfg)

	cached := `{"code":0,"message":"Status received","data":{"status":"Success"}}`
	mock.ExpectQuery(`SELECT idem_key`).
		WithArgs("AMS000001:task-1:Success").
		WillReturnRows(sqlmock.NewRows([]string{"idem_key", "response", "created_at"}).
			AddRow("AMS000001:task-1:Success", []byte(cached), time.Now()))
	mockDeviceRegistry(mock, "AMS000001", "org-1001", "V9", "1.0", "v2.4.0", "v2.4.0")

	body := `{"device_id":"AMS000001","task_id":"task-1","status":"success","target_version":"v2.4.0"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/device/v1/report-status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
