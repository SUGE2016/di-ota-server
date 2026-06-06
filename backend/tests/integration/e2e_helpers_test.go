//go:build e2e

package integration_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func e2eBaseURL() string {
	if v := strings.TrimSpace(os.Getenv("OTA_E2E_BASE_URL")); v != "" {
		return strings.TrimSuffix(v, "/")
	}
	return "http://localhost:8080"
}

func e2eStackAvailable(base string) bool {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(base + "/healthz")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func requireE2EStack(t *testing.T) string {
	t.Helper()
	base := e2eBaseURL()
	if !e2eStackAvailable(base) {
		t.Skip("docker compose stack not available at " + base)
	}
	return base
}

func e2eHTTPClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
}

func e2eBearer(t *testing.T, base string) string {
	t.Helper()
	if token := strings.TrimSpace(os.Getenv("OTA_E2E_BEARER_TOKEN")); token != "" {
		return "Bearer " + token
	}

	client := e2eHTTPClient()

	if os.Getenv("OTA_E2E_LOCAL_AUTH") == "true" {
		body := `{"username":"admin","password":"Admin@123456"}`
		req, _ := http.NewRequest(http.MethodPost, base+"/api/v1/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("local login: %v", err)
		}
		defer resp.Body.Close()
		var out struct {
			Data struct {
				AccessToken string `json:"access_token"`
			} `json:"data"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&out)
		if out.Data.AccessToken != "" {
			return "Bearer " + out.Data.AccessToken
		}
	}

	loginReq, _ := http.NewRequest(http.MethodGet, base+"/api/v1/auth/sso/login", nil)
	loginResp, err := client.Do(loginReq)
	if err != nil {
		t.Fatalf("sso login: %v", err)
	}
	defer loginResp.Body.Close()
	var loginOut struct {
		Data struct {
			RedirectURL string `json:"redirect_url"`
		} `json:"data"`
	}
	if err := json.NewDecoder(loginResp.Body).Decode(&loginOut); err != nil {
		t.Fatalf("sso login decode: %v", err)
	}
	redirectURL, err := url.Parse(loginOut.Data.RedirectURL)
	if err != nil || redirectURL == nil {
		t.Fatalf("invalid redirect_url: %q", loginOut.Data.RedirectURL)
	}
	state := redirectURL.Query().Get("state")
	cbURL := fmt.Sprintf("%s/api/v1/auth/sso/callback?code=mock-code&state=%s", base, url.QueryEscape(state))
	cbResp, err := client.Get(cbURL)
	if err != nil {
		t.Fatalf("sso callback: %v", err)
	}
	defer cbResp.Body.Close()
	var cbOut struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(cbResp.Body).Decode(&cbOut); err != nil {
		t.Fatalf("sso callback decode: %v", err)
	}
	if cbOut.Data.AccessToken == "" {
		t.Fatal("sso callback access_token empty")
	}
	return "Bearer " + cbOut.Data.AccessToken
}

func e2eJSON(t *testing.T, method, url, bearer string, payload any) (int, map[string]any) {
	t.Helper()
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if bearer != "" {
		req.Header.Set("Authorization", bearer)
	}
	resp, err := e2eHTTPClient().Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	var out map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

func e2eFixturePath(name string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join("..", "..", "..", "tests", "fixtures", name)
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "tests", "fixtures", name)
}

func e2eImportDevicesCSV(t *testing.T, base, bearer string) {
	t.Helper()
	path := e2eFixturePath("devices_valid.csv")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", "devices_valid.csv")
	if err != nil {
		t.Fatalf("form file: %v", err)
	}
	if _, err := io.Copy(part, f); err != nil {
		t.Fatalf("copy csv: %v", err)
	}
	_ = w.Close()

	req, err := http.NewRequest(http.MethodPost, base+"/api/v1/devices/import-csv", &buf)
	if err != nil {
		t.Fatalf("import request: %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", bearer)
	resp, err := e2eHTTPClient().Do(req)
	if err != nil {
		t.Fatalf("import csv: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("import csv status=%d body=%s", resp.StatusCode, string(b))
	}
}

func e2eUploadPackage(t *testing.T, base, bearer, productCode, version string) string {
	t.Helper()
	firmware := []byte(fmt.Sprintf("ota-e2e-firmware-%s-%d", version, time.Now().UnixNano()))
	sum := sha256.Sum256(firmware)
	fileHash := hex.EncodeToString(sum[:])
	fileSize := int64(len(firmware))

	status, uploadOut := e2eJSON(t, http.MethodPost, base+"/api/v1/packages/upload-url", bearer, map[string]any{
		"file_name":    "firmware.bin",
		"content_type": "application/octet-stream",
		"file_hash":    fileHash,
	})
	if status != http.StatusOK {
		t.Fatalf("upload-url status=%d out=%v", status, uploadOut)
	}
	data := uploadOut["data"].(map[string]any)
	packageID := data["package_id"].(string)
	uploadURL := data["upload_url"].(string)
	_ = data["required_headers"]

	putReq, err := http.NewRequest(http.MethodPut, uploadURL, bytes.NewReader(firmware))
	if err != nil {
		t.Fatalf("put request: %v", err)
	}
	// Presigned PUT 仅签名 host；附加 x-amz-meta-* 会导致 MinIO 400
	putResp, err := http.DefaultClient.Do(putReq)
	if err != nil {
		t.Fatalf("put firmware: %v", err)
	}
	putResp.Body.Close()
	if putResp.StatusCode < 200 || putResp.StatusCode >= 300 {
		t.Fatalf("put firmware status=%d", putResp.StatusCode)
	}

	status, completeOut := e2eJSON(t, http.MethodPost, base+"/api/v1/packages/complete", bearer, map[string]any{
		"package_id":   packageID,
		"product_code": productCode,
		"version":      version,
		"file_hash":    fileHash,
		"signature":    "e2e-test-signature",
		"file_size":    fileSize,
	})
	if status != http.StatusOK {
		t.Fatalf("complete status=%d out=%v", status, completeOut)
	}
	return packageID
}
