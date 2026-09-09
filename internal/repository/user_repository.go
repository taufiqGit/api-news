package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/taufiqgit/news-api/internal/database/db"
	"github.com/taufiqgit/news-api/internal/domain/entity"
	domainrepo "github.com/taufiqgit/news-api/internal/domain/repository"
)

type userRepository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewUserRepository(pool *pgxpool.Pool) domainrepo.UserRepository {
	return &userRepository{
		pool: pool,
		q:    db.New(pool),
	}
}

func (r *userRepository) Create(ctx context.Context, u entity.User) (*entity.User, error) {
	row, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Name:      u.Name,
		Email:     u.Email,
		Password:  u.Password,
		Role:      u.Role,
		AvatarUrl: ptrToText(u.AvatarURL),
	})
	if err != nil {
		return nil, err
	}
	return toUserEntity(row), nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toUserEntity(row), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toUserEntity(row), nil
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]entity.User, error) {
	rows, err := r.q.ListUsers(ctx, db.ListUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	users := make([]entity.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, *toUserEntity(row))
	}
	return users, nil
}

func (r *userRepository) Update(ctx context.Context, id uuid.UUID, u entity.UserUpdate) (*entity.User, error) {
	row, err := r.q.UpdateUser(ctx, db.UpdateUserParams{
		ID:        id,
		Name:      ptrToText(u.Name),
		Email:     ptrToText(u.Email),
		AvatarUrl: ptrToText(u.AvatarURL),
		Role:      ptrToText(u.Role),
		IsActive:  ptrToBool(u.IsActive),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toUserEntity(row), nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteUser(ctx, id)
}

func toUserEntity(row db.User) *entity.User {
	return &entity.User{
		ID:        row.ID,
		Name:      row.Name,
		Email:     row.Email,
		Password:  row.Password,
		Role:      row.Role,
		AvatarURL: textToPtr(row.AvatarUrl),
		IsActive:  row.IsActive,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
