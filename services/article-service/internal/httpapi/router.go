package httpapi

import (
	"database/sql"
	"errors"
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

		v1.GET("/sources", func(c *gin.Context) {
			sources, err := articleService.ListSources(c.Request.Context())
			if err != nil {
				writeServiceError(c, err)
				return
			}
			c.JSON(http.StatusOK, gin.H{"items": sources})
		})

		v1.GET("/sources/:id", func(c *gin.Context) {
			id, err := parseIDParam(c, "source id")
			if err != nil {
				return
			}

			source, err := articleService.GetSource(c.Request.Context(), id)
			if err != nil {
				writeServiceError(c, err)
				return
			}
			c.JSON(http.StatusOK, source)
		})

		v1.GET("/fetch-jobs", func(c *gin.Context) {
			sourceID, _ := strconv.ParseInt(c.Query("source_id"), 10, 64)
			page, _ := strconv.Atoi(defaultString(c.Query("page"), "1"))
			pageSize, _ := strconv.Atoi(defaultString(c.Query("page_size"), "20"))

			result, err := articleService.ListFetchJobs(c.Request.Context(), domain.ListFetchJobsParams{
				SourceID: sourceID,
				Status:   c.Query("status"),
				Page:     page,
				PageSize: pageSize,
			})
			if err != nil {
				writeServiceError(c, err)
				return
			}

			c.JSON(http.StatusOK, result)
		})

		v1.POST("/sources", func(c *gin.Context) {
			var req service.SourceInput
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			source, err := articleService.CreateSource(c.Request.Context(), req)
			if err != nil {
				writeServiceError(c, err)
				return
			}
			c.JSON(http.StatusCreated, source)
		})

		v1.PATCH("/sources/:id", func(c *gin.Context) {
			id, err := parseIDParam(c, "source id")
			if err != nil {
				return
			}

			var req service.SourceInput
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			source, err := articleService.UpdateSource(c.Request.Context(), id, req)
			if err != nil {
				writeServiceError(c, err)
				return
			}
			c.JSON(http.StatusOK, source)
		})

		v1.DELETE("/sources/:id", func(c *gin.Context) {
			id, err := parseIDParam(c, "source id")
			if err != nil {
				return
			}

			if err := articleService.DeleteSource(c.Request.Context(), id); err != nil {
				writeServiceError(c, err)
				return
			}
			c.Status(http.StatusNoContent)
		})
	}

	internal := router.Group("/internal")
	{
		internal.POST("/fetch-jobs/start", func(c *gin.Context) {
			var req service.StartFetchJobInput
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			result, err := articleService.StartFetchJob(c.Request.Context(), req)
			if err != nil {
				writeServiceError(c, err)
				return
			}
			c.JSON(http.StatusAccepted, result)
		})

		internal.POST("/fetch-jobs/:id/finish", func(c *gin.Context) {
			jobID, err := parseIDParam(c, "fetch job id")
			if err != nil {
				return
			}

			var req service.FinishFetchJobInput
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if err := articleService.FinishFetchJob(c.Request.Context(), jobID, req); err != nil {
				writeServiceError(c, err)
				return
			}
			c.Status(http.StatusNoContent)
		})

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

func parseIDParam(c *gin.Context, label string) (int64, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + label})
		return 0, err
	}
	return id, nil
}

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
