package integration_test

import (
	"database/sql"
	"testing"

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
