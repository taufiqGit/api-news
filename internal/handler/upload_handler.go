package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/taufiqgit/news-api/internal/handler/response"
	"github.com/taufiqgit/news-api/internal/storage"
)

type UploadHandler struct {
	storage storage.Storage
}

func NewUploadHandler(s storage.Storage) *UploadHandler {
	return &UploadHandler{storage: s}
}

// UploadImage POST /api/v1/uploads/images
//
//	@Summary		Upload gambar
//	@Description	Upload gambar (jpeg/png/webp/gif/avif, max 5MB). Mengembalikan URL publik untuk dipakai di artikel.
//	@Tags			Uploads
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			file	formData	file	true	"File gambar (jpeg/png/webp/gif/avif)"
//	@Param			folder	formData	string	false	"Folder tujuan (default: uploads)"
//	@Success		201		{object}	response.Response	"Upload berhasil, berisi url & key"
//	@Failure		400		{object}	response.Response	"File tidak valid"
//	@Router			/uploads/images [post]
func (h *UploadHandler) UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file field 'file' is required", err.Error())
		return
	}
	defer file.Close()

	folder := c.DefaultPostForm("folder", "uploads")

	info, err := h.storage.Upload(c.Request.Context(), file, header, folder)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUnsupportedType):
			response.BadRequest(c, "unsupported file type. Use jpeg/png/webp/gif/avif", nil)
		case errors.Is(err, storage.ErrFileTooLarge):
			response.BadRequest(c, "file too large. Max 20MB", nil)
		default:
			response.InternalError(c, "failed to upload file")
		}
		return
	}

	response.Created(c, "file uploaded", info)
}
