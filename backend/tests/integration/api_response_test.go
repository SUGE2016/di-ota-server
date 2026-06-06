package integration_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAdminReleaseTasksList_PlainStringVersion(t *testing.T) {
	_, mock, r := newTestRouter(t, defaultTestConfig())

	now := time.Now()
	mock.ExpectQuery(`FROM t_release_task rt`).
		WithArgs(int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"task_id", "package_id", "target_group", "product_model", "hardware_version",
			"failure_threshold", "state", "created_at", "product_code", "version",
		}).AddRow("task-1", "pkg-1", "org-1001", "V9", "1.0", "0.05", "Running", now, "AMS", "v2.4.0"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/release-tasks", nil)
	req.Header.Set("Authorization", bearerHeader())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("len(data) = %d, want 1", len(resp.Data))
	}
	version, ok := resp.Data[0]["version"].(string)
	if !ok || version != "v2.4.0" {
		t.Fatalf("version = %#v, want plain string v2.4.0", resp.Data[0]["version"])
	}
	if _, hasValid := resp.Data[0]["Valid"]; hasValid {
		t.Fatalf("response must not expose sql.NullString shape: %#v", resp.Data[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminUsersList_NullLastLoginAt(t *testing.T) {
	_, mock, r := newTestRouter(t, defaultTestConfig())

	now := time.Now()
	mock.ExpectQuery(`FROM t_user u`).
		WithArgs("", "", "", int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "username", "display_name", "password_hash", "status", "auth_source",
			"last_login_at", "last_operation_at", "created_at", "updated_at", "roles",
		}).AddRow("u-1", "admin", "Admin", "hash", "active", "local", sql.NullTime{}, now, now, now, []byte(`["admin"]`)))
	mock.ExpectQuery(`SELECT COUNT\(\*\)`).
		WithArgs("", "", "").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", bearerHeader())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	raw := w.Body.String()
	if raw == "" {
		t.Fatal("empty body")
	}
	if err := assertJSONNoNullStringShape([]byte(raw)); err != nil {
		t.Fatal(err)
	}

	var resp struct {
		Data struct {
			Users []map[string]interface{} `json:"users"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(resp.Data.Users) != 1 {
		t.Fatalf("len(users) = %d, want 1", len(resp.Data.Users))
	}
	if v, ok := resp.Data.Users[0]["last_login_at"]; !ok || v != nil {
		t.Fatalf("last_login_at = %#v, want null", v)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func assertJSONNoNullStringShape(raw []byte) error {
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return err
	}
	return walkNoNullStringShape(v, "$")
}

func walkNoNullStringShape(v interface{}, path string) error {
	switch x := v.(type) {
	case map[string]interface{}:
		if _, hasString := x["String"]; hasString {
			if _, hasValid := x["Valid"]; hasValid {
				return &nullStringShapeError{path: path}
			}
		}
		for k, child := range x {
			if err := walkNoNullStringShape(child, path+"."+k); err != nil {
				return err
			}
		}
	case []interface{}:
		for i, child := range x {
			if err := walkNoNullStringShape(child, path+"["+itoa(i)+"]"); err != nil {
				return err
			}
		}
	}
	return nil
}

type nullStringShapeError struct{ path string }

func (e *nullStringShapeError) Error() string {
	return "found sql.NullString-like object at " + e.path
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [20]byte
	n := len(b)
	for i > 0 {
		n--
		b[n] = byte('0' + i%10)
		i /= 10
	}
	return string(b[n:])
}
