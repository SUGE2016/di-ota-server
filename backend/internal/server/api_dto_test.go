package server

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/sqlc-dev/pqtype"
	"ota-server/backend/internal/store"
)

func TestNullTimeJSON_InvalidReturnsNil(t *testing.T) {
	if v := nullTimeJSON(sql.NullTime{}); v != nil {
		t.Fatalf("expected nil, got %#v", v)
	}
	now := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)
	if v := nullTimeJSON(sql.NullTime{Time: now, Valid: true}); v != now {
		t.Fatalf("expected time, got %#v", v)
	}
}

func TestMapReleaseTaskListRows_OmitsInvalidNullFields(t *testing.T) {
	rows := []store.ListReleaseTasksRow{{
		TaskID:           "task-1",
		PackageID:        "pkg-1",
		TargetGroup:      "org-1001",
		ProductModel:     "V9",
		HardwareVersion:  "1.0",
		FailureThreshold: "0.05",
		State:            "Running",
		CreatedAt:        time.Now(),
		ProductCode:      sql.NullString{String: "AMS", Valid: true},
		Version:          sql.NullString{String: "v2.4.0", Valid: true},
	}}
	out := mapReleaseTaskListRows(rows)
	raw, err := json.Marshal(out[0])
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) == "" {
		t.Fatal("empty json")
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["version"] != "v2.4.0" {
		t.Fatalf("version = %#v", m["version"])
	}
	if _, ok := m["Valid"]; ok {
		t.Fatalf("must not serialize NullString wrapper: %#v", m)
	}
}

func TestMapTReleaseTask_ScheduleTimeNull(t *testing.T) {
	task := store.TReleaseTask{
		TaskID:    "task-1",
		PackageID: "pkg-1",
		State:     "Draft",
		CreatedAt: time.Now(),
	}
	out := mapTReleaseTask(task)
	raw, _ := json.Marshal(out)
	var m map[string]interface{}
	_ = json.Unmarshal(raw, &m)
	if v, ok := m["schedule_time"]; !ok || v != nil {
		t.Fatalf("schedule_time = %#v, want null", v)
	}
}

func TestMapAuditLog_PlainFields(t *testing.T) {
	log := store.TAuditLog{
		ID:            1,
		TraceID:       "trace-1",
		Operator:      "admin",
		OperationType: "RESUME",
		ResourceID:    "task-1",
		BeforeState:   pqtype.NullRawMessage{Valid: false},
		AfterState:    pqtype.NullRawMessage{Valid: false},
		CreatedAt:     time.Now(),
	}
	raw, err := json.Marshal(mapAuditLog(log))
	if err != nil {
		t.Fatal(err)
	}
	if err := assertNoNullStringJSON(raw); err != nil {
		t.Fatal(err)
	}
}

func assertNoNullStringJSON(raw []byte) error {
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return err
	}
	return walkJSON(v)
}

func walkJSON(v interface{}) error {
	switch x := v.(type) {
	case map[string]interface{}:
		if _, ok := x["String"]; ok {
			if _, ok2 := x["Valid"]; ok2 {
				return errNullStringShape
			}
		}
		for _, child := range x {
			if err := walkJSON(child); err != nil {
				return err
			}
		}
	case []interface{}:
		for _, child := range x {
			if err := walkJSON(child); err != nil {
				return err
			}
		}
	}
	return nil
}

var errNullStringShape = &nullStringJSONError{}

type nullStringJSONError struct{}

func (e *nullStringJSONError) Error() string { return "sql.NullString-like JSON object" }
