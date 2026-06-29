// Package respond provides uniform JSON response helpers for Gin handlers.
package respond

import (
	"backend/src/internal/logger"
	"backend/src/pkg/apperr"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// OK sends 200 with data.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}

// Created sends 201 with data.
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, data)
}

// NoContent sends 204.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Err inspects the error:
//   - *apperr.AppError → uses its Code and Message
//   - anything else    → 500, logs the internal error
func Err(c *gin.Context, err error) {
	if ae := apperr.As(err); ae != nil {
		c.JSON(ae.Code, gin.H{"error": ae.Message})
		return
	}
	logger.Error("unhandled error", zap.String("path", c.FullPath()), zap.Error(err))
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
