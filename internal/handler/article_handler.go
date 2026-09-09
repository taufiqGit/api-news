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

type ArticleHandler struct {
	articleUsecase usecase.ArticleUsecase
}

func NewArticleHandler(uc usecase.ArticleUsecase) *ArticleHandler {
	return &ArticleHandler{articleUsecase: uc}
}

// Create POST /api/v1/articles (auth: author/editor/admin)
//
//	@Summary		Buat artikel
//	@Description	Membuat artikel baru (role: author/editor/admin)
//	@Tags			Articles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		entity.ArticleCreate	true	"Data artikel"
//	@Success		201		{object}	response.Response	"Artikel berhasil dibuat"
//	@Failure		400		{object}	response.Response	"Request tidak valid"
//	@Router			/articles [post]
func (h *ArticleHandler) Create(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		response.Unauthorized(c, "invalid user id")
		return
	}

	var req entity.ArticleCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	article, err := h.articleUsecase.Create(c.Request.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidStatus):
			response.BadRequest(c, "invalid status", nil)
		default:
			response.InternalError(c, "failed to create article")
		}
		return
	}

	response.Created(c, "article created", article)
}

// GetByID GET /api/v1/articles/:id
//
//	@Summary		Detail artikel by ID
//	@Description	Mengambil detail artikel berdasarkan ID
//	@Tags			Articles
//	@Produce		json
//	@Param			id	path	string	true	"ID artikel"
//	@Success		200	{object}	response.Response	"Data artikel"
//	@Failure		404	{object}	response.Response	"Artikel tidak ditemukan"
//	@Router			/articles/{id} [get]
func (h *ArticleHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid article id", nil)
		return
	}

	article, err := h.articleUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "article not found")
		} else {
			response.InternalError(c, "failed to get article")
		}
		return
	}

	response.Success(c, http.StatusOK, "article fetched", article)
}

// GetBySlug GET /api/v1/articles/slug/:slug
//
//	@Summary		Detail artikel by slug
//	@Description	Mengambil detail artikel berdasarkan slug, view count otomatis bertambah
//	@Tags			Articles
//	@Produce		json
//	@Param			slug	path	string	true	"Slug artikel"
//	@Success		200	{object}	response.Response	"Data artikel"
//	@Failure		404	{object}	response.Response	"Artikel tidak ditemukan"
//	@Router			/articles/slug/{slug} [get]
func (h *ArticleHandler) GetBySlug(c *gin.Context) {
	article, err := h.articleUsecase.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "article not found")
		} else {
			response.InternalError(c, "failed to get article")
		}
		return
	}

	// Hitung view count saat artikel diakses publik
	_ = h.articleUsecase.IncrementView(c.Request.Context(), article.ID)

	response.Success(c, http.StatusOK, "article fetched", article)
}

// List GET /api/v1/articles?page=1&limit=10&status=published&category_id=...&website_id=...
//
//	@Summary		List artikel
//	@Description	Menampilkan daftar artikel dengan filter & pagination
//	@Tags			Articles
//	@Produce		json
//	@Param			page		query		int		false	"Nomor halaman"				default(1)
//	@Param			limit		query		int		false	"Jumlah per halaman"		default(10)
//	@Param			status		query		string	false	"Filter status: draft/published/archived"
//	@Param			category_id	query		string	false	"Filter kategori (UUID)"
//	@Param			website_id	query		string	false	"Filter website (UUID)"
//	@Success		200			{object}	response.Response	"Daftar artikel dengan meta pagination"
//	@Router			/articles [get]
func (h *ArticleHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var categoryID *uuid.UUID
	if raw := c.Query("category_id"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			categoryID = &id
		}
	}

	var websiteID *uuid.UUID
	if raw := c.Query("website_id"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			websiteID = &id
		}
	}

	query := entity.ArticleQuery{
		Page:       page,
		Limit:      limit,
		Offset:     (page - 1) * limit,
		Status:     c.Query("status"),
		CategoryID: categoryID,
		WebsiteID:  websiteID,
	}

	items, total, err := h.articleUsecase.List(c.Request.Context(), query)
	if err != nil {
		response.InternalError(c, "failed to list articles")
		return
	}

	meta := &response.Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	}
	response.SuccessWithMeta(c, http.StatusOK, "articles fetched", items, meta)
}

