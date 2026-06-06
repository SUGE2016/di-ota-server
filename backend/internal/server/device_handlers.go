package server

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"ota-server/backend/internal/store"
)

func parseDeviceCatalogFilter(c *gin.Context, limit, offset int) store.ListDeviceCatalogFilter {
	filter := store.ListDeviceCatalogFilter{
		Limit:            int32(limit),
		Offset:           int32(offset),
		Search:           strings.TrimSpace(c.Query("search")),
		DeviceGroup:      strings.TrimSpace(c.Query("group")),
		ProductModel:     strings.TrimSpace(c.Query("product_model")),
		Tag:              strings.TrimSpace(c.Query("tag")),
		EligibilityState: strings.TrimSpace(c.Query("eligibility_state")),
	}
	switch strings.ToLower(strings.TrimSpace(c.Query("abnormal"))) {
	case "true", "1", "yes":
		filter.AbnormalOnly = true
	}
	return filter
}

func parseUpgradeRecordLimit(c *gin.Context) int32 {
	limit := int32(50)
	if l := strings.TrimSpace(c.Query("limit")); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
			limit = int32(n)
		}
	}
	return limit
}
