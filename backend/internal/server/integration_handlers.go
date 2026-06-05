package server

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"ota-server/backend/internal/config"
	"ota-server/backend/internal/store"
)

func requireIntegrationAuth(c *gin.Context, cfg *config.Config) bool {
	if !cfg.Integration.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 1002, "message": "integration is disabled"})
		return false
	}
	expected := strings.TrimSpace(cfg.Integration.ServiceToken)
	if expected == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "integration token is not configured"})
		return false
	}
	header := c.GetHeader("Authorization")
	if !hasBearer(header) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1001, "message": "unauthorized"})
		return false
	}
	parts := strings.SplitN(header, " ", 2)
	provided := strings.TrimSpace(parts[1])
	if subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1001, "message": "unauthorized"})
		return false
	}
	return true
}

func registerIntegrationRoutes(api *gin.RouterGroup, cfg *config.Config, q *store.Queries) {
	integ := api.Group("/integrations")
	{
		integ.POST("/catalog/sync", func(c *gin.Context) {
			if !requireIntegrationAuth(c, cfg) {
				return
			}

			var req struct {
				BatchID     string `json:"batch_id"`
				Mode        string `json:"mode"`
				Source      string `json:"source"`
				SyncOptions struct {
					VersionPolicy    string `json:"version_policy"`
					OnIdentityChange string `json:"on_identity_change"`
					StaleBatchBefore string `json:"stale_batch_before"`
				} `json:"sync_options"`
				Devices []struct {
					DeviceID            string          `json:"device_id"`
					ProductCode         string          `json:"product_code"`
					ProductModel        string          `json:"product_model"`
					HardwareVersion     string          `json:"hardware_version"`
					CurrentVersion      string          `json:"current_version"`
					DeviceGroup         string          `json:"device_group"`
					Tags                json.RawMessage `json:"tags"`
					UpdatedAt           string          `json:"updated_at"`
					ForceIdentityUpdate bool            `json:"force_identity_update"`
					ForceCatalogVersion bool            `json:"force_catalog_version"`
				} `json:"devices"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 1002, "message": "invalid request"})
				return
			}
			batchID := strings.TrimSpace(req.BatchID)
			if batchID == "" || len(req.Devices) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"code": 1002, "message": "batch_id and devices are required"})
				return
			}
			if len(req.Devices) > 5000 {
				c.JSON(http.StatusBadRequest, gin.H{"code": 1002, "message": "devices exceed limit 5000"})
				return
			}

			if existing, err := q.GetCatalogSyncBatch(c.Request.Context(), batchID); err == nil {
				var cached gin.H
				_ = json.Unmarshal(existing.Response, &cached)
				c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": cached})
				return
			}

			opts := CatalogSyncOptions{
				VersionPolicy:    req.SyncOptions.VersionPolicy,
				OnIdentityChange: req.SyncOptions.OnIdentityChange,
			}
			if strings.TrimSpace(req.SyncOptions.StaleBatchBefore) != "" {
				t, err := time.Parse(time.RFC3339, strings.TrimSpace(req.SyncOptions.StaleBatchBefore))
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"code": 1002, "message": "stale_batch_before must be RFC3339"})
					return
				}
				opts.StaleBatchBefore = &t
			}

			source := strings.TrimSpace(req.Source)
			if source == "" {
				source = "backend"
			}

			result := CatalogSyncResult{
				WarningRows:  []CatalogSyncWarning{},
				RejectedRows: []CatalogSyncReject{},
			}

			for i, d := range req.Devices {
				group := strings.TrimSpace(d.DeviceGroup)
				if group == "" {
					group = "default"
				}
				tags := d.Tags
				if len(tags) == 0 {
					tags = json.RawMessage(`{}`)
				}
				var updatedAt *time.Time
				if strings.TrimSpace(d.UpdatedAt) != "" {
					t, err := time.Parse(time.RFC3339, strings.TrimSpace(d.UpdatedAt))
					if err != nil {
						result.Rejected++
						result.RejectedRows = append(result.RejectedRows, CatalogSyncReject{
							Index: i, DeviceID: d.DeviceID, Code: "invalid_field", Message: "updated_at must be RFC3339",
						})
						continue
					}
					updatedAt = &t
				}

				in := CatalogDeviceInput{
					DeviceID:            strings.TrimSpace(d.DeviceID),
					ProductCode:         strings.TrimSpace(d.ProductCode),
					ProductModel:        strings.TrimSpace(d.ProductModel),
					HardwareVersion:     strings.TrimSpace(d.HardwareVersion),
					CurrentVersion:      strings.TrimSpace(d.CurrentVersion),
					DeviceGroup:         group,
					Tags:                tags,
					UpdatedAt:           updatedAt,
					ForceIdentityUpdate: d.ForceIdentityUpdate,
					ForceCatalogVersion: d.ForceCatalogVersion,
				}
				if in.DeviceID == "" || in.ProductCode == "" || in.ProductModel == "" || in.HardwareVersion == "" {
					result.Rejected++
					result.RejectedRows = append(result.RejectedRows, CatalogSyncReject{
						Index: i, DeviceID: in.DeviceID, Code: "invalid_field",
						Message: "device_id/product_code/product_model/hardware_version are required",
					})
					continue
				}

				warn, reject, err := applyCatalogDevice(c.Request.Context(), q, source, opts, in)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "sync failed"})
					return
				}
				if reject != nil {
					reject.Index = i
					result.Rejected++
					result.RejectedRows = append(result.RejectedRows, *reject)
					continue
				}
				result.Accepted++
				if warn.Code != "" {
					result.Warned++
					result.WarningRows = append(result.WarningRows, warn)
				}
			}

			data := gin.H{
				"batch_id":      batchID,
				"accepted":      result.Accepted,
				"warned":        result.Warned,
				"rejected":      result.Rejected,
				"warning_rows":  result.WarningRows,
				"rejected_rows": result.RejectedRows,
			}
			respBytes, _ := json.Marshal(data)
			_ = q.InsertCatalogSyncBatch(c.Request.Context(), store.CatalogSyncBatch{
				BatchID:       batchID,
				Source:        source,
				Mode:          strings.TrimSpace(req.Mode),
				AcceptedCount: int32(result.Accepted),
				WarnedCount:   int32(result.Warned),
				RejectedCount: int32(result.Rejected),
				Response:      respBytes,
			})
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": data})
		})

		integ.GET("/pending-upgrades", func(c *gin.Context) {
			if !requireIntegrationAuth(c, cfg) {
				return
			}

			limit := int32(500)
			offset := int32(0)
			if v := c.Query("limit"); v != "" {
				if n, err := parseInt32(v); err == nil && n > 0 && n <= 2000 {
					limit = n
				}
			}
			if v := c.Query("offset"); v != "" {
				if n, err := parseInt32(v); err == nil && n >= 0 {
					offset = n
				}
			}
			group := strings.TrimSpace(c.Query("device_group"))

			rows, err := q.ListPendingUpgrades(c.Request.Context(), group, limit+1, offset)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "query pending upgrades failed"})
				return
			}

			hasMore := len(rows) > int(limit)
			if hasMore {
				rows = rows[:limit]
			}

			items := make([]gin.H, 0, len(rows))
			expires := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
			for _, row := range rows {
				if !inCanaryRange(row.DeviceID, row.TaskID, row.CanaryPercent) {
					continue
				}
				dev, err := q.GetDeviceRegistry(c.Request.Context(), row.DeviceID)
				if err != nil {
					continue
				}
				reported := effectiveReportedVersion(dev.ReportedVersion, dev.CatalogVersion)
				if reported != "" && versionAtLeast(reported, row.TargetVersion) {
					continue
				}
				if reported != "" && row.MinUpgradable != "" && CompareVersion(reported, row.MinUpgradable) < 0 {
					continue
				}
				items = append(items, gin.H{
					"device_id":         row.DeviceID,
					"task_id":           row.TaskID,
					"target_version":    row.TargetVersion,
					"package_id":        row.PackageID,
					"hint_expires_at":   expires,
				})
			}

			cursor := ""
			if hasMore {
				cursor = base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(int(offset + limit))))
			}

			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "ok",
				"data": gin.H{
					"cursor":   cursor,
					"items":    items,
					"has_more": hasMore,
				},
			})
		})
	}
}

func parseInt32(v string) (int32, error) {
	n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 32)
	return int32(n), err
}