// ListPublished GET /api/v1/articles/published?page=1&limit=10&website_id=...
//
//	@Summary		List artikel published
//	@Description	Menampilkan artikel yang sudah published (untuk publik), bisa difilter per website
//	@Tags			Articles
//	@Produce		json
//	@Param			page		query	int	false	"Nomor halaman"			default(1)
//	@Param			limit		query	int	false	"Jumlah per halaman"	default(10)
//	@Param			website_id	query	string	false	"Filter website (UUID)"
//	@Success		200			{object}	response.Response	"Daftar artikel published"
//	@Router			/articles/published [get]
func (h *ArticleHandler) ListPublished(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var websiteID *uuid.UUID
	if raw := c.Query("website_id"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			websiteID = &id
		}
	}

	items, err := h.articleUsecase.ListPublished(c.Request.Context(), page, limit, websiteID)
	if err != nil {
		response.InternalError(c, "failed to list published articles")
		return
	}

	response.Success(c, http.StatusOK, "published articles fetched", items)
}

// Update PUT /api/v1/articles/:id
//
//	@Summary		Update artikel
//	@Description	Mengubah data artikel (role: author/editor/admin)
//	@Tags			Articles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string				true	"ID artikel"
//	@Param			body	body	entity.ArticleUpdate	true	"Data yang diubah"
//	@Success		200		{object}	response.Response	"Artikel berhasil diubah"
//	@Failure		404		{object}	response.Response	"Artikel tidak ditemukan"
//	@Router			/articles/{id} [put]
func (h *ArticleHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid article id", nil)
		return
	}

	var req entity.ArticleUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	article, err := h.articleUsecase.Update(c.Request.Context(), id, req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			response.NotFound(c, "article not found")
		case errors.Is(err, usecase.ErrInvalidStatus):
			response.BadRequest(c, "invalid status", nil)
		default:
			response.InternalError(c, "failed to update article")
		}
		return
	}

	response.Success(c, http.StatusOK, "article updated", article)
}

// UpdateStatus PATCH /api/v1/articles/:id/status
//
//	@Summary		Update status artikel
//	@Description	Mengubah status artikel: draft/published/archived
//	@Tags			Articles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string	true	"ID artikel"
//	@Param			body	body	object{status=string}	true	"Status baru"
//	@Success		200		{object}	response.Response	"Status berhasil diubah"
//	@Failure		400		{object}	response.Response	"Status tidak valid"
//	@Router			/articles/{id}/status [patch]
func (h *ArticleHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid article id", nil)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	article, err := h.articleUsecase.UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			response.NotFound(c, "article not found")
		case errors.Is(err, usecase.ErrInvalidStatus):
			response.BadRequest(c, "invalid status", nil)
		default:
			response.InternalError(c, "failed to update article status")
		}
		return
	}

	response.Success(c, http.StatusOK, "article status updated", article)
}

// Delete DELETE /api/v1/articles/:id
//
//	@Summary		Hapus artikel
//	@Description	Menghapus artikel (role: author/editor/admin)
//	@Tags			Articles
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"ID artikel"
//	@Success		200	{object}	response.Response	"Artikel berhasil dihapus"
//	@Failure		404	{object}	response.Response	"Artikel tidak ditemukan"
//	@Router			/articles/{id} [delete]
func (h *ArticleHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid article id", nil)
		return
	}

	if err := h.articleUsecase.Delete(c.Request.Context(), id); err != nil {
		response.InternalError(c, "failed to delete article")
		return
	}

	response.Success(c, http.StatusOK, "article deleted", nil)
}
