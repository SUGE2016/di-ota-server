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
	"ota-server/backend/internal/server"
)

func mockDeviceRegistry(mock sqlmock.Sqlmock, deviceID, group, model, hw, reported, catalog string) {
	now := time.Now()
	mock.ExpectQuery(`SELECT device_id, device_group`).
		WithArgs(deviceID).
		WillReturnRows(sqlmock.NewRows([]string{
			"device_id", "device_group", "product_model", "hardware_version", "product_code", "tags",
			"current_version", "reported_version", "catalog_version", "catalog_synced_at", "catalog_source",
			"eligibility_state", "inconsistency_flags", "last_seen_at", "registered_at", "last_heartbeat",
			"secret_provisioned",
		}).AddRow(deviceID, group, model, hw, "AMS", []byte(`{}`), reported, reported, catalog, now, "csv", "active", []byte(`[]`), now, now, now, true))
}

func mockDeviceAuthCredential(mock sqlmock.Sqlmock, deviceID, secret, eligibility string) {
	mock.ExpectQuery(`SELECT device_id, device_secret, eligibility_state`).
		WithArgs(deviceID).
		WillReturnRows(sqlmock.NewRows([]string{"device_id", "device_secret", "eligibility_state"}).
			AddRow(deviceID, secret, eligibility))
}

func deviceAuthHeader(deviceID, secret, method, path string, body []byte) string {
	return server.BuildDeviceAuthorizationHeader(deviceID, secret, method, path, body, time.Now().Unix())
}

func TestDeviceCheckUpdate_AuthDisabled_NotInCatalog(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Auth.DeviceAPIAuthEnabled = false
	_, mock, r := newTestRouter(t, cfg)

	mock.ExpectQuery(`SELECT device_id, device_group`).
		WithArgs("AMS000001").
		WillReturnError(sql.ErrNoRows)

	body := `{"device_id":"AMS000001","current_version":"v2.3"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/device/v1/check-update", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}

func TestDeviceCheckUpdate_AuthEnabled_MissingSignature(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Auth.DeviceAPIAuthEnabled = true
	_, _, r := newTestRouter(t, cfg)

	body := `{"device_id":"AMS000001","current_version":"v2.3.0"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/device/v1/check-update", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestDeviceCheckUpdate_MissingDeviceID(t *testing.T) {
	_, _, r := newTestRouter(t, defaultTestConfig())

	body := `{"current_version":"v2.3.0"}`
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

	body := `{"device_id":"AMS000001","current_version":"v2.3.0"}`
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

func TestDeviceCheckUpdate_PerDeviceAuthSuccess(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Auth.DeviceAPIAuthEnabled = true
	_, mock, r := newTestRouter(t, cfg)

	body := []byte(`{"device_id":"AMS000001","current_version":"v2.3.0"}`)
	auth := deviceAuthHeader("AMS000001", "device-secret-1", http.MethodPost, "/device/v1/check-update", body)

	mockDeviceAuthCredential(mock, "AMS000001", "device-secret-1", "active")
	mockDeviceRegistry(mock, "AMS000001", "org-1001", "V9", "1.0", "v2.3.0", "v2.3.0")
	mock.ExpectQuery(`FROM t_release_task t`).WillReturnRows(sqlmock.NewRows([]string{
		"task_id", "package_id", "target_group", "product_model", "hardware_version",
		"failure_threshold", "state", "created_at", "canary_percent", "schedule_time", "force_upgrade",
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/device/v1/check-update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", auth)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}

func TestDeviceReportStatus_MissingTaskID(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Auth.DeviceAPIAuthEnabled = true
	_, mock, r := newTestRouter(t, cfg)

	body := []byte(`{"device_id":"AMS000001","status":"downloading"}`)
	auth := deviceAuthHeader("AMS000001", "device-secret-1", http.MethodPost, "/device/v1/report-status", body)
	mockDeviceAuthCredential(mock, "AMS000001", "device-secret-1", "active")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/device/v1/report-status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", auth)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestDeviceReportStatus_InvalidStatus(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Auth.DeviceAPIAuthEnabled = true
	_, mock, r := newTestRouter(t, cfg)

	body := []byte(`{"device_id":"AMS000001","task_id":"task-1","status":"unknown"}`)
	auth := deviceAuthHeader("AMS000001", "device-secret-1", http.MethodPost, "/device/v1/report-status", body)
	mockDeviceAuthCredential(mock, "AMS000001", "device-secret-1", "active")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/device/v1/report-status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", auth)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}

func TestSetDeviceSecret_NotFound(t *testing.T) {
	_, mock, r := newTestRouter(t, defaultTestConfig())
	mock.ExpectExec(`UPDATE t_device SET device_secret`).
		WithArgs("AMS000001", "secret-1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	body := `{"device_secret":"secret-1"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/devices/AMS000001/device-secret", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerHeader())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}
