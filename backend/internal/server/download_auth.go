package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"ota-server/backend/internal/config"
)

func deviceDownloadSigningKey(cfg *config.Config) string {
	if k := strings.TrimSpace(cfg.Auth.DeviceSigningSecret); k != "" {
		return k
	}
	return strings.TrimSpace(cfg.S3.SecretAccessKey)
}

func buildPackageDownloadSignature(cfg *config.Config, packageID string, expiresAt int64) string {
	payload := fmt.Sprintf("%s:%d", packageID, expiresAt)
	mac := hmac.New(sha256.New, []byte(deviceDownloadSigningKey(cfg)))
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func verifyPackageDownloadSignature(cfg *config.Config, packageID, expiresRaw, signature string) error {
	if !cfg.Auth.DeviceDownloadHMACEnabled {
		return nil
	}
	expiresAt, err := strconv.ParseInt(strings.TrimSpace(expiresRaw), 10, 64)
	if err != nil || expiresAt <= time.Now().Unix() {
		return fmt.Errorf("download link expired")
	}
	expected := buildPackageDownloadSignature(cfg, packageID, expiresAt)
	if !constantTimeEqual(signature, expected) {
		return fmt.Errorf("invalid download signature")
	}
	return nil
}

func constantTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}
	return v == 0
}

func buildDeviceProxyDownloadURL(cfg *config.Config, packageID string) string {
	expiresAt := time.Now().Add(time.Duration(cfg.S3.SignedURLTTLSec) * time.Second).Unix()
	sig := buildPackageDownloadSignature(cfg, packageID, expiresAt)
	base := strings.TrimSuffix(strings.TrimSpace(cfg.API.PublicBaseURL), "/")
	if base == "" {
		base = "http://localhost:8080"
	}
	return fmt.Sprintf("%s/device/v1/packages/%s/download?expires=%d&signature=%s",
		base, url.PathEscape(packageID), expiresAt, url.QueryEscape(sig))
}

func buildCheckUpdateDownloadURL(cfg *config.Config, packageID string) string {
	if cfg.Auth.DeviceDownloadHMACEnabled {
		return buildDeviceProxyDownloadURL(cfg, packageID)
	}
	return buildSignedDownloadURL(cfg, packageID)
}
