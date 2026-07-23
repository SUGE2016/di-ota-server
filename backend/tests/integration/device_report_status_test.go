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

func mockReportStatusModeRelaxed(mock sqlmock.Sqlmock, deviceID, model string) {
	mockDeviceRegistry(mock, deviceID, "org-1001", model, "1.0", "v2.3", "v2.3")
	mock.ExpectQuery(`SELECT report_status_mode FROM t_product_model_policy`).
		WithArgs(model).
		WillReturnError(sql.ErrNoRows)
}

func mockReportStatusModeStrict(mock sqlmock.Sqlmock, deviceID, model string) {
	mockDeviceRegistry(mock, deviceID, "org-1001", model, "1.0", "v2.3", "v2.3")
	mock.ExpectQuery(`SELECT report_status_mode FROM t_product_model_policy`).
		WithArgs(model).
		WillReturnRows(sqlmock.NewRows([]string{"report_status_mode"}).AddRow("strict"))
}

func TestDeviceReportStatus_HappyPath_FirstReport(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Auth.DeviceAPIAuthEnabled = false
	_, mock, r := newTestRouter(t, cfg)

	mock.ExpectQuery(`SELECT idem_key`).WithArgs("AMS000001:task-1:Downloading").
		WillReturnError(sql.ErrNoRows)
	mockReportStatusModeRelaxed(mock, "AMS000001", "V9")
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

func TestDeviceReportStatus_StrictInvalidTransition(t *testing.T) {
	cfg := defaultTestConfig()
	_, mock, r := newTestRouter(t, cfg)

	mock.ExpectQuery(`SELECT idem_key`).WillReturnError(sql.ErrNoRows)
	mockReportStatusModeStrict(mock, "AMS000001", "StrictModel")
	mock.ExpectQuery(`SELECT status`).WithArgs("AMS000001", "task-1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("Upgrading"))

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

func TestDeviceReportStatus_SuccessThenIntermediateIgnored(t *testing.T) {
	cfg := defaultTestConfig()
	_, mock, r := newTestRouter(t, cfg)

	mock.ExpectQuery(`SELECT idem_key`).WillReturnError(sql.ErrNoRows)
	mockReportStatusModeRelaxed(mock, "AMS000001", "V9")
	mock.ExpectQuery(`SELECT status`).WithArgs("AMS000001", "task-1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("Success"))

	body := `{"device_id":"AMS000001","task_id":"task-1","status":"downloading"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/device/v1/report-status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if int(resp["code"].(float64)) != 0 {
		t.Fatalf("code = %v, want 0", resp["code"])
	}
	data := resp["data"].(map[string]interface{})
	if data["ignored"] != true {
		t.Fatalf("expected ignored=true, got %v", data["ignored"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestDeviceReportStatus_RelaxedFailedToSuccess(t *testing.T) {
	cfg := defaultTestConfig()
	_, mock, r := newTestRouter(t, cfg)

	mock.ExpectQuery(`SELECT idem_key`).WillReturnError(sql.ErrNoRows)
	mockReportStatusModeRelaxed(mock, "AMS000001", "V9")
	mock.ExpectQuery(`SELECT status`).WithArgs("AMS000001", "task-1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("Failed"))
	mock.ExpectQuery(`SELECT task_id, package_id`).WithArgs("task-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"task_id", "package_id", "target_group", "product_model", "hardware_version",
			"failure_threshold", "state", "created_at", "canary_percent", "schedule_time", "force_upgrade",
		}).AddRow("task-1", "pkg-1", "org-1001", "V9", "1.0", "0.1", "Running", time.Now(), int32(100), nil, false))
	mock.ExpectQuery(`SELECT package_id, product_code, version, file_hash, signature, status, min_upgradable_version`).
		WithArgs("pkg-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"package_id", "product_code", "version", "file_hash", "signature", "status", "min_upgradable_version",
		}).AddRow("pkg-1", "AMS", "v2.4.0", "abc", "sig", "Published", ""))
	mock.ExpectQuery(`INSERT INTO t_upgrade_record`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "device_id", "task_id", "status", "created_at", "source_version", "target_version", "error_code",
		}).AddRow(int64(2), "AMS000001", "task-1", "Success", time.Now(), "v2.3", "v2.4.0", ""))
	mockDeviceRegistry(mock, "AMS000001", "org-1001", "V9", "1.0", "v2.3.0", "v2.3.0")
	mock.ExpectExec(`UPDATE t_device SET`).
		WithArgs("AMS000001", "v2.4.0").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`INSERT INTO t_alert`).
		WillReturnRows(sqlmock.NewRows([]string{
			"alert_id", "alert_type", "severity", "status", "resource_type", "resource_id",
			"message", "detail", "dedupe_key", "created_at", "updated_at",
		}).AddRow("alert-1", "设备升级状态", "info", "open", "device", "AMS000001",
			"msg", []byte(`{}`), "upgrade_status:AMS000001:task-1:Success", time.Now(), time.Now()))
	mock.ExpectQuery(`INSERT INTO t_idempotency`).
		WillReturnRows(sqlmock.NewRows([]string{"idem_key", "response", "created_at"}).
			AddRow("AMS000001:task-1:Success", []byte(`{"code":0}`), time.Now()))

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
