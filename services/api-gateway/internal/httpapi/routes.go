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
			proxyRequest(c, client, targetURL)
		})
		v1.GET("/articles/:id", func(c *gin.Context) {
			targetURL := articleServiceURL + "/api/v1/articles/" + c.Param("id")
			proxyRequest(c, client, targetURL)
		})
		v1.GET("/sources", func(c *gin.Context) {
			targetURL := articleServiceURL + "/api/v1/sources"
			proxyRequest(c, client, targetURL)
		})
		v1.GET("/sources/:id", func(c *gin.Context) {
			targetURL := articleServiceURL + "/api/v1/sources/" + c.Param("id")
			proxyRequest(c, client, targetURL)
		})
		v1.GET("/fetch-jobs", func(c *gin.Context) {
			targetURL := articleServiceURL + "/api/v1/fetch-jobs"
			proxyRequest(c, client, targetURL)
		})
		v1.POST("/sources", func(c *gin.Context) {
			targetURL := articleServiceURL + "/api/v1/sources"
			proxyRequest(c, client, targetURL)
		})
		v1.PATCH("/sources/:id", func(c *gin.Context) {
			targetURL := articleServiceURL + "/api/v1/sources/" + c.Param("id")
			proxyRequest(c, client, targetURL)
		})
		v1.DELETE("/sources/:id", func(c *gin.Context) {
			targetURL := articleServiceURL + "/api/v1/sources/" + c.Param("id")
			proxyRequest(c, client, targetURL)
		})
	}
}

func proxyRequest(c *gin.Context, client *http.Client, targetURL string) {
	parsed, err := url.Parse(targetURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	parsed.RawQuery = c.Request.URL.RawQuery

	req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, parsed.String(), c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// gateway は透過プロキシとして振る舞い、header と body をそのまま upstream に渡す。
	req.Header = c.Request.Header.Clone()

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
