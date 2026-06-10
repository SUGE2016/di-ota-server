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

func TestAlerts_ListAndAcknowledge(t *testing.T) {
	_, mock, r := newTestRouter(t, defaultTestConfig())
	bearer := loginBearer(t, mock, r)
	now := time.Now()

	mock.ExpectQuery(`FROM t_alert`).
		WithArgs("", "", int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"alert_id", "alert_type", "severity", "status", "resource_type", "resource_id",
			"message", "detail", "dedupe_key", "created_at", "updated_at",
		}).AddRow("alert-1", "upgrade_failed", "warning", "open", "device", "AMS000001",
			"upgrade failed", []byte(`{}`), "dedupe-1", now, now))
	mock.ExpectQuery(`COUNT\(\*\) FROM t_alert`).
		WithArgs("", "").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts", nil)
	req.Header.Set("Authorization", bearer)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", w.Code, w.Body.String())
	}

	mock.ExpectQuery(`UPDATE t_alert SET status`).
		WithArgs("alert-1", "acknowledged").
		WillReturnRows(sqlmock.NewRows([]string{
			"alert_id", "alert_type", "severity", "status", "resource_type", "resource_id",
			"message", "detail", "dedupe_key", "created_at", "updated_at",
		}).AddRow("alert-1", "upgrade_failed", "warning", "acknowledged", "device", "AMS000001",
			"upgrade failed", []byte(`{}`), "dedupe-1", now, now))

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/actions", bytes.NewBufferString(`{"action":"acknowledge","alert_ids":["alert-1"]}`))
	req2.Header.Set("Authorization", bearer)
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("ack status=%d body=%s", w2.Code, w2.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json: %v", err)
	}
	data := resp["data"].(map[string]any)
	if int(data["updated"].(float64)) != 1 {
		t.Fatalf("updated=%v want 1", data["updated"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
