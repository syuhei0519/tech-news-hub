package httpapi

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"tech-feed-hub/notification-service/internal/domain"
	"tech-feed-hub/notification-service/internal/service"
)

func NewRouter(db *sql.DB, notificationService *service.NotificationService) *gin.Engine {
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
		v1.GET("/notifications", func(c *gin.Context) {
			isRead, err := parseOptionalBoolQuery(c, "is_read")
			if err != nil {
				return
			}

			page, _ := strconv.Atoi(defaultString(c.Query("page"), "1"))
			pageSize, _ := strconv.Atoi(defaultString(c.Query("page_size"), "20"))

			result, err := notificationService.ListNotifications(c.Request.Context(), domain.ListNotificationsParams{
				IsRead:   isRead,
				Page:     page,
				PageSize: pageSize,
			})
			if err != nil {
				writeServiceError(c, err)
				return
			}
			c.JSON(http.StatusOK, result)
		})

		v1.PATCH("/notifications/:id/read-status", func(c *gin.Context) {
			id, err := parseIDParam(c, "notification id")
			if err != nil {
				return
			}

			var req service.UpdateReadStatusInput
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			notification, err := notificationService.UpdateReadStatus(c.Request.Context(), id, req)
			if err != nil {
				writeServiceError(c, err)
				return
			}
			c.JSON(http.StatusOK, notification)
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

func parseOptionalBoolQuery(c *gin.Context, key string) (*bool, error) {
	raw := c.Query(key)
	if raw == "" {
		return nil, nil
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + key})
		return nil, err
	}
	return &value, nil
}

func parseIDParam(c *gin.Context, label string) (int64, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + label})
		return 0, errors.New("invalid id")
	}
	return id, nil
}

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
