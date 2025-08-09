package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (m *MiddlewareImpl) IsAuthorized() gin.HandlerFunc {
	return func(c *gin.Context) {
		APIKey := c.GetHeader("X-API-Key")
		if APIKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "")
			return
		}
		if APIKey == m.Env.Admin.ApiKey {
			c.Next()
			return
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "unauthorized")
			return
		}

	}
}
