package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"ota-server/backend/internal/config"
	"ota-server/backend/internal/store"
)

func registerProductModelPolicyRoutes(api *gin.RouterGroup, cfg *config.Config, q *store.Queries) {
	api.GET("/product-model-policies", func(c *gin.Context) {
		if _, ok := authUserFromRequest(c, cfg, q); !ok {
			return
		}
		policies, err := q.ListProductModelPolicies(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "query policies failed"})
			return
		}
		models, err := q.ListProductModelsForPolicyUI(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "query product models failed"})
			return
		}
		byModel := make(map[string]store.ProductModelPolicy, len(policies))
		for _, p := range policies {
			byModel[p.ProductModel] = p
		}
		items := make([]gin.H, 0, len(models))
		for _, model := range models {
			if p, ok := byModel[model]; ok {
				items = append(items, mapProductModelPolicy(p))
				continue
			}
			items = append(items, gin.H{
				"product_model":      model,
				"report_status_mode": store.ReportStatusModeRelaxed,
				"updated_at":         nil,
				"updated_by":         "",
				"explicit":           false,
			})
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data": gin.H{
				"default_mode": store.ReportStatusModeRelaxed,
				"policies":     items,
			},
		})
	})

	api.PUT("/product-model-policies/:product_model", func(c *gin.Context) {
		user, ok := requireRoles(c, cfg, q) // admin only (requireRoles with no extra roles)
		if !ok {
			return
		}
		productModel := strings.TrimSpace(c.Param("product_model"))
		if productModel == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1002, "message": "product_model is required"})
			return
		}
		var req struct {
			ReportStatusMode string `json:"report_status_mode"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1002, "message": "invalid request"})
			return
		}
		mode := store.NormalizeReportStatusMode(req.ReportStatusMode)
		// relaxed + delete explicit row keeps default; still upsert so UI shows updater
		p, err := q.UpsertProductModelPolicy(c.Request.Context(), productModel, mode, user.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "save policy failed"})
			return
		}
		item := mapProductModelPolicy(p)
		item["explicit"] = true
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": item})
	})
}

func mapProductModelPolicy(p store.ProductModelPolicy) gin.H {
	var updatedAt any
	if !p.UpdatedAt.IsZero() {
		updatedAt = p.UpdatedAt.UTC().Format(time.RFC3339Nano)
	}
	return gin.H{
		"product_model":      p.ProductModel,
		"report_status_mode": p.ReportStatusMode,
		"updated_at":         updatedAt,
		"updated_by":         p.UpdatedBy,
		"explicit":           true,
	}
}
