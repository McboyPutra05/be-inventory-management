package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	RequestIDHeader = "X-Request-ID"
	RequestIDKey    = "request_id"
)

type contextKey string

const ContextKeyRequestID contextKey = "request_id"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader(RequestIDHeader)
		if reqID == "" {
			reqID = uuid.NewString()
		}

		c.Header(RequestIDHeader, reqID)
		c.Set(RequestIDKey, reqID)

		ctx := context.WithValue(c.Request.Context(), ContextKeyRequestID, reqID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

