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
	"ota-server/backend/internal/config"
)

func TestCatalogSync_AcceptDevice(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Integration.Enabled = true
	cfg.Integration.ServiceToken = "integration-token"
	_, mock, r := newTestRouter(t, cfg)

	mock.ExpectQuery(`FROM t_catalog_sync_batch`).WithArgs("batch-1").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`SELECT device_id, device_group`).WithArgs("AMS000001").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`INSERT INTO t_device`).WillReturnRows(sqlmock.NewRows([]string{
		"device_id", "device_group", "product_model", "hardware_version", "product_code", "tags",
		"current_version", "reported_version", "catalog_version", "catalog_synced_at", "catalog_source",
		"eligibility_state", "inconsistency_flags", "last_seen_at", "registered_at", "last_heartbeat",
		"secret_provisioned",
	}).AddRow("AMS000001", "org-1001", "V9", "1.0", "AMS", []byte(`{}`), "v2.3.0", "", "v2.3.0", time.Now(), "backend", "active", []byte(`[]`), nil, time.Now(), time.Now(), false))
	mock.ExpectExec(`INSERT INTO t_catalog_sync_batch`).WillReturnResult(sqlmock.NewResult(0, 1))

	body := `{
	  "batch_id":"batch-1",
	  "mode":"upsert",
	  "devices":[{
	    "device_id":"AMS000001",
	    "product_code":"AMS",
	    "product_model":"V9",
	    "hardware_version":"1.0",
	    "current_version":"v2.3.0",
	    "device_group":"org-1001"
	  }]
	}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/catalog/sync", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer integration-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if int(data["accepted"].(float64)) != 1 {
		t.Fatalf("accepted=%v", data["accepted"])
	}
}

func TestCatalogSync_Unauthorized(t *testing.T) {
	cfg := &config.Config{
		API: config.APIConfig{Port: "8080"},
		Integration: config.IntegrationConfig{Enabled: true, ServiceToken: "secret"},
	}
	_, _, r := newTestRouter(t, cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/catalog/sync", bytes.NewBufferString(`{"batch_id":"b1","devices":[{"device_id":"x","product_code":"A","product_model":"M","hardware_version":"1"}]}`))
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", w.Code)
	}
}
