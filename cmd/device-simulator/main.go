package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type apiResp struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func main() {
	baseURL := flag.String("url", "http://localhost:8080", "OTA API base URL")
	deviceID := flag.String("device-id", "AMS000001", "device_id (sn)")
	group := flag.String("group", "org-1001", "device_group")
	model := flag.String("model", "V9", "product_model")
	hw := flag.String("hw", "1.0", "hardware_version")
	version := flag.String("version", "v2.3.0", "current_version")
	token := flag.String("token", "", "Bearer token (DEVICE_API_TOKEN)")
	skipDownload := flag.Bool("skip-download", true, "skip HTTP download step")
	dryRun := flag.Bool("dry-run", false, "only check-update")
	flag.Parse()

	client := &http.Client{Timeout: 30 * time.Second}
	authHeader := ""
	if strings.TrimSpace(*token) != "" {
		authHeader = "Bearer " + strings.TrimSpace(*token)
	}

	fmt.Println("==> check-update")
	checkData, err := checkUpdate(client, *baseURL, authHeader, *deviceID, *group, *model, *hw, *version)
	if err != nil {
		fatalf("check-update: %v", err)
	}
	printJSON("check-update response", checkData)

	var parsed map[string]interface{}
	_ = json.Unmarshal(checkData, &parsed)
	hasUpdate, _ := parsed["has_update"].(bool)
	if !hasUpdate {
		fmt.Println("no upgrade available, done")
		return
	}
	if *dryRun {
		fmt.Println("dry-run, stop")
		return
	}

	taskID, _ := parsed["task_id"].(string)
	targetVersion, _ := parsed["target_version"].(string)
	downloadURL, _ := parsed["download_url"].(string)
	if taskID == "" {
		fatalf("missing task_id in response")
	}

	steps := []struct {
		status string
		delay  time.Duration
	}{
		{"pending", 200 * time.Millisecond},
		{"downloading", 300 * time.Millisecond},
		{"downloaded", 300 * time.Millisecond},
		{"upgrading", 500 * time.Millisecond},
		{"success", 0},
	}

	for _, step := range steps {
		time.Sleep(step.delay)
		fmt.Printf("==> report-status %s\n", step.status)
		body := map[string]string{
			"device_id":      *deviceID,
			"task_id":        taskID,
			"status":         step.status,
			"source_version": *version,
			"target_version": targetVersion,
		}
		resp, err := postJSON(client, *baseURL+"/device/v1/report-status", authHeader, body)
		if err != nil {
			fatalf("report-status %s: %v", step.status, err)
		}
		printJSON("report-status response", resp)
	}

	if !*skipDownload && downloadURL != "" {
		fmt.Println("==> download", downloadURL)
		req, _ := http.NewRequest(http.MethodGet, downloadURL, nil)
		res, err := client.Do(req)
		if err != nil {
			fatalf("download: %v", err)
		}
		defer res.Body.Close()
		n, _ := io.Copy(io.Discard, res.Body)
		fmt.Printf("downloaded %d bytes, status=%d\n", n, res.StatusCode)
	}

	fmt.Println("upgrade simulation completed")
}

func checkUpdate(client *http.Client, baseURL, authHeader, deviceID, group, model, hw, version string) (json.RawMessage, error) {
	body := map[string]string{
		"device_id":         deviceID,
		"group":             group,
		"product_model":     model,
		"hardware_version":  hw,
		"current_version":   version,
	}
	return postJSON(client, baseURL+"/device/v1/check-update", authHeader, body)
}

func postJSON(client *http.Client, url, authHeader string, body any) (json.RawMessage, error) {
	payload, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", res.StatusCode, string(raw))
	}
	var wrapped apiResp
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, err
	}
	if wrapped.Code != 0 && wrapped.Code != 2001 {
		return nil, fmt.Errorf("code=%d message=%s", wrapped.Code, wrapped.Message)
	}
	return wrapped.Data, nil
}

func printJSON(title string, data json.RawMessage) {
	var pretty bytes.Buffer
	_ = json.Indent(&pretty, data, "", "  ")
	fmt.Printf("%s:\n%s\n\n", title, pretty.String())
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
