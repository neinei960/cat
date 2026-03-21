package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/pkg/auth"
	"github.com/neinei960/cat/server/pkg/response"
)

func WxAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.ErrorWithCode(c, http.StatusUnauthorized, 401, "未登录")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.ErrorWithCode(c, http.StatusUnauthorized, 401, "token格式错误")
			c.Abort()
			return
		}

		claims, err := auth.ParseCustomerToken(parts[1])
		if err != nil {
			response.ErrorWithCode(c, http.StatusUnauthorized, 401, "token无效或已过期")
			c.Abort()
			return
		}

		c.Set("customer_id", claims.CustomerID)
		c.Set("shop_id", claims.ShopID)
		c.Set("openid", claims.OpenID)
		c.Next()
	}
}
