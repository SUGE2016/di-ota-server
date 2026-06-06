package server

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"ota-server/backend/internal/store"
)

func mapAlert(a store.AlertRecord) gin.H {
	return gin.H{
		"alert_id":      a.AlertID,
		"alert_type":    a.AlertType,
		"severity":      a.Severity,
		"status":        a.Status,
		"resource_type": a.ResourceType,
		"resource_id":   a.ResourceID,
		"message":       a.Message,
		"detail":        jsonRawOrEmpty(a.Detail),
		"created_at":    a.CreatedAt,
		"updated_at":    a.UpdatedAt,
	}
}

func mapAlerts(items []store.AlertRecord) []gin.H {
	out := make([]gin.H, 0, len(items))
	for _, item := range items {
		out = append(out, mapAlert(item))
	}
	return out
}

func jsonRawOrEmpty(raw []byte) interface{} {
	if len(raw) == 0 {
		return gin.H{}
	}
	var v interface{}
	if err := json.Unmarshal(raw, &v); err == nil {
		return v
	}
	return gin.H{}
}
