package server

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"ota-server/backend/internal/config"
	"ota-server/backend/internal/store"
)

const roleAdmin = "admin"
const roleSecretAdmin = "secret_admin"

func authUserFromRequest(c *gin.Context, cfg *config.Config, q *store.Queries) (store.UserRecord, bool) {
	operator, ok := operatorFromBearer(c.GetHeader("Authorization"), cfg)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1001, "message": "unauthorized"})
		return store.UserRecord{}, false
	}
	user, err := q.GetUserByUsername(c.Request.Context(), operator)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 1001, "message": "unauthorized"})
			return store.UserRecord{}, false
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "query user failed"})
		return store.UserRecord{}, false
	}
	if user.Status != "enabled" {
		c.JSON(http.StatusForbidden, gin.H{"code": 1003, "message": "user disabled"})
		return store.UserRecord{}, false
	}
	return user, true
}

func userHasRole(user store.UserRecord, role string) bool {
	role = strings.ToLower(strings.TrimSpace(role))
	for _, item := range user.Roles {
		if strings.EqualFold(item, role) {
			return true
		}
	}
	return false
}

func userHasAnyRole(user store.UserRecord, roles ...string) bool {
	for _, role := range roles {
		if userHasRole(user, role) {
			return true
		}
	}
	return false
}

func requireRoles(c *gin.Context, cfg *config.Config, q *store.Queries, roles ...string) (store.UserRecord, bool) {
	user, ok := authUserFromRequest(c, cfg, q)
	if !ok {
		return store.UserRecord{}, false
	}
	if userHasRole(user, roleAdmin) || userHasAnyRole(user, roles...) {
		return user, true
	}
	c.JSON(http.StatusForbidden, gin.H{"code": 1003, "message": "insufficient role"})
	return user, false
}

func requireSecretAdmin(c *gin.Context, cfg *config.Config, q *store.Queries) (store.UserRecord, bool) {
	return requireRoles(c, cfg, q, roleSecretAdmin)
}
