package handler

import (
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/taufiqgit/news-api/internal/domain/entity"
	"github.com/taufiqgit/news-api/internal/handler/response"
	"github.com/taufiqgit/news-api/internal/repository"
	"github.com/taufiqgit/news-api/internal/usecase"
)

type WebsiteHandler struct {
	websiteUsecase usecase.WebsiteUsecase
}

func NewWebsiteHandler(uc usecase.WebsiteUsecase) *WebsiteHandler {
	return &WebsiteHandler{websiteUsecase: uc}
}

// Create POST /api/v1/websites
//
//	@Summary		Buat website
//	@Description	Membuat website baru (role: admin). Setiap website punya artikel sendiri-sendiri.
//	@Tags			Websites
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		entity.WebsiteCreate	true	"Data website"
//	@Success		201		{object}	response.Response	"Website berhasil dibuat"
//	@Failure		400		{object}	response.Response	"Request tidak valid"
//	@Router			/websites [post]
func (h *WebsiteHandler) Create(c *gin.Context) {
	var req entity.WebsiteCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	website, err := h.websiteUsecase.Create(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrWebsiteSlugTaken):
			response.BadRequest(c, "website slug already used", nil)
		default:
			response.InternalError(c, "failed to create website")
		}
		return
	}

	response.Created(c, "website created", website)
}

// GetByID GET /api/v1/websites/:id
//
//	@Summary		Detail website by ID
//	@Description	Mengambil detail website berdasarkan ID
//	@Tags			Websites
//	@Produce		json
//	@Param			id	path	string	true	"ID website"
//	@Success		200	{object}	response.Response	"Data website"
//	@Failure		404	{object}	response.Response	"Website tidak ditemukan"
//	@Router			/websites/{id} [get]
func (h *WebsiteHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid website id", nil)
		return
	}

	website, err := h.websiteUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "website not found")
		} else {
			response.InternalError(c, "failed to get website")
		}
		return
	}

	response.Success(c, http.StatusOK, "website fetched", website)
}

// GetBySlug GET /api/v1/websites/slug/:slug
//
//	@Summary		Detail website by slug
//	@Description	Mengambil detail website berdasarkan slug
//	@Tags			Websites
//	@Produce		json
//	@Param			slug	path	string	true	"Slug website"
//	@Success		200	{object}	response.Response	"Data website"
//	@Failure		404	{object}	response.Response	"Website tidak ditemukan"
//	@Router			/websites/slug/{slug} [get]
func (h *WebsiteHandler) GetBySlug(c *gin.Context) {
	website, err := h.websiteUsecase.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "website not found")
		} else {
			response.InternalError(c, "failed to get website")
		}
		return
	}

	response.Success(c, http.StatusOK, "website fetched", website)
}

// List GET /api/v1/websites?page=1&limit=10
//
//	@Summary		List website
//	@Description	Menampilkan daftar semua website dengan pagination
//	@Tags			Websites
//	@Produce		json
//	@Param			page	query		int	false	"Nomor halaman"	default(1)
//	@Param			limit	query		int	false	"Jumlah per halaman"	default(10)
//	@Success		200		{object}	response.Response	"Daftar website dengan meta pagination"
//	@Router			/websites [get]
func (h *WebsiteHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	items, total, err := h.websiteUsecase.List(c.Request.Context(), page, limit)
	if err != nil {
		response.InternalError(c, "failed to list websites")
		return
	}

	meta := &response.Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	}
	response.SuccessWithMeta(c, http.StatusOK, "websites fetched", items, meta)
}

// Update PUT /api/v1/websites/:id
//
//	@Summary		Update website
//	@Description	Mengubah data website (role: admin)
//	@Tags			Websites
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string				true	"ID website"
//	@Param			body	body	entity.WebsiteUpdate	true	"Data yang diubah"
//	@Success		200		{object}	response.Response	"Website berhasil diubah"
//	@Failure		404		{object}	response.Response	"Website tidak ditemukan"
//	@Router			/websites/{id} [put]
func (h *WebsiteHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid website id", nil)
		return
	}

	var req entity.WebsiteUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	website, err := h.websiteUsecase.Update(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "website not found")
		} else {
			response.InternalError(c, "failed to update website")
		}
		return
	}

	response.Success(c, http.StatusOK, "website updated", website)
}

// Delete DELETE /api/v1/websites/:id
//
//	@Summary		Hapus website
//	@Description	Menghapus website beserta seluruh artikelnya (role: admin)
//	@Tags			Websites
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"ID website"
//	@Success		200	{object}	response.Response	"Website berhasil dihapus"
//	@Failure		404	{object}	response.Response	"Website tidak ditemukan"
//	@Router			/websites/{id} [delete]
func (h *WebsiteHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid website id", nil)
		return
	}

	if err := h.websiteUsecase.Delete(c.Request.Context(), id); err != nil {
		response.InternalError(c, "failed to delete website")
		return
	}

	response.Success(c, http.StatusOK, "website deleted", nil)
}
