package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func (m *MiddlewareImpl) OriginMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.Env.Api.OriginMiddleware {
			origin := c.Request.Header.Get("Origin")
			referer := c.Request.Header.Get("Referer")
			log.Info().Msgf("Origin: %s, Referer: %s", origin, referer)
			// Check if the Origin or Referer header matches the allowed frontend
			if (origin != "" && !strings.HasPrefix(origin, m.Env.Api.Url)) ||
				(referer != "" && !strings.HasPrefix(referer, m.Env.Api.Url)) {
				log.Error().Msg("Invalid origin or referer")
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: invalid origin"})
				return
			}
		}
		c.Next()
	}
}
