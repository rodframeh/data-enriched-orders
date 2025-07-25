package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Code    int         `json:"code"`
}

// JSON sends a JSON response with raw data
func JSON(c *gin.Context, code int, data interface{}) {
	c.JSON(code, data)
}

// Error sends an error JSON response
func Error(c *gin.Context, code int, err string, message string) {
	c.JSON(code, ErrorResponse{
		Error:   err,
		Message: message,
		Code:    code,
	})
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "bad_request", message)
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, "not_found", message)
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, "internal_server_error", message)
}

// Conflict sends a 409 Conflict response
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, "conflict", message)
}

// Created sends a 201 Created response
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

// OK sends a 200 OK response
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}
