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
			"eligibility_state", "inconsistency_flags", "last_seen_at", "registered_at", "last_heartbeat",
		}).AddRow("AMS000001", "org-1001", "V9", "1.0", "AMS", []byte(`{}`), "v2.4.0", "v2.4.0", "v2.4.0", now, "csv", "active", []byte(`[]`), now, now, now))

	mock.ExpectQuery(`UPDATE t_device SET`).WillReturnRows(sqlmock.NewRows([]string{
		"device_id", "device_group", "product_model", "hardware_version", "product_code", "tags",
		"current_version", "reported_version", "catalog_version", "catalog_synced_at", "catalog_source",
		"eligibility_state", "inconsistency_flags", "last_seen_at", "registered_at", "last_heartbeat",
	}).AddRow("AMS000001", "org-1001", "V9", "1.0", "AMS", []byte(`{}`), "v2.4.0", "v2.4.0", "v2.1.0", now, "backend", "active", []byte(`["version_rollback_ignored"]`), now, now, now))

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

func TestBuildTaskSnapshot_InsertTaskTargets(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	q := store.New(db)

	task := store.TReleaseTask{
		TaskID:          "task-1",
		TargetGroup:     "org-1001",
		ProductModel:    "V9",
		HardwareVersion: "1.0",
	}

	mock.ExpectQuery(`SELECT device_id FROM t_device`).
		WithArgs("org-1001", "V9", "1.0").
		WillReturnRows(sqlmock.NewRows([]string{"device_id"}).
			AddRow("AMS000001").
			AddRow("AMS000002"))
	mock.ExpectExec(`INSERT INTO t_task_target`).
		WithArgs("task-1", "AMS000001").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO t_task_target`).
		WithArgs("task-1", "AMS000002").
		WillReturnResult(sqlmock.NewResult(0, 1))

	n, err := buildTaskSnapshot(context.Background(), q, task)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("matched = %d, want 2", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
