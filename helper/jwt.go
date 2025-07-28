package helper

import "github.com/gin-gonic/gin"

func SetAccessToken(c *gin.Context, token string) {

	// Set the cookie
	c.SetCookie(
		"access_token",   // name
		token,            // value
		3600,             // maxAge in seconds (e.g., 1 hour)
		"/",              // path
		"yourdomain.com", // domain, or "" for current domain
		true,             // secure (true = HTTPS only)
		true,             // httpOnly (true = JS can't access)
	)

	c.JSON(200, gin.H{"message": "Token set in cookie"})
}
