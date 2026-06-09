package server

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"ota-server/backend/internal/config"
)

func TestGenerateAndValidateOIDCState(t *testing.T) {
	cfg := &config.Config{
		Auth: config.AuthConfig{JWTSecret: "jwt-secret-fallback"},
		OIDC: config.OIDCConfig{StateSigningKey: "state-signing-key", StateTTLSec: 60},
	}

	state, err := generateOIDCState(cfg)
	if err != nil {
		t.Fatalf("generateOIDCState() error = %v", err)
	}

	if err := validateOIDCState(cfg, state); err != nil {
		t.Fatalf("validateOIDCState() error = %v", err)
	}
}

func TestValidateOIDCState_Expired(t *testing.T) {
	cfg := &config.Config{
		Auth: config.AuthConfig{JWTSecret: "jwt-secret-fallback"},
		OIDC: config.OIDCConfig{StateSigningKey: "state-signing-key", StateTTLSec: 60},
	}

	payload := oidcStatePayload{
		Nonce: "nonce-1",
		Exp:   time.Now().Add(-1 * time.Minute).Unix(),
	}
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	sig := signOIDCState(rawPayload, oidcStateSigningKey(cfg))
	state := base64.RawURLEncoding.EncodeToString(rawPayload) + "." + base64.RawURLEncoding.EncodeToString(sig)

	if err := validateOIDCState(cfg, state); err == nil {
		t.Fatalf("validateOIDCState() expected error for expired state")
	}
}

func TestRequireDeviceAuth_Disabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/device/v1/check-update", nil)

	cfg := &config.Config{Auth: config.AuthConfig{DeviceAPIAuthEnabled: false}}
	if ok := requireDeviceAuth(c, cfg, nil, []byte(`{"device_id":"AMS000001"}`)); !ok {
		t.Fatalf("requireDeviceAuth() = false, want true when auth disabled")
	}
}

func TestRequireDeviceAuth_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/device/v1/check-update", nil)

	cfg := &config.Config{Auth: config.AuthConfig{DeviceAPIAuthEnabled: true}}
	if ok := requireDeviceAuth(c, cfg, nil, []byte(`{"device_id":"AMS000001"}`)); ok {
		t.Fatalf("requireDeviceAuth() = true, want false when authorization is missing")
	}
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
