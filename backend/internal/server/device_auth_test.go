package server

import (
	"strings"
	"testing"
)

func TestBuildDeviceAuthorizationHeader(t *testing.T) {
	body := []byte(`{"device_id":"AMS000001","current_version":"v2.3.0"}`)
	ts := int64(1_700_000_000)
	header := BuildDeviceAuthorizationHeader("AMS000001", "secret-abc", "POST", "/device/v1/check-update", body, ts)
	params, err := parseDeviceAuthorization(header)
	if err != nil {
		t.Fatalf("parseDeviceAuthorization() error = %v", err)
	}
	if params.DeviceID != "AMS000001" || params.Timestamp != ts || params.Signature == "" {
		t.Fatalf("unexpected params: %+v", params)
	}

	payload := buildDeviceSignaturePayload(ts, "POST", "/device/v1/check-update", deviceRequestBodyHash(body))
	if signDeviceRequest("secret-abc", payload) != params.Signature {
		t.Fatal("signature mismatch")
	}
}

func TestParseDeviceAuthorization_Invalid(t *testing.T) {
	if _, err := parseDeviceAuthorization("Bearer token"); err == nil {
		t.Fatal("expected error for bearer scheme")
	}
	if _, err := parseDeviceAuthorization("Device device_id=abc"); err == nil {
		t.Fatal("expected error for missing timestamp/signature")
	}
}

func TestDeviceRequestBodyHash_EmptyBody(t *testing.T) {
	got := deviceRequestBodyHash(nil)
	want := deviceRequestBodyHash([]byte{})
	if got != want {
		t.Fatalf("hash mismatch: %q vs %q", got, want)
	}
	if len(got) != 64 {
		t.Fatalf("expected sha256 hex, got %q", got)
	}
}

func TestBuildDeviceAuthorizationHeader_Deterministic(t *testing.T) {
	body := []byte(`{"device_id":"AMS000001"}`)
	a := BuildDeviceAuthorizationHeader("AMS000001", "secret", "POST", "/device/v1/report-status", body, 42)
	b := BuildDeviceAuthorizationHeader("AMS000001", "secret", "POST", "/device/v1/report-status", body, 42)
	if a != b {
		t.Fatalf("expected deterministic header, got %q vs %q", a, b)
	}
	if !strings.HasPrefix(a, "Device device_id=AMS000001,") {
		t.Fatalf("unexpected header prefix: %q", a)
	}
}
