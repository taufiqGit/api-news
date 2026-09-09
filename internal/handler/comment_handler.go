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

type CommentHandler struct {
	commentUsecase usecase.CommentUsecase
}

func NewCommentHandler(uc usecase.CommentUsecase) *CommentHandler {
	return &CommentHandler{commentUsecase: uc}
}

// Create POST /api/v1/comments
//
//	@Summary		Buat komentar
//	@Description	Membuat komentar pada artikel (perlu login)
//	@Tags			Comments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		entity.CommentCreate	true	"Data komentar"
//	@Success		201		{object}	response.Response	"Komentar berhasil dibuat"
//	@Failure		400		{object}	response.Response	"Request tidak valid"
//	@Failure		404		{object}	response.Response	"Artikel tidak ditemukan"
//	@Router			/comments [post]
func (h *CommentHandler) Create(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		response.Unauthorized(c, "invalid user id")
		return
	}

	var req entity.CommentCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	comment, err := h.commentUsecase.Create(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "article not found")
		} else {
			response.InternalError(c, "failed to create comment")
		}
		return
	}

	response.Created(c, "comment created", comment)
}

// ListByArticle GET /api/v1/articles/:id/comments
//
//	@Summary		List komentar artikel
//	@Description	Menampilkan komentar pada sebuah artikel (publik)
//	@Tags			Comments
//	@Produce		json
//	@Param			id	path	string	true	"ID artikel"
//	@Success		200	{object}	response.Response	"Daftar komentar"
//	@Router			/articles/{id}/comments [get]
func (h *CommentHandler) ListByArticle(c *gin.Context) {
	articleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid article id", nil)
		return
	}

	comments, err := h.commentUsecase.ListByArticle(c.Request.Context(), articleID)
	if err != nil {
		response.InternalError(c, "failed to list comments")
		return
	}

	response.Success(c, http.StatusOK, "comments fetched", comments)
}

// Delete DELETE /api/v1/comments/:id
//
//	@Summary		Hapus komentar
//	@Description	Menghapus komentar (perlu login)
//	@Tags			Comments
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"ID komentar"
//	@Success		200	{object}	response.Response	"Komentar berhasil dihapus"
//	@Failure		404	{object}	response.Response	"Komentar tidak ditemukan"
//	@Router			/comments/{id} [delete]
func (h *CommentHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid comment id", nil)
		return
	}

	if err := h.commentUsecase.Delete(c.Request.Context(), id); err != nil {
		response.InternalError(c, "failed to delete comment")
		return
	}

	response.Success(c, http.StatusOK, "comment deleted", nil)
}
