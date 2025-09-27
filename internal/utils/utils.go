package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ParseUUIDParam parses UUID from URL parameter
func ParseUUIDParam(c *gin.Context, paramName string) (uuid.UUID, error) {
	paramValue := c.Param(paramName)
	return uuid.Parse(paramValue)
}