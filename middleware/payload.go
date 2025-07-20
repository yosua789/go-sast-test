package middleware

import (
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func (m *MiddlewareImpl) PayloadPasser() gin.HandlerFunc {
	return func(c *gin.Context) {
		rawJSON, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read request body")
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(strings.NewReader(string(rawJSON)))
		c.Set("rawPayload", string(rawJSON))
		c.Next()
	}
	// Note: This is a placeholder for the actual implementation.
}
