package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response adalah format standar API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Success mengirim response sukses
func Success(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessWithMeta mengirim response sukses dengan pagination
func SuccessWithMeta(c *gin.Context, status int, message string, data interface{}, meta *Meta) {
	c.JSON(status, Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Created mengirim response 201
func Created(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusCreated, message, data)
}

// Error mengirim response error
func Error(c *gin.Context, status int, message string, errors interface{}) {
	c.JSON(status, Response{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

// BadRequest mengirim response 400
func BadRequest(c *gin.Context, message string, errors interface{}) {
	Error(c, http.StatusBadRequest, message, errors)
}

// Unauthorized mengirim response 401
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message, nil)
}

// Forbidden mengirim response 403
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message, nil)
}

// NotFound mengirim response 404
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message, nil)
}

// InternalError mengirim response 500
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message, nil)
}
