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

func mockDeviceRegistry(mock sqlmock.Sqlmock, deviceID, group, model, hw, reported, catalog string) {
	now := time.Now()
	mock.ExpectQuery(`SELECT device_id, device_group`).
		WithArgs(deviceID).
		WillReturnRows(sqlmock.NewRows([]string{
			"device_id", "device_group", "product_model", "hardware_version", "product_code", "tags",
			"current_version", "reported_version", "catalog_version", "catalog_synced_at", "catalog_source",
			"eligibility_state", "inconsistency_flags", "last_seen_at", "registered_at", "last_heartbeat",
		}).AddRow(deviceID, group, model, hw, "AMS", []byte(`{}`), reported, reported, catalog, now, "csv", "active", []byte(`[]`), now, now, now))
}

func TestDeviceCheckUpdate_AuthDisabled_NotInCatalog(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Auth.DeviceAPIAuthEnabled = false
	_, mock, r := newTestRouter(t, cfg)

	mock.ExpectQuery(`SELECT device_id, device_group`).
		WithArgs("AMS000001").
		WillReturnError(sql.ErrNoRows)

	body := `{"device_id":"AMS000001","group":"org-1001","product_model":"V9","hardware_version":"1.0","current_version":"v2.3"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/device/v1/check-update", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}

func TestDeviceCheckUpdate_AuthEnabled_MissingBearer(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Auth.DeviceAPIAuthEnabled = true
	_, _, r := newTestRouter(t, cfg)

	body := `{"device_id":"AMS000001","group":"org-1001","product_model":"V9","hardware_version":"1.0"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/device/v1/check-update", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestDeviceCheckUpdate_MissingRequiredFields(t *testing.T) {
	_, _, r := newTestRouter(t, defaultTestConfig())

	body := `{"device_id":"AMS000001","group":"org-1001"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/device/v1/check-update", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestDeviceCheckUpdate_NoRunningTask(t *testing.T) {
	_, mock, r := newTestRouter(t, defaultTestConfig())

	mockDeviceRegistry(mock, "AMS000001", "org-1001", "V9", "1.0", "v2.3.0", "v2.3.0")
	mock.ExpectQuery(`FROM t_release_task t`).WillReturnRows(sqlmock.NewRows([]string{
		"task_id", "package_id", "target_group", "product_model", "hardware_version",
		"failure_threshold", "state", "created_at", "canary_percent", "schedule_time", "force_upgrade",
	}))

	body := `{"device_id":"AMS000001","group":"org-1001","product_model":"V9","hardware_version":"1.0","current_version":"v2.3.0"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/device/v1/check-update", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json: %v", err)
	}
	if int(resp["code"].(float64)) != 2001 {
		t.Fatalf("code = %v, want 2001", resp["code"])
	}
}

func TestDeviceReportStatus_MissingTaskID(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Auth.DeviceAPIAuthEnabled = true
	_, _, r := newTestRouter(t, cfg)

	body := `{"device_id":"AMS000001","status":"downloading"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/device/v1/report-status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer shared-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestDeviceReportStatus_InvalidStatus(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Auth.DeviceAPIAuthEnabled = true
	_, _, r := newTestRouter(t, cfg)

	body := `{"device_id":"AMS000001","task_id":"task-1","status":"unknown"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/device/v1/report-status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer shared-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}
