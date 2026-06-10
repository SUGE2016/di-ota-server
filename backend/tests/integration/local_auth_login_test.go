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
	"golang.org/x/crypto/bcrypt"
)

func TestLocalAuthLogin_CreatedUser(t *testing.T) {
	_, mock, r := newTestRouter(t, defaultTestConfig())

	password := "Operator@123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}

	mockLocalUserLogin(mock, "operator", string(hash), "user-operator-1")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"username":"operator","password":"Operator@123"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", w.Code, w.Body.String())
	}
	var resp struct {
		Code int `json:"code"`
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json: %v", err)
	}
	if resp.Code != 0 || resp.Data.AccessToken == "" {
		t.Fatalf("unexpected login response: %s", w.Body.String())
	}
}

func TestLocalAuthLogin_UnknownUser(t *testing.T) {
	_, mock, r := newTestRouter(t, defaultTestConfig())

	mock.ExpectQuery(`FROM t_user u`).
		WithArgs("missing").
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "username", "display_name", "password_hash", "status", "auth_source",
			"last_login_at", "last_operation_at", "created_at", "updated_at", "roles",
		}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"username":"missing","password":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("login status = %d, body = %s", w.Code, w.Body.String())
	}
}

func TestLocalAuthLogin_DisabledUser(t *testing.T) {
	_, mock, r := newTestRouter(t, defaultTestConfig())
	now := time.Now()

	mock.ExpectQuery(`FROM t_user u`).
		WithArgs("disabled").
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "username", "display_name", "password_hash", "status", "auth_source",
			"last_login_at", "last_operation_at", "created_at", "updated_at", "roles",
		}).AddRow("user-disabled", "disabled", "disabled", "$2a$12$A3La56.CqRH4oiOoMjnsGuwcgPv.h5xByKaYYGK/tfi4FWtbS9V4S", "disabled", "local", nil, now, now, now, []byte(`[]`)))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"username":"disabled","password":"Admin@123456"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("login status = %d, body = %s", w.Code, w.Body.String())
	}
}
