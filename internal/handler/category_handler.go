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

type CategoryHandler struct {
	categoryUsecase usecase.CategoryUsecase
}

func NewCategoryHandler(uc usecase.CategoryUsecase) *CategoryHandler {
	return &CategoryHandler{categoryUsecase: uc}
}

// Create POST /api/v1/categories
//
//	@Summary		Buat kategori
//	@Description	Membuat kategori baru (role: editor/admin)
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		entity.CategoryCreate	true	"Data kategori"
//	@Success		201		{object}	response.Response	"Kategori berhasil dibuat"
//	@Failure		400		{object}	response.Response	"Request tidak valid"
//	@Router			/categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var req entity.CategoryCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	category, err := h.categoryUsecase.Create(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrSlugTaken):
			response.BadRequest(c, "slug already used", nil)
		default:
			response.InternalError(c, "failed to create category")
		}
		return
	}

	response.Created(c, "category created", category)
}

// GetByID GET /api/v1/categories/:id
//
//	@Summary		Detail kategori by ID
//	@Description	Mengambil detail kategori berdasarkan ID
//	@Tags			Categories
//	@Produce		json
//	@Param			id	path	string	true	"ID kategori"
//	@Success		200	{object}	response.Response	"Data kategori"
//	@Failure		404	{object}	response.Response	"Kategori tidak ditemukan"
//	@Router			/categories/{id} [get]
func (h *CategoryHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid category id", nil)
		return
	}

	category, err := h.categoryUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "category not found")
		} else {
			response.InternalError(c, "failed to get category")
		}
		return
	}

	response.Success(c, http.StatusOK, "category fetched", category)
}

// GetBySlug GET /api/v1/categories/slug/:slug
//
//	@Summary		Detail kategori by slug
//	@Description	Mengambil detail kategori berdasarkan slug
//	@Tags			Categories
//	@Produce		json
//	@Param			slug	path	string	true	"Slug kategori"
//	@Success		200	{object}	response.Response	"Data kategori"
//	@Failure		404	{object}	response.Response	"Kategori tidak ditemukan"
//	@Router			/categories/slug/{slug} [get]
func (h *CategoryHandler) GetBySlug(c *gin.Context) {
	category, err := h.categoryUsecase.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "category not found")
		} else {
			response.InternalError(c, "failed to get category")
		}
		return
	}

	response.Success(c, http.StatusOK, "category fetched", category)
}

// List GET /api/v1/categories?page=1&limit=10
//
//	@Summary		List kategori
//	@Description	Menampilkan daftar kategori dengan pagination
//	@Tags			Categories
//	@Produce		json
//	@Param			page	query		int	false	"Nomor halaman"	default(1)
//	@Param			limit	query		int	false	"Jumlah per halaman"	default(10)
//	@Success		200		{object}	response.Response	"Daftar kategori dengan meta pagination"
//	@Router			/categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	items, total, err := h.categoryUsecase.List(c.Request.Context(), page, limit)
	if err != nil {
		response.InternalError(c, "failed to list categories")
		return
	}

	meta := &response.Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	}
	response.SuccessWithMeta(c, http.StatusOK, "categories fetched", items, meta)
}

// Update PUT /api/v1/categories/:id
//
//	@Summary		Update kategori
//	@Description	Mengubah data kategori (role: editor/admin)
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string				true	"ID kategori"
//	@Param			body	body	entity.CategoryUpdate	true	"Data yang diubah"
//	@Success		200		{object}	response.Response	"Kategori berhasil diubah"
//	@Failure		404		{object}	response.Response	"Kategori tidak ditemukan"
//	@Router			/categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid category id", nil)
		return
	}

	var req entity.CategoryUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	category, err := h.categoryUsecase.Update(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "category not found")
		} else {
			response.InternalError(c, "failed to update category")
		}
		return
	}

	response.Success(c, http.StatusOK, "category updated", category)
}

// Delete DELETE /api/v1/categories/:id
//
//	@Summary		Hapus kategori
//	@Description	Menghapus kategori (role: editor/admin)
//	@Tags			Categories
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"ID kategori"
//	@Success		200	{object}	response.Response	"Kategori berhasil dihapus"
//	@Failure		404	{object}	response.Response	"Kategori tidak ditemukan"
//	@Router			/categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid category id", nil)
		return
	}

	if err := h.categoryUsecase.Delete(c.Request.Context(), id); err != nil {
		response.InternalError(c, "failed to delete category")
		return
	}

	response.Success(c, http.StatusOK, "category deleted", nil)
}
