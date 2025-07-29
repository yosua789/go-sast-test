package middleware

import (
	"assist-tix/helper"
	"assist-tix/lib"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func (m *MiddlewareImpl) TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil {
			log.Error().Err(err).Msg("Failed to get access token from cookie")
		}

		err = helper.AccessTokenValid(c.Request, m.Env, token)
		if err != nil {
			lib.RespondError(c, http.StatusUnauthorized, "Unauthorized", err, 40101, false)
			c.Abort()
			return
		}
		transactionID, err := helper.GetDataFromAccessToken(c.Request, m.Env)
		if err != nil {
			lib.RespondError(c, http.StatusUnauthorized, "Unauthorized", err, 40102, false)
			c.Abort()
			return
		}
		c.Set("transaction_id", transactionID)

		c.Next()

	}
}
