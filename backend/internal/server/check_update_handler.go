package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"ota-server/backend/internal/config"
	"ota-server/backend/internal/store"
)

type checkUpdateRequest struct {
	DeviceID        string `json:"device_id"`
	Group           string `json:"group"`
	ProductModel    string `json:"product_model"`
	HardwareVersion string `json:"hardware_version"`
	CurrentVersion  string `json:"current_version"`
}

func noUpgradeResponse(c *gin.Context, reason string, retryAfterSec int, currentVersion string) {
	data := gin.H{
		"has_update":      false,
		"reason":          reason,
		"retry_after_sec": retryAfterSec,
	}
	if strings.TrimSpace(currentVersion) != "" {
		data["current_version"] = strings.TrimSpace(currentVersion)
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    2001,
		"message": "No available upgrade",
		"data":    data,
	})
}

func handleCheckUpdate(c *gin.Context, cfg *config.Config, q *store.Queries) {
	var req checkUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1002, "message": "invalid request"})
		return
	}
	req.DeviceID = strings.TrimSpace(req.DeviceID)
	req.Group = strings.TrimSpace(req.Group)
	req.ProductModel = strings.TrimSpace(req.ProductModel)
	req.HardwareVersion = strings.TrimSpace(req.HardwareVersion)
	req.CurrentVersion = strings.TrimSpace(req.CurrentVersion)

	if req.DeviceID == "" || req.Group == "" || req.ProductModel == "" || req.HardwareVersion == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1002, "message": "device_id/group/product_model/hardware_version are required"})
		return
	}

	dev, err := q.GetDeviceRegistry(c.Request.Context(), req.DeviceID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    2003,
			"message": "not_in_catalog",
			"data":    gin.H{"has_update": false, "reason": "not_in_catalog", "retry_after_sec": 3600},
		})
		return
	}
	if dev.EligibilityState == "blocked" {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    2004,
			"message": "device_blocked",
			"data":    gin.H{"has_update": false, "reason": "device_blocked", "retry_after_sec": 3600},
		})
		return
	}
	if dev.DeviceGroup != req.Group || dev.ProductModel != req.ProductModel || dev.HardwareVersion != req.HardwareVersion {
		c.JSON(http.StatusConflict, gin.H{
			"code":    2006,
			"message": "identity_mismatch",
			"data": gin.H{
				"has_update":      false,
				"reason":          "identity_mismatch",
				"retry_after_sec": 3600,
			},
		})
		return
	}

	reported := policyReportedVersion(dev.ReportedVersion, dev.CatalogVersion, req.CurrentVersion)

	tasks, err := q.ListRunningTasksForDevice(c.Request.Context(), req.DeviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "query tasks failed"})
		return
	}
	if len(tasks) == 0 {
		noUpgradeResponse(c, "no_running_task", 3600, reported)
		return
	}

	var matched *store.TReleaseTask
	for i := range tasks {
		candidate := tasks[i]
		if !inCanaryRange(req.DeviceID, candidate.TaskID, candidate.CanaryPercent) {
			continue
		}
		matched = &candidate
		break
	}
	if matched == nil {
		noUpgradeResponse(c, "not_in_canary", 1800, reported)
		return
	}

	pkg, err := q.GetPackageDetail(c.Request.Context(), matched.PackageID)
	if err != nil {
		noUpgradeResponse(c, "no_running_task", 3600, reported)
		return
	}

	if reported != "" {
		if versionAtLeast(reported, pkg.Version) {
			noUpgradeResponse(c, "already_latest", 86400, reported)
			return
		}
		if pkg.MinUpgradableVersion != "" && CompareVersion(reported, pkg.MinUpgradableVersion) < 0 {
			noUpgradeResponse(c, "version_not_eligible", 86400, reported)
			return
		}
	}

	threshold, _ := strconv.ParseFloat(matched.FailureThreshold, 64)
	downloadURL := buildCheckUpdateDownloadURL(cfg, pkg.PackageID)
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data": gin.H{
			"has_update":        true,
			"task_id":           matched.TaskID,
			"package_id":        pkg.PackageID,
			"target_version":    pkg.Version,
			"file_hash":         pkg.FileHash,
			"signature":         pkg.Signature,
			"download_url":      downloadURL,
			"current_version":   reported,
			"upgrade_mode":      "full",
			"retry_policy":      "full-retry",
			"target_group":      matched.TargetGroup,
			"target_model":      matched.ProductModel,
			"target_hardware":   matched.HardwareVersion,
			"failure_threshold": threshold,
			"retry_after_sec":   86400,
		},
	})
}
