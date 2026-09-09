package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/taufiqgit/news-api/internal/domain/entity"
	"github.com/taufiqgit/news-api/internal/domain/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken         = errors.New("email already registered")
)

type UserUsecase interface {
	Register(ctx context.Context, req entity.UserCreate) (*entity.User, error)
	Login(ctx context.Context, req entity.LoginRequest) (*entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	List(ctx context.Context, page, limit int) ([]entity.User, error)
	Update(ctx context.Context, id uuid.UUID, req entity.UserUpdate) (*entity.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type userUsecase struct {
	userRepo repository.UserRepository
}

func NewUserUsecase(userRepo repository.UserRepository) UserUsecase {
	return &userUsecase{userRepo: userRepo}
}

func (u *userUsecase) Register(ctx context.Context, req entity.UserCreate) (*entity.User, error) {
	// Cek email sudah terpakai
	if _, err := u.userRepo.GetByEmail(ctx, req.Email); err == nil {
		return nil, ErrEmailTaken
	}

	role := req.Role
	if role == "" {
		role = "author"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := entity.User{
		Name:      req.Name,
		Email:     strings.ToLower(req.Email),
		Password:  string(hash),
		Role:      role,
		AvatarURL: req.AvatarURL,
		IsActive:  true,
	}

	return u.userRepo.Create(ctx, user)
}

func (u *userUsecase) Login(ctx context.Context, req entity.LoginRequest) (*entity.User, error) {
	user, err := u.userRepo.GetByEmail(ctx, strings.ToLower(req.Email))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (u *userUsecase) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return u.userRepo.GetByID(ctx, id)
}

func (u *userUsecase) List(ctx context.Context, page, limit int) ([]entity.User, error) {
	offset := (page - 1) * limit
	return u.userRepo.List(ctx, limit, offset)
}

func (u *userUsecase) Update(ctx context.Context, id uuid.UUID, req entity.UserUpdate) (*entity.User, error) {
	return u.userRepo.Update(ctx, id, req)
}

func (u *userUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.userRepo.Delete(ctx, id)
}
