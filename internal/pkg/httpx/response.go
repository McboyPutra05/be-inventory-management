package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResponseSuccess struct {
	Data any `json:"data"`
}

type ResponseSuccessWithMeta struct {
	Data any `json:"data"`
	Meta any `json:"meta"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorBody struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

type ResponseError struct {
	Error ErrorBody `json:"error"`
}

func Ok(c *gin.Context, data any) {
	c.JSON(http.StatusOK, ResponseSuccess{Data: data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, ResponseSuccess{Data: data})
}

func OkWithMeta(c *gin.Context, data any, meta any) {
	c.JSON(http.StatusOK, ResponseSuccessWithMeta{
		Data: data,
		Meta: meta,
	})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func SendError(c *gin.Context, httpStatus int, code, message string, details ...FieldError) {
	c.JSON(httpStatus, ResponseError{
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

