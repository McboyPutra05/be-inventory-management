package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intern/inventory-api/internal/pkg/httpx"
)

type HealthHandler struct {
	sqlDB *sql.DB
}

func NewHealthHandler(sqlDB *sql.DB) *HealthHandler {
	return &HealthHandler{sqlDB: sqlDB}
}

func (h *HealthHandler) Check(c *gin.Context) {
	if err := h.sqlDB.PingContext(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "database unreachable",
		})
		return
	}

	httpx.Ok(c, gin.H{
		"status":   "ok",
		"database": "connected",
	})
}

