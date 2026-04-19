package httpapi

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, articleServiceURL string, notificationServiceURL string) {
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
		v1.PATCH("/articles/:id/read-status", func(c *gin.Context) {
			targetURL := articleServiceURL + "/api/v1/articles/" + c.Param("id") + "/read-status"
			proxyRequest(c, client, targetURL)
		})
		v1.PATCH("/articles/:id/favorite-status", func(c *gin.Context) {
			targetURL := articleServiceURL + "/api/v1/articles/" + c.Param("id") + "/favorite-status"
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
		v1.GET("/notifications", func(c *gin.Context) {
			targetURL := notificationServiceURL + "/api/v1/notifications"
			proxyRequest(c, client, targetURL)
		})
		v1.PATCH("/notifications/:id/read-status", func(c *gin.Context) {
			targetURL := notificationServiceURL + "/api/v1/notifications/" + c.Param("id") + "/read-status"
			proxyRequest(c, client, targetURL)
		})
		v1.GET("/exports/articles.csv", func(c *gin.Context) {
			targetURL := articleServiceURL + "/api/v1/articles/export.csv"
			filename := "articles-" + time.Now().UTC().Format("20060102-150405") + ".csv"
			proxyCSVDownload(c, client, targetURL, filename)
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
		c.JSON(proxyErrorStatus(err), gin.H{"error": err.Error()})
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

func proxyCSVDownload(c *gin.Context, client *http.Client, targetURL string, filename string) {
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
	req.Header = c.Request.Header.Clone()

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(proxyErrorStatus(err), gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	if resp.StatusCode >= http.StatusBadRequest {
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
		return
	}

	// ダウンロード時の見え方は gateway で固定し、upstream 側は CSV 本文の生成だけに寄せる。
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(resp.StatusCode, "text/csv; charset=utf-8", body)
}

func proxyErrorStatus(err error) int {
	if err == nil {
		return http.StatusBadGateway
	}

	var netErr net.Error
	switch {
	case err == context.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout
	case errors.As(err, &netErr) && netErr.Timeout():
		return http.StatusGatewayTimeout
	default:
		return http.StatusBadGateway
	}
}
