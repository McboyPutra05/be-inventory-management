package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		if raw != "" {
			path = path + "?" + raw
		}

		reqID, _ := c.Get(RequestIDKey)
		userID, _ := c.Get("user_id")

		attrs := []any{
			"status", statusCode,
			"method", method,
			"path", path,
			"ip", clientIP,
			"latency", latency.String(),
			"request_id", reqID,
		}

		if userID != nil {
			attrs = append(attrs, "user_id", userID)
		}

		if len(c.Errors) > 0 {
			attrs = append(attrs, "errors", c.Errors.String())
		}

		if statusCode >= 500 {
			slog.ErrorContext(c.Request.Context(), "HTTP Server Error", attrs...)
		} else if statusCode >= 400 {
			slog.WarnContext(c.Request.Context(), "HTTP Client Error", attrs...)
		} else {
			slog.InfoContext(c.Request.Context(), "HTTP Request", attrs...)
		}
	}
}

