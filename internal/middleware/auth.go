package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/intern/inventory-api/internal/pkg/httpx"
	"github.com/intern/inventory-api/internal/pkg/token"
)

const (
	UserIDKey = "user_id"
)

type contextKeyUser string

const ContextKeyUserID contextKeyUser = "user_id"

func Auth(tokenManager *token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			httpx.SendError(c, http.StatusUnauthorized, httpx.CodeUnauthorized, "Token autentikasi tidak ditemukan")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			httpx.SendError(c, http.StatusUnauthorized, httpx.CodeUnauthorized, "Format token harus Bearer <token>")
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		userID, err := tokenManager.Verify(tokenString)
		if err != nil {
			httpx.SendError(c, http.StatusUnauthorized, httpx.CodeUnauthorized, "Token tidak valid atau telah kedaluwarsa")
			c.Abort()
			return
		}

		c.Set(UserIDKey, userID)
		ctx := context.WithValue(c.Request.Context(), ContextKeyUserID, userID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(UserIDKey)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

