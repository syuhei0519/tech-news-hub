package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"tech-feed-hub/collector-service/internal/collector"
)

func RegisterRoutes(router *gin.Engine, service *collector.Service) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.POST("/api/v1/collect/run", func(c *gin.Context) {
		result, err := service.Run(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"results": result})
	})
}
