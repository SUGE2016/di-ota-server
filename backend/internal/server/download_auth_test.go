package server

import (
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"ota-server/backend/internal/config"
)

func TestBuildCheckUpdateDownloadURL_HMACEnabled(t *testing.T) {
	cfg := &config.Config{
		API: config.APIConfig{PublicBaseURL: "http://localhost:8080"},
		Auth: config.AuthConfig{
			DeviceDownloadHMACEnabled: true,
			DeviceSigningSecret:       "test-signing-key",
		},
		S3: config.S3Config{SignedURLTTLSec: 600},
	}
	downloadURL := buildCheckUpdateDownloadURL(cfg, "pkg-abc")
	if !strings.HasPrefix(downloadURL, "http://localhost:8080/device/v1/packages/pkg-abc/download?") {
		t.Fatalf("unexpected url: %s", downloadURL)
	}
	u, err := url.Parse(downloadURL)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	expires := u.Query().Get("expires")
	sig := u.Query().Get("signature")
	if err := verifyPackageDownloadSignature(cfg, "pkg-abc", expires, sig); err != nil {
		t.Fatalf("verify signature: %v", err)
	}
}

func TestVerifyPackageDownloadSignature_Expired(t *testing.T) {
	cfg := &config.Config{
		Auth: config.AuthConfig{
			DeviceDownloadHMACEnabled: true,
			DeviceSigningSecret:       "test-signing-key",
		},
	}
	expired := strconv.FormatInt(time.Now().Add(-time.Minute).Unix(), 10)
	sig := buildPackageDownloadSignature(cfg, "pkg-1", mustParseInt64(expired))
	if err := verifyPackageDownloadSignature(cfg, "pkg-1", expired, sig); err == nil {
		t.Fatal("expected expired error")
	}
}

func mustParseInt64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}
