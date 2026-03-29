package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/response"
)

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.ErrorWithCode(c, http.StatusForbidden, 403, "无权访问")
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			response.ErrorWithCode(c, http.StatusForbidden, 403, "无权访问")
			c.Abort()
			return
		}

		for _, r := range roles {
			if roleStr == r {
				c.Next()
				return
			}
		}

		response.ErrorWithCode(c, http.StatusForbidden, 403, "权限不足")
		c.Abort()
	}
}

func RequireMinRole(minRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.ErrorWithCode(c, http.StatusForbidden, 403, "无权访问")
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok || !model.HasStaffRoleAtLeast(roleStr, minRole) {
			response.ErrorWithCode(c, http.StatusForbidden, 403, "权限不足")
			c.Abort()
			return
		}

		c.Next()
	}
}
