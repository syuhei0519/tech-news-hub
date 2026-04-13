package httpapi

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"tech-feed-hub/article-service/internal/domain"
	"tech-feed-hub/article-service/internal/service"
)

func NewRouter(db *sql.DB, articleService *service.ArticleService) *gin.Engine {
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		if err := db.PingContext(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := router.Group("/api/v1")
	{
		v1.GET("/articles", func(c *gin.Context) {
			sourceID, _ := strconv.ParseInt(c.Query("source_id"), 10, 64)
			page, _ := strconv.Atoi(defaultString(c.Query("page"), "1"))
			pageSize, _ := strconv.Atoi(defaultString(c.Query("page_size"), "20"))

			result, err := articleService.ListArticles(c.Request.Context(), domain.ListArticlesParams{
				Query:    c.Query("q"),
				Category: c.Query("category"),
				SourceID: sourceID,
				Page:     page,
				PageSize: pageSize,
				Sort:     c.Query("sort"),
				Order:    c.Query("order"),
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, result)
		})

		v1.GET("/articles/:id", func(c *gin.Context) {
			id, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
				return
			}

			article, err := articleService.GetArticle(c.Request.Context(), id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if article == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
				return
			}

			c.JSON(http.StatusOK, article)
		})
	}

	internal := router.Group("/internal")
	{
		internal.POST("/ingest", func(c *gin.Context) {
			var req service.IngestRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			result, err := articleService.Ingest(c.Request.Context(), req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusAccepted, result)
		})
	}

	return router
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
