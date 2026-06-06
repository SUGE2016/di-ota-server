package main

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"ota-server/backend/internal/config"
	"ota-server/backend/internal/store"
)

func testWorkerConfig(retention int64) *config.Config {
	return &config.Config{
		Worker: config.WorkerConfig{TaskStatsRetentionHours: retention},
	}
}

func TestRunWorkerCycle_WithRetention_ExecutesCleanup(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO t_task_stats").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("UPDATE t_release_task").
		WillReturnRows(sqlmock.NewRows([]string{"task_id", "failure_threshold", "failure_rate"}))
	mock.ExpectExec("DELETE FROM t_task_stats").WithArgs(int64(24)).WillReturnResult(sqlmock.NewResult(0, 3))

	if err := runWorkerCycle(db, store.New(db), testWorkerConfig(24)); err != nil {
		t.Fatalf("runWorkerCycle() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRunWorkerCycle_ZeroRetention_SkipsCleanup(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO t_task_stats").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("UPDATE t_release_task").
		WillReturnRows(sqlmock.NewRows([]string{"task_id", "failure_threshold", "failure_rate"}))

	if err := runWorkerCycle(db, store.New(db), testWorkerConfig(0)); err != nil {
		t.Fatalf("runWorkerCycle() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
