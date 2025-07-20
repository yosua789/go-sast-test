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
		log.Info().Msg("Using Cors Middleware")
		c.Header("Content-Type", "application/json")
		c.Header("Accept", "application/json")

		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Max-Age", "86400")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
		// c.Header("Access-Control-Allow-Credentials", "true")

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
