//go:build e2e

package integration_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestE2E01_PackageUploadFlow(t *testing.T) {
	base := requireE2EStack(t)
	bearer := e2eBearer(t, base)
	packageID := e2eUploadPackage(t, base, bearer, "AMS", "e2e-1.0.0")
	if packageID == "" {
		t.Fatal("package_id empty")
	}

	status, out := e2eJSON(t, http.MethodGet, base+"/api/v1/packages/"+packageID, bearer, nil)
	if status != http.StatusOK {
		t.Fatalf("get package status=%d out=%v", status, out)
	}
	data := out["data"].(map[string]any)
	if data["status"] != "Published" {
		t.Fatalf("package status=%v want Published", data["status"])
	}
}

func TestE2E02_CSVImportAndCreateTask(t *testing.T) {
	base := requireE2EStack(t)
	bearer := e2eBearer(t, base)
	e2eImportDevicesCSV(t, base, bearer)
	packageID := e2eUploadPackage(t, base, bearer, "AMS", fmt.Sprintf("e2e-2-%d", time.Now().Unix()))

	status, out := e2eJSON(t, http.MethodPost, base+"/api/v1/release-tasks", bearer, map[string]any{
		"package_id":        packageID,
		"group":             "org-1001",
		"product_model":     "V9",
		"hardware_version":  "1.0",
		"failure_threshold": 0.05,
		"canary_percent":    100,
		"start_now":         true,
	})
	if status != http.StatusOK {
		t.Fatalf("create task status=%d out=%v", status, out)
	}
	task := out["data"].(map[string]any)
	if task["state"] != "Running" {
		t.Fatalf("task state=%v want Running", task["state"])
	}
}

func TestE2E03_DeviceUpgradeFlow(t *testing.T) {
	base := requireE2EStack(t)
	bearer := e2eBearer(t, base)
	e2eImportDevicesCSV(t, base, bearer)
	version := fmt.Sprintf("e2e-3-%d", time.Now().Unix())
	packageID := e2eUploadPackage(t, base, bearer, "AMS", version)

	status, taskOut := e2eJSON(t, http.MethodPost, base+"/api/v1/release-tasks", bearer, map[string]any{
		"package_id":        packageID,
		"group":             "org-1001",
		"product_model":     "V9",
		"hardware_version":  "1.0",
		"failure_threshold": 0.05,
		"canary_percent":    100,
		"start_now":         true,
	})
	if status != http.StatusOK {
		t.Fatalf("create task status=%d out=%v", status, taskOut)
	}
	taskID := taskOut["data"].(map[string]any)["task_id"].(string)

	status, checkOut := e2eJSON(t, http.MethodPost, base+"/device/v1/check-update", "", map[string]any{
		"device_id":       "AMS000001",
		"current_version": "v2.3",
	})
	if status != http.StatusOK {
		t.Fatalf("check-update status=%d out=%v", status, checkOut)
	}
	if int(checkOut["code"].(float64)) != 0 {
		t.Fatalf("check-update code=%v", checkOut["code"])
	}
	checkData := checkOut["data"].(map[string]any)
	if checkData["has_update"] != true {
		t.Fatalf("has_update=%v", checkData["has_update"])
	}
	if checkData["task_id"] != taskID {
		t.Fatalf("task_id=%v want %s", checkData["task_id"], taskID)
	}

	for _, st := range []string{"Pending", "Downloading", "Downloaded", "Verifying", "Upgrading", "Success"} {
		status, repOut := e2eJSON(t, http.MethodPost, base+"/device/v1/report-status", "", map[string]any{
			"device_id":      "AMS000001",
			"task_id":        taskID,
			"status":         st,
			"source_version": "v2.3",
			"target_version": version,
		})
		if status != http.StatusOK {
			t.Fatalf("report-status %s status=%d out=%v", st, status, repOut)
		}
	}

	status, detailOut := e2eJSON(t, http.MethodGet, base+"/api/v1/release-tasks/"+taskID, bearer, nil)
	if status != http.StatusOK {
		t.Fatalf("task detail status=%d out=%v", status, detailOut)
	}
	taskData := detailOut["data"].(map[string]any)
	taskObj, _ := taskData["task"].(map[string]any)
	if taskObj == nil {
		t.Fatalf("task detail missing task: %v", detailOut)
	}

	status, histOut := e2eJSON(t, http.MethodGet, base+"/api/v1/devices/AMS000001/upgrade-records?limit=20", bearer, nil)
	if status != http.StatusOK {
		t.Fatalf("upgrade-records status=%d out=%v", status, histOut)
	}
	records, _ := histOut["data"].(map[string]any)["records"].([]any)
	foundSuccess := false
	for _, item := range records {
		rec, _ := item.(map[string]any)
		if rec["task_id"] == taskID && rec["status"] == "Success" {
			foundSuccess = true
			break
		}
	}
	if !foundSuccess {
		t.Fatalf("expected Success record for task %s, records=%v", taskID, records)
	}
}

