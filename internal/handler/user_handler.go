package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/taufiqgit/news-api/internal/domain/entity"
	"github.com/taufiqgit/news-api/internal/handler/response"
	"github.com/taufiqgit/news-api/internal/middleware"
	"github.com/taufiqgit/news-api/internal/usecase"
	"github.com/taufiqgit/news-api/internal/repository"
)

type UserHandler struct {
	userUsecase usecase.UserUsecase
	tokenMgr    *middleware.TokenManager
}

func NewUserHandler(uc usecase.UserUsecase, tm *middleware.TokenManager) *UserHandler {
	return &UserHandler{userUsecase: uc, tokenMgr: tm}
}

// Register POST /api/v1/auth/register
//
//	@Summary		Register user baru
//	@Description	Mendaftarkan user baru (default role: author)
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		entity.UserCreate	true	"Data registrasi"
//	@Success		201		{object}	response.Response	"User berhasil dibuat"
//	@Failure		400		{object}	response.Response	"Request tidak valid"
//	@Failure		500		{object}	response.Response	"Internal server error"
//	@Router			/auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req entity.UserCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	user, err := h.userUsecase.Register(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrEmailTaken):
			response.BadRequest(c, "email already registered", nil)
		default:
			response.InternalError(c, "failed to register user")
		}
		return
	}

	response.Created(c, "user registered", user)
}

// Login POST /api/v1/auth/login
//
//	@Summary		Login user
//	@Description	Login dengan email & password, mengembalikan JWT token
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		entity.LoginRequest	true	"Kredensial login"
//	@Success		200		{object}	response.Response	"Login sukses, berisi token & user"
//	@Failure		400		{object}	response.Response	"Request tidak valid"
//	@Failure		401		{object}	response.Response	"Email atau password salah"
//	@Router			/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req entity.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	user, err := h.userUsecase.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			response.Unauthorized(c, "invalid email or password")
		} else {
			response.InternalError(c, "failed to login")
		}
		return
	}

	token, err := h.tokenMgr.Generate(user.ID.String(), user.Role)
	if err != nil {
		response.InternalError(c, "failed to generate token")
		return
	}

	response.Success(c, http.StatusOK, "login success", gin.H{
		"token": token,
		"user":  user,
	})
}

// GetMe GET /api/v1/users/me
//
//	@Summary		Profil user saat ini
//	@Description	Mengambil data user berdasarkan token JWT
//	@Tags			Users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Response	"Data user"
//	@Failure		401	{object}	response.Response	"Tidak terautentikasi"
//	@Router			/users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		response.Unauthorized(c, "invalid user id")
		return
	}

	user, err := h.userUsecase.GetByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "user not found")
		} else {
			response.InternalError(c, "failed to get user")
		}
		return
	}

	response.Success(c, http.StatusOK, "user fetched", user)
}

// List GET /api/v1/users?page=1&limit=10
//
//	@Summary		List user
//	@Description	Menampilkan daftar user (hanya admin)
//	@Tags			Users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int	false	"Nomor halaman"	default(1)
//	@Param			limit	query		int	false	"Jumlah per halaman"	default(10)
//	@Success		200		{object}	response.Response	"Daftar user"
//	@Failure		403		{object}	response.Response	"Forbidden"
//	@Router			/users [get]
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	users, err := h.userUsecase.List(c.Request.Context(), page, limit)
	if err != nil {
		response.InternalError(c, "failed to list users")
		return
	}

	response.Success(c, http.StatusOK, "users fetched", users)
}

// Update PUT /api/v1/users/:id
//
//	@Summary		Update user
//	@Description	Mengubah data user (hanya admin)
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string				true	"ID user"
//	@Param			body	body	entity.UserUpdate	true	"Data yang diubah"
//	@Success		200		{object}	response.Response	"User berhasil diubah"
//	@Failure		404		{object}	response.Response	"User tidak ditemukan"
//	@Router			/users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid user id", nil)
		return
	}

	var req entity.UserUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	user, err := h.userUsecase.Update(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "user not found")
		} else {
			response.InternalError(c, "failed to update user")
		}
		return
	}

	response.Success(c, http.StatusOK, "user updated", user)
}

// Delete DELETE /api/v1/users/:id
//
//	@Summary		Hapus user
//	@Description	Menghapus user (hanya admin)
//	@Tags			Users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"ID user"
//	@Success		200	{object}	response.Response	"User berhasil dihapus"
//	@Failure		404	{object}	response.Response	"User tidak ditemukan"
//	@Router			/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid user id", nil)
		return
	}

	if err := h.userUsecase.Delete(c.Request.Context(), id); err != nil {
		response.InternalError(c, "failed to delete user")
		return
	}

	response.Success(c, http.StatusOK, "user deleted", nil)
}
