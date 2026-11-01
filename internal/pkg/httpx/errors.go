package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/intern/inventory-api/internal/domain"
	"github.com/intern/inventory-api/internal/pkg/token"
)

const (
	CodeBadRequest         = "BAD_REQUEST"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeNotFound           = "NOT_FOUND"
	CodeEmailAlreadyExists = "EMAIL_ALREADY_EXISTS"
	CodeInsufficientStock  = "INSUFFICIENT_STOCK"
	CodeValidationError    = "VALIDATION_ERROR"
	CodeInternalError      = "INTERNAL_ERROR"
)

func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	var valErrs validator.ValidationErrors
	if errors.As(err, &valErrs) {
		details := make([]FieldError, 0, len(valErrs))
		for _, fe := range valErrs {
			field := toSnakeCase(fe.Field())
			details = append(details, FieldError{
				Field:   field,
				Message: formatValidationMessage(field, fe),
			})
		}
		SendError(c, http.StatusUnprocessableEntity, CodeValidationError, "Data yang dikirim tidak valid", details...)
		return
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		fieldName := typeErr.Field
		if fieldName == "" {
			fieldName = "body"
		}
		SendError(c, http.StatusUnprocessableEntity, CodeValidationError, "Tipe data tidak valid", FieldError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s harus bertipe %s", fieldName, typeErr.Type.String()),
		})
		return
	}

	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		SendError(c, http.StatusBadRequest, CodeBadRequest, "Request body tidak boleh kosong")
		return
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		SendError(c, http.StatusBadRequest, CodeBadRequest, "Format JSON tidak valid")
		return
	}

	if errors.Is(err, domain.ErrNotFound) {
		SendError(c, http.StatusNotFound, CodeNotFound, "Data tidak ditemukan")
		return
	}

	if errors.Is(err, domain.ErrEmailTaken) {
		SendError(c, http.StatusConflict, CodeEmailAlreadyExists, "Email sudah terdaftar")
		return
	}

	if errors.Is(err, domain.ErrInsufficientStock) {
		SendError(c, http.StatusConflict, CodeInsufficientStock, "Stok tidak mencukupi untuk pengurangan ini")
		return
	}

	if errors.Is(err, domain.ErrInvalidCredentials) || errors.Is(err, token.ErrInvalidToken) {
		SendError(c, http.StatusUnauthorized, CodeUnauthorized, "Email atau password salah")
		return
	}

	slog.ErrorContext(c.Request.Context(), "unhandled error occurred", "error", err.Error())
	SendError(c, http.StatusInternalServerError, CodeInternalError, "Terjadi kesalahan pada server")
}

func formatValidationMessage(field string, fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s wajib diisi", field)
	case "email":
		return "format email tidak valid"
	case "min":
		if fe.Type().String() == "string" {
			return fmt.Sprintf("%s minimal %s karakter", field, fe.Param())
		}
		return fmt.Sprintf("%s minimal bernilai %s", field, fe.Param())
	case "max":
		if fe.Type().String() == "string" {
			return fmt.Sprintf("%s maksimal %s karakter", field, fe.Param())
		}
		return fmt.Sprintf("%s maksimal bernilai %s", field, fe.Param())
	case "gte":
		if fe.Param() == "0" {
			return fmt.Sprintf("%s tidak boleh negatif", field)
		}
		return fmt.Sprintf("%s minimal bernilai %s", field, fe.Param())
	case "ne":
		return fmt.Sprintf("%s tidak boleh bernilai %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s tidak valid (%s)", field, fe.Tag())
	}
}

func toSnakeCase(str string) string {
	var result strings.Builder
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

