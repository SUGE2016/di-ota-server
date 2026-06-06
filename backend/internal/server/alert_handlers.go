package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"ota-server/backend/internal/config"
	"ota-server/backend/internal/store"
)

func registerAlertRoutes(api *gin.RouterGroup, q *store.Queries) {
	api.GET("/alerts", func(c *gin.Context) {
		if !hasBearer(c.GetHeader("Authorization")) {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 1001, "message": "unauthorized"})
			return
		}
		limit, offset := parseLimitOffset(c, 20, 200)
		status := strings.TrimSpace(c.Query("status"))
		severity := strings.TrimSpace(c.Query("severity"))
		items, err := q.ListAlerts(c.Request.Context(), store.ListAlertsParams{
			Status: status, Severity: severity, Limit: int32(limit), Offset: int32(offset),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "query alerts failed"})
			return
		}
		total, _ := q.CountAlerts(c.Request.Context(), status, severity)
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"alerts": mapAlerts(items), "total": total}})
	})

	api.POST("/alerts/actions", func(c *gin.Context) {
		if !hasBearer(c.GetHeader("Authorization")) {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 1001, "message": "unauthorized"})
			return
		}
		var req struct {
			Action   string   `json:"action"`
			AlertIDs []string `json:"alert_ids"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1002, "message": "invalid request"})
			return
		}
		action := strings.ToLower(strings.TrimSpace(req.Action))
		next := ""
		switch action {
		case "acknowledge", "ack":
			next = "acknowledged"
		case "close", "closed":
			next = "closed"
		default:
			c.JSON(http.StatusBadRequest, gin.H{"code": 1002, "message": "action must be acknowledge or close"})
			return
		}
		n, err := q.UpdateAlertsStatusByIDs(c.Request.Context(), req.AlertIDs, next)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "update alerts failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"updated": n}})
	})
}

func parseLimitOffset(c *gin.Context, defaultLimit, maxLimit int) (int, int) {
	limit := defaultLimit
	offset := 0
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= maxLimit {
			limit = n
		}
	}
	if o := c.Query("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}
	return limit, offset
}

func UpsertTaskFailureAlert(ctx context.Context, q *store.Queries, taskID string, failureRate float64, threshold float64) error {
	detail, _ := json.Marshal(gin.H{
		"failure_rate":      failureRate,
		"failure_threshold": threshold,
		"source":            "worker_auto_pause",
	})
	_, err := q.CreateAlert(ctx, store.AlertRecord{
		AlertID:      "alert-" + uuid.NewString(),
		AlertType:    "任务失败率过高",
		Severity:     "critical",
		Status:       "open",
		ResourceType: "task",
		ResourceID:   taskID,
		Message:      fmt.Sprintf("任务 %s 失败率 %.2f%% 超过阈值 %.2f%%，已自动暂停", taskID, failureRate*100, threshold*100),
		Detail:       detail,
		DedupeKey:    "task_failure_rate:" + taskID,
	})
	return err
}

func UpsertUpgradeStatusAlert(ctx context.Context, q *store.Queries, deviceID, taskID, status, targetVersion string) error {
	severity := "info"
	if status == "Failed" {
		severity = "warning"
	}
	detail, _ := json.Marshal(gin.H{"device_id": deviceID, "task_id": taskID, "status": status, "target_version": targetVersion})
	_, err := q.CreateAlert(ctx, store.AlertRecord{
		AlertID:      "alert-" + uuid.NewString(),
		AlertType:    "设备升级状态",
		Severity:     severity,
		Status:       "open",
		ResourceType: "device",
		ResourceID:   deviceID,
		Message:      fmt.Sprintf("设备 %s 任务 %s 上报 %s", deviceID, taskID, status),
		Detail:       detail,
		DedupeKey:    fmt.Sprintf("upgrade_status:%s:%s:%s", deviceID, taskID, status),
	})
	return err
}

func logAlertError(err error) {
	if err != nil {
		log.Printf("alert upsert failed: %v", err)
	}
}

func EmitUpgradeWebhookAsync(cfg *config.Config, deviceID, taskID, status, sourceVersion, targetVersion string) {
	if !cfg.Integration.WebhookEnabled || strings.TrimSpace(cfg.Integration.WebhookURL) == "" {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := sendOTAWebhook(ctx, cfg, webhookPayload{
			Event:  "upgrade.status_changed",
			Device: deviceID, TaskID: taskID, Status: status,
			SourceVersion: sourceVersion, TargetVersion: targetVersion,
		}); err != nil {
			log.Printf("webhook send failed: %v", err)
		}
	}()
}

func EmitTaskPausedWebhookAsync(cfg *config.Config, taskID, reason string, failureRate, threshold float64) {
	if !cfg.Integration.WebhookEnabled || strings.TrimSpace(cfg.Integration.WebhookURL) == "" {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := sendOTAWebhook(ctx, cfg, webhookPayload{
			Event: "task.auto_paused", TaskID: taskID, Status: "Paused",
			Detail: gin.H{"reason": reason, "failure_rate": failureRate, "failure_threshold": threshold},
		}); err != nil {
			log.Printf("webhook send failed: %v", err)
		}
	}()
}
