package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/intern/inventory-api/internal/dto"
	"github.com/intern/inventory-api/internal/middleware"
	"github.com/intern/inventory-api/internal/pkg/httpx"
	"github.com/intern/inventory-api/internal/service"
)

type ItemHandler struct {
	itemService service.ItemService
}

func NewItemHandler(itemService service.ItemService) *ItemHandler {
	return &ItemHandler{itemService: itemService}
}

func (h *ItemHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		httpx.SendError(c, http.StatusUnauthorized, httpx.CodeUnauthorized, "User ID not found in session")
		return
	}

	var req dto.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.HandleError(c, err)
		return
	}

	res, err := h.itemService.Create(c.Request.Context(), userID, &req)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	httpx.Created(c, res)
}

func (h *ItemHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.SendError(c, http.StatusBadRequest, httpx.CodeBadRequest, "ID harus berupa UUID yang valid")
		return
	}

	res, err := h.itemService.Get(c.Request.Context(), id)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	httpx.Ok(c, res)
}

func (h *ItemHandler) List(c *gin.Context) {
	var query dto.ListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		httpx.HandleError(c, err)
		return
	}

	items, meta, err := h.itemService.List(c.Request.Context(), &query)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	httpx.OkWithMeta(c, items, meta)
}

func (h *ItemHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.SendError(c, http.StatusBadRequest, httpx.CodeBadRequest, "ID harus berupa UUID yang valid")
		return
	}

	var req dto.UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.HandleError(c, err)
		return
	}

	res, err := h.itemService.Update(c.Request.Context(), id, &req)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	httpx.Ok(c, res)
}

func (h *ItemHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.SendError(c, http.StatusBadRequest, httpx.CodeBadRequest, "ID harus berupa UUID yang valid")
		return
	}

	if err := h.itemService.Delete(c.Request.Context(), id); err != nil {
		httpx.HandleError(c, err)
		return
	}

	httpx.NoContent(c)
}

func (h *ItemHandler) AdjustStock(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.SendError(c, http.StatusBadRequest, httpx.CodeBadRequest, "ID harus berupa UUID yang valid")
		return
	}

	var req dto.AdjustStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.HandleError(c, err)
		return
	}

	res, err := h.itemService.AdjustStock(c.Request.Context(), id, req.Delta)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	httpx.Ok(c, res)
}

