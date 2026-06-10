//go:build !e2e

package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAdminDevicesList_FilterByGroupAndModel(t *testing.T) {
	_, mock, r := newTestRouter(t, defaultTestConfig())
	now := time.Now()

	mock.ExpectQuery(`FROM t_device`).
		WithArgs("", "org-1001", "V9", "", "", false, int32(10), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"device_id", "device_group", "product_model", "hardware_version", "product_code", "tags",
			"current_version", "reported_version", "catalog_version", "catalog_synced_at", "catalog_source",
			"eligibility_state", "inconsistency_flags", "last_seen_at", "registered_at", "last_heartbeat",
			"secret_provisioned",
		}).AddRow("AMS000001", "org-1001", "V9", "1.0", "AMS", []byte(`{}`), "v2.3", "v2.3", "v2.3", now, "csv", "active", []byte(`[]`), now, now, now, false))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM t_device`).
		WithArgs("", "org-1001", "V9", "", "", false).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices?limit=10&group=org-1001&product_model=V9", nil)
	req.Header.Set("Authorization", bearerHeader())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data struct {
			Total   int64                    `json:"total"`
			Devices []map[string]interface{} `json:"devices"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json: %v", err)
	}
	if resp.Data.Total != 1 || len(resp.Data.Devices) != 1 {
		t.Fatalf("unexpected list payload: %+v", resp.Data)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminDevicesList_AbnormalFilter(t *testing.T) {
	_, mock, r := newTestRouter(t, defaultTestConfig())

	mock.ExpectQuery(`FROM t_device`).
		WithArgs("", "", "", "", "", true, int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"device_id", "device_group", "product_model", "hardware_version", "product_code", "tags",
			"current_version", "reported_version", "catalog_version", "catalog_synced_at", "catalog_source",
			"eligibility_state", "inconsistency_flags", "last_seen_at", "registered_at", "last_heartbeat",
			"secret_provisioned",
		}))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM t_device`).
		WithArgs("", "", "", "", "", true).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices?abnormal=true", nil)
	req.Header.Set("Authorization", bearerHeader())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminDeviceUpgradeRecords(t *testing.T) {
	_, mock, r := newTestRouter(t, defaultTestConfig())
	now := time.Now()

	mockDeviceRegistry(mock, "AMS000001", "org-1001", "V9", "1.0", "v2.4.0", "v2.4.0")
	mock.ExpectQuery(`FROM t_upgrade_record`).
		WithArgs("AMS000001", int32(50)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "device_id", "task_id", "status", "created_at", "source_version", "target_version", "error_code",
		}).AddRow(int64(1), "AMS000001", "task-1", "Success", now, "v2.3.0", "v2.4.0", ""))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices/AMS000001/upgrade-records", nil)
	req.Header.Set("Authorization", bearerHeader())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data struct {
			Records []map[string]interface{} `json:"records"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(resp.Data.Records) != 1 {
		t.Fatalf("records = %+v", resp.Data.Records)
	}
	if resp.Data.Records[0]["status"] != "Success" {
		t.Fatalf("status = %v", resp.Data.Records[0]["status"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
