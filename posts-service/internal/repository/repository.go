package repository

import (
	"blog-system/posts-service/internal/models"
	"context"
)

type PostRepository interface {
	Create(ctx context.Context, post *models.Post) error
	GetByID(ctx context.Context, id string) (*models.Post, error)
	List(ctx context.Context, offset, limit int) ([]*models.Post, int, error)
	Update(ctx context.Context, post *models.Post) error
	Delete(ctx context.Context, id string) error
}
