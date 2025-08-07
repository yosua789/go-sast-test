package middleware

import (
	"assist-tix/config"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Middleware interface {
	CORSMiddleware() gin.HandlerFunc
	PayloadPasser() gin.HandlerFunc
	TokenAuthMiddleware() gin.HandlerFunc
}

type MiddlewareImpl struct {
	Env *config.EnvironmentVariable
}

func NewMiddleware(env *config.EnvironmentVariable) Middleware {
	return &MiddlewareImpl{
		Env: env,
	}
}

func (m *MiddlewareImpl) CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Header("Accept", "application/json")

		c.Writer.Header().Set("Access-Control-Allow-Origin", m.Env.Api.Url)
		c.Header("Access-Control-Max-Age", "86400")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; object-src 'none'; frame-ancestors 'none';")

		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains;")

		c.Header("X-Content-Type-Options", "nosniff")

		c.Header("X-Frame-Options", "DENY")

		c.Header("X-XSS-Protection", "1; mode=block")

		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		if c.Request.Method == "OPTIONS" {
			log.Info().Msg("Abort Options")
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Access-Control-Allow-Headers, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
