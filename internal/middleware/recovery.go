package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/intern/inventory-api/internal/pkg/httpx"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				slog.ErrorContext(
					c.Request.Context(),
					"panic recovered in HTTP handler",
					"panic", fmt.Sprint(r),
					"stack", stack,
				)

				httpx.SendError(c, http.StatusInternalServerError, httpx.CodeInternalError, "Terjadi kesalahan pada server")
				c.Abort()
			}
		}()
		c.Next()
	}
}

