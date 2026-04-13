package httpapi

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf("path=%s method=%s status=%d duration_ms=%d", c.FullPath(), c.Request.Method, c.Writer.Status(), time.Since(start).Milliseconds())
	}
}
