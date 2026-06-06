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
	"github.com/gin-gonic/gin"
	"ota-server/backend/internal/config"
	"ota-server/backend/internal/server"
	"ota-server/backend/internal/store"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestRouter(t *testing.T, cfg *config.Config) (*sql.DB, sqlmock.Sqlmock, *gin.Engine) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	q := store.New(db)
	r := server.NewRouter(cfg, q)
	return db, mock, r
}

func defaultTestConfig() *config.Config {
	return &config.Config{
		API: config.APIConfig{Port: "8080"},
		Auth: config.AuthConfig{
			JWTSecret:            "test-jwt-secret",
			LocalAuthEnabled:     true,
			LocalAdminUsername:   "admin",
			LocalAdminPassHash:   "$2a$12$A3La56.CqRH4oiOoMjnsGuwcgPv.h5xByKaYYGK/tfi4FWtbS9V4S",
			DeviceAPIAuthEnabled: false,
			DeviceAPIToken:       "shared-token",
		},
		S3: config.S3Config{
			PublicBaseURL:   "http://localhost:9000/ota-packages",
			SecretAccessKey: "minioadmin",
			SignedURLTTLSec: 600,
		},
		Integration: config.IntegrationConfig{
			Enabled:      false,
			ServiceToken: "integration-token",
		},
	}
}

func bearerHeader() string {
	return "Bearer integration-test-token"
}

func loginBearer(t *testing.T, r *gin.Engine) string {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"username":"admin","password":"Admin@123456"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("login json: %v", err)
	}
	if resp.Data.AccessToken == "" {
		t.Fatal("login access_token empty")
	}
	return "Bearer " + resp.Data.AccessToken
}

func mockRunningTasksForDevice(mock sqlmock.Sqlmock, deviceID, taskID, packageID string) {
	now := time.Now()
	mock.ExpectQuery(`FROM t_release_task t`).
		WithArgs(deviceID).
		WillReturnRows(sqlmock.NewRows([]string{
			"task_id", "package_id", "target_group", "product_model", "hardware_version",
			"failure_threshold", "state", "created_at", "canary_percent", "schedule_time", "force_upgrade",
		}).AddRow(taskID, packageID, "org-1001", "V9", "1.0", "0.05", "Running", now, int32(100), nil, false))
}

func mockPackageDetail(mock sqlmock.Sqlmock, packageID, version string) {
	mock.ExpectQuery(`FROM t_package WHERE package_id`).
		WithArgs(packageID).
		WillReturnRows(sqlmock.NewRows([]string{
			"package_id", "product_code", "version", "file_hash", "signature", "status", "min_upgradable_version",
		}).AddRow(packageID, "AMS", version, "abc123", "sig", "ready", ""))
}

func presignTestS3Config() config.S3Config {
	return config.S3Config{
		Endpoint:        "http://minio:9000",
		PublicBaseURL:   "http://localhost:9000/ota-packages",
		Region:          "us-east-1",
		Bucket:          "ota-packages",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		SignedURLTTLSec: 600,
	}
}
