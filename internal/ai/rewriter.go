package ai

import (
	"context"

	"github.com/taufiqgit/news-api/internal/domain/entity"
)

// Rewriter mengubah artikel sumber menjadi artikel baru yang orisinal.
type Rewriter interface {
	Rewrite(ctx context.Context, req entity.RewriteRequest) (*entity.RewriteResult, error)
}
