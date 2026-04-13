package httpapi

import (
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, articleServiceURL string) {
	client := &http.Client{}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := router.Group("/api/v1")
	{
		v1.GET("/articles", func(c *gin.Context) {
			targetURL := articleServiceURL + "/api/v1/articles"
			proxyGet(c, client, targetURL)
		})
		v1.GET("/articles/:id", func(c *gin.Context) {
			targetURL := articleServiceURL + "/api/v1/articles/" + c.Param("id")
			proxyGet(c, client, targetURL)
		})
	}
}

func proxyGet(c *gin.Context, client *http.Client, targetURL string) {
	parsed, err := url.Parse(targetURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	parsed.RawQuery = c.Request.URL.RawQuery

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, parsed.String(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}
