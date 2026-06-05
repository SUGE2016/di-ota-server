package server

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"ota-server/backend/internal/store"
)

func TestApplyCatalogDevice_VersionRollbackIgnored(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	q := store.New(db)
	now := time.Now()

	mock.ExpectQuery(`SELECT device_id, device_group`).WithArgs("AMS000001").
		WillReturnRows(sqlmock.NewRows([]string{
			"device_id", "device_group", "product_model", "hardware_version", "product_code", "tags",
			"current_version", "reported_version", "catalog_version", "catalog_synced_at", "catalog_source",
			"eligibility_state", "inconsistency_flags", "last_seen_at", "registered_at",
		}).AddRow("AMS000001", "org-1001", "V9", "1.0", "AMS", []byte(`{}`), "v2.4.0", "v2.4.0", "v2.4.0", now, "csv", "active", []byte(`[]`), now, now))

	mock.ExpectQuery(`UPDATE t_device SET`).WillReturnRows(sqlmock.NewRows([]string{
		"device_id", "device_group", "product_model", "hardware_version", "product_code", "tags",
		"current_version", "reported_version", "catalog_version", "catalog_synced_at", "catalog_source",
		"eligibility_state", "inconsistency_flags", "last_seen_at", "registered_at",
	}).AddRow("AMS000001", "org-1001", "V9", "1.0", "AMS", []byte(`{}`), "v2.4.0", "v2.4.0", "v2.1.0", now, "backend", "active", []byte(`["version_rollback_ignored"]`), now, now))

	warn, reject, err := applyCatalogDevice(context.Background(), q, "backend", CatalogSyncOptions{}, CatalogDeviceInput{
		DeviceID: "AMS000001", ProductCode: "AMS", ProductModel: "V9", HardwareVersion: "1.0",
		CurrentVersion: "v2.1.0", DeviceGroup: "org-1001", Tags: []byte(`{}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if reject != nil {
		t.Fatalf("reject=%+v", reject)
	}
	if warn.Code != "version_rollback_ignored" {
		t.Fatalf("warn=%+v", warn)
	}
}
