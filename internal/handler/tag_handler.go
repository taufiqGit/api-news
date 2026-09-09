package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/taufiqgit/news-api/internal/domain/entity"
	"github.com/taufiqgit/news-api/internal/handler/response"
	"github.com/taufiqgit/news-api/internal/repository"
	"github.com/taufiqgit/news-api/internal/usecase"
)

type TagHandler struct {
	tagUsecase usecase.TagUsecase
}

func NewTagHandler(uc usecase.TagUsecase) *TagHandler {
	return &TagHandler{tagUsecase: uc}
}

// Create POST /api/v1/tags
//
//	@Summary		Buat tag
//	@Description	Membuat tag baru (role: editor/admin)
//	@Tags			Tags
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		entity.TagCreate	true	"Data tag"
//	@Success		201		{object}	response.Response	"Tag berhasil dibuat"
//	@Failure		400		{object}	response.Response	"Request tidak valid"
//	@Router			/tags [post]
func (h *TagHandler) Create(c *gin.Context) {
	var req entity.TagCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	tag, err := h.tagUsecase.Create(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrSlugTaken):
			response.BadRequest(c, "slug already used", nil)
		default:
			response.InternalError(c, "failed to create tag")
		}
		return
	}

	response.Created(c, "tag created", tag)
}

// List GET /api/v1/tags
//
//	@Summary		List tag
//	@Description	Menampilkan semua tag
//	@Tags			Tags
//	@Produce		json
//	@Success		200	{object}	response.Response	"Daftar tag"
//	@Router			/tags [get]
func (h *TagHandler) List(c *gin.Context) {
	tags, err := h.tagUsecase.List(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to list tags")
		return
	}

	response.Success(c, http.StatusOK, "tags fetched", tags)
}

// Update PUT /api/v1/tags/:id
//
//	@Summary		Update tag
//	@Description	Mengubah data tag (role: editor/admin)
//	@Tags			Tags
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string			true	"ID tag"
//	@Param			body	body	entity.TagUpdate	true	"Data yang diubah"
//	@Success		200		{object}	response.Response	"Tag berhasil diubah"
//	@Failure		404		{object}	response.Response	"Tag tidak ditemukan"
//	@Router			/tags/{id} [put]
func (h *TagHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid tag id", nil)
		return
	}

	var req entity.TagUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	tag, err := h.tagUsecase.Update(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "tag not found")
		} else {
			response.InternalError(c, "failed to update tag")
		}
		return
	}

	response.Success(c, http.StatusOK, "tag updated", tag)
}

// Delete DELETE /api/v1/tags/:id
//
//	@Summary		Hapus tag
//	@Description	Menghapus tag (role: editor/admin)
//	@Tags			Tags
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"ID tag"
//	@Success		200	{object}	response.Response	"Tag berhasil dihapus"
//	@Failure		404	{object}	response.Response	"Tag tidak ditemukan"
//	@Router			/tags/{id} [delete]
func (h *TagHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid tag id", nil)
		return
	}

	if err := h.tagUsecase.Delete(c.Request.Context(), id); err != nil {
		response.InternalError(c, "failed to delete tag")
		return
	}

	response.Success(c, http.StatusOK, "tag deleted", nil)
}
