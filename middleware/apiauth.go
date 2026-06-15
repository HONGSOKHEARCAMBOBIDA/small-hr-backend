package middleware

import (
	"log"
	"mysql/constant/share"
	"os"

	"github.com/gin-gonic/gin"
)

func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ទាញ API Key ពី Header
		key := c.GetHeader("x-api-key")

		// ពិនិត្យ Key
		if key == "" {
			log.Printf("key missing")
			share.ResponseError(c, 999, "API Key missing")
			c.Abort()
			return
		}

		if key != os.Getenv("API_KEY_SECRET") {
			log.Printf("key invali")
			share.ResponseError(c, 999, "Invalid API Key")
			c.Abort()
			return
		}

		c.Next()
	}
}
