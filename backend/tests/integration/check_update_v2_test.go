//go:build !e2e

package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDeviceCheckUpdate_ReadOnly_AlreadyLatestWhenStoredAheadOfRequest(t *testing.T) {
	cfg := defaultTestConfig()
	_, mock, r := newTestRouter(t, cfg)

	mockDeviceRegistry(mock, "AMS000001", "org-1001", "V9", "1.0", "v2.4.0", "v2.4.0")
	mockRunningTasksForDevice(mock, "AMS000001", "task-1", "pkg-1")
	mockPackageDetail(mock, "pkg-1", "v2.4.0")

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
	data := resp["data"].(map[string]interface{})
	if data["reason"] != "already_latest" {
		t.Fatalf("reason = %v, want already_latest", data["reason"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations (check-update must not write device): %v", err)
	}
}

func TestDeviceCheckUpdate_HasUpdate_DownloadURLUsesPublicHost(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.S3 = presignTestS3Config()
	_, mock, r := newTestRouter(t, cfg)

	mockDeviceRegistry(mock, "AMS000001", "org-1001", "V9", "1.0", "v2.3.0", "v2.3.0")
	mockRunningTasksForDevice(mock, "AMS000001", "task-1", "pkg-test")
	mockPackageDetail(mock, "pkg-test", "v2.4.0")

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
	if int(resp["code"].(float64)) != 0 {
		t.Fatalf("code = %v, want 0", resp["code"])
	}
	data := resp["data"].(map[string]interface{})
	if data["has_update"] != true {
		t.Fatalf("has_update = %v, want true", data["has_update"])
	}
	downloadURL, _ := data["download_url"].(string)
	if strings.Contains(downloadURL, "minio:9000") {
		t.Fatalf("download_url should not use internal host: %s", downloadURL)
	}
	if !strings.HasPrefix(downloadURL, "http://localhost:9000/ota-packages/ota/pkg-test") {
		t.Fatalf("unexpected download_url: %s", downloadURL)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