func TestE2E04_TaskPauseStopsCheckUpdate(t *testing.T) {
	base := requireE2EStack(t)
	bearer := e2eBearer(t, base)
	e2eImportDevicesCSV(t, base, bearer)
	version := fmt.Sprintf("e2e-4-%d", time.Now().Unix())
	packageID := e2eUploadPackage(t, base, bearer, "AMS", version)

	status, taskOut := e2eJSON(t, http.MethodPost, base+"/api/v1/release-tasks", bearer, map[string]any{
		"package_id":        packageID,
		"group":             "org-1001",
		"product_model":     "V9",
		"hardware_version":  "1.0",
		"failure_threshold": 0.05,
		"canary_percent":    100,
		"start_now":         true,
	})
	if status != http.StatusOK {
		t.Fatalf("create task status=%d out=%v", status, taskOut)
	}
	taskID := taskOut["data"].(map[string]any)["task_id"].(string)

	status, pauseOut := e2eJSON(t, http.MethodPost, base+"/api/v1/release-tasks/"+taskID+"/actions", bearer, map[string]any{
		"action": "pause",
	})
	if status != http.StatusOK {
		t.Fatalf("pause task status=%d out=%v", status, pauseOut)
	}

	status, checkOut := e2eJSON(t, http.MethodPost, base+"/device/v1/check-update", "", map[string]any{
		"device_id":       "AMS000001",
		"current_version": "v2.3",
	})
	if status != http.StatusOK {
		t.Fatalf("check-update status=%d out=%v", status, checkOut)
	}
	code := int(checkOut["code"].(float64))
	if code != 2001 && code != 2002 {
		t.Fatalf("expected no update after pause, code=%d out=%v", code, checkOut)
	}
}

func TestE2E05_AlertsListAfterUpgradeFailure(t *testing.T) {
	base := requireE2EStack(t)
	bearer := e2eBearer(t, base)
	e2eImportDevicesCSV(t, base, bearer)
	version := fmt.Sprintf("e2e-5-%d", time.Now().Unix())
	packageID := e2eUploadPackage(t, base, bearer, "AMS", version)

	status, taskOut := e2eJSON(t, http.MethodPost, base+"/api/v1/release-tasks", bearer, map[string]any{
		"package_id":        packageID,
		"group":             "org-1001",
		"product_model":     "V9",
		"hardware_version":  "1.0",
		"failure_threshold": 0.05,
		"canary_percent":    100,
		"start_now":         true,
	})
	if status != http.StatusOK {
		t.Fatalf("create task status=%d out=%v", status, taskOut)
	}
	taskID := taskOut["data"].(map[string]any)["task_id"].(string)

	status, _ = e2eJSON(t, http.MethodPost, base+"/device/v1/report-status", "", map[string]any{
		"device_id":      "AMS000001",
		"task_id":        taskID,
		"status":         "Failed",
		"source_version": "v2.3",
		"target_version": version,
		"error_code":     "E2E_FAIL",
		"error_message":  "e2e simulated failure",
	})
	if status != http.StatusOK {
		t.Fatalf("report-status Failed status=%d", status)
	}

	status, alertOut := e2eJSON(t, http.MethodGet, base+"/api/v1/alerts?limit=20", bearer, nil)
	if status != http.StatusOK {
		t.Fatalf("alerts status=%d out=%v", status, alertOut)
	}
	data := alertOut["data"].(map[string]any)
	alerts, _ := data["alerts"].([]any)
	if len(alerts) == 0 {
		t.Fatal("expected at least one alert after upgrade failure")
	}
}
