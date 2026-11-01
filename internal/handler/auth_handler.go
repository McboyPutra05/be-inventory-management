package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/intern/inventory-api/internal/dto"
	"github.com/intern/inventory-api/internal/pkg/httpx"
	"github.com/intern/inventory-api/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.HandleError(c, err)
		return
	}

	res, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	httpx.Created(c, res)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.HandleError(c, err)
		return
	}

	res, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	httpx.Ok(c, res)
}

