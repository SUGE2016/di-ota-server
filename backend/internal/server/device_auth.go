package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"ota-server/backend/internal/config"
	"ota-server/backend/internal/store"
)

type deviceAuthParams struct {
	DeviceID  string
	Timestamp int64
	Signature string
}

func readRequestBody(c *gin.Context) ([]byte, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	c.Request.Body = io.NopCloser(strings.NewReader(string(body)))
	return body, nil
}

func deviceRequestBodyHash(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func buildDeviceSignaturePayload(timestamp int64, method, path, bodyHash string) string {
	return fmt.Sprintf("%d\n%s\n%s\n%s", timestamp, strings.ToUpper(method), path, bodyHash)
}

func signDeviceRequest(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func BuildDeviceAuthorizationHeader(deviceID, secret, method, path string, body []byte, timestamp int64) string {
	payload := buildDeviceSignaturePayload(timestamp, method, path, deviceRequestBodyHash(body))
	sig := signDeviceRequest(secret, payload)
	return fmt.Sprintf("Device device_id=%s,timestamp=%d,signature=%s", deviceID, timestamp, sig)
}

func parseDeviceAuthorization(header string) (deviceAuthParams, error) {
	header = strings.TrimSpace(header)
	if !strings.HasPrefix(strings.ToLower(header), "device ") {
		return deviceAuthParams{}, fmt.Errorf("invalid authorization scheme")
	}
	raw := strings.TrimSpace(header[len("Device"):])
	if raw == "" {
		return deviceAuthParams{}, fmt.Errorf("missing authorization parameters")
	}

	out := deviceAuthParams{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		idx := strings.Index(part, "=")
		if idx <= 0 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(part[:idx]))
		value := strings.TrimSpace(part[idx+1:])
		switch key {
		case "device_id":
			out.DeviceID = value
		case "timestamp":
			ts, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return deviceAuthParams{}, fmt.Errorf("invalid timestamp")
			}
			out.Timestamp = ts
		case "signature":
			out.Signature = value
		}
	}
	if out.DeviceID == "" || out.Timestamp <= 0 || out.Signature == "" {
		return deviceAuthParams{}, fmt.Errorf("device_id/timestamp/signature are required")
	}
	return out, nil
}

func parseDeviceIDFromBody(body []byte) string {
	var payload struct {
		DeviceID string `json:"device_id"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	return strings.TrimSpace(payload.DeviceID)
}

func requireDeviceAuth(c *gin.Context, cfg *config.Config, q *store.Queries, rawBody []byte) bool {
	if !cfg.Auth.DeviceAPIAuthEnabled {
		return true
	}

	params, err := parseDeviceAuthorization(c.GetHeader("Authorization"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1001, "message": "unauthorized"})
		return false
	}

	bodyDeviceID := parseDeviceIDFromBody(rawBody)
	if bodyDeviceID != "" && bodyDeviceID != params.DeviceID {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1001, "message": "device_id mismatch"})
		return false
	}

	now := time.Now().Unix()
	tolerance := cfg.Auth.DeviceAuthTimestampToleranceSec
	if tolerance <= 0 {
		tolerance = 300
	}
	if delta := now - params.Timestamp; delta > tolerance || delta < -tolerance {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1001, "message": "timestamp out of range"})
		return false
	}

	cred, err := q.GetDeviceAuthCredential(c.Request.Context(), params.DeviceID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"code": 2003, "message": "not_in_catalog"})
		return false
	}
	if cred.EligibilityState == "blocked" {
		c.JSON(http.StatusForbidden, gin.H{"code": 2004, "message": "device_blocked"})
		return false
	}
	if strings.TrimSpace(cred.DeviceSecret) == "" {
		c.JSON(http.StatusForbidden, gin.H{"code": 2007, "message": "device_secret_not_provisioned"})
		return false
	}

	payload := buildDeviceSignaturePayload(params.Timestamp, c.Request.Method, c.Request.URL.Path, deviceRequestBodyHash(rawBody))
	expected := signDeviceRequest(cred.DeviceSecret, payload)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(params.Signature)) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1001, "message": "unauthorized"})
		return false
	}

	return true
}
