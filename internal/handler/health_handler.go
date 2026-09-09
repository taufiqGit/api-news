package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/taufiqgit/news-api/internal/handler/response"
)

// HealthCheck GET /healthz
//
//	@Summary		Health check
//	@Description	Cek status kesehatan server dan koneksi database
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	response.Response
//	@Router			/healthz [get]
func HealthCheck(c *gin.Context) {
	response.Success(c, http.StatusOK, "ok", gin.H{
		"status":   "healthy",
		"time":     time.Now().Format(time.RFC3339),
		"service":  "news-api",
		"hot_reload": true,
		"version":  "1.0.2",
	})
}
