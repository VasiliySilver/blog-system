package postgres

import (
	"context"
	"errors"

	"blog-system/posts-service/internal/logger"
	"blog-system/posts-service/internal/models"

	"go.uber.org/zap"

	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(ctx context.Context, post *models.Post) error {
	log := logger.Get()

	log.Debug("creating post in database",
		zap.String("title", post.Title),
		zap.String("author_id", post.AuthorID))

	if err := r.db.WithContext(ctx).Create(post).Error; err != nil {
		log.Error("failed to create post in database",
			zap.Error(err),
			zap.String("title", post.Title))
		return err
	}

	return nil
}

func (r *PostRepository) GetByID(ctx context.Context, id string) (*models.Post, error) {
	var post models.Post
	err := r.db.WithContext(ctx).First(&post, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("post not found")
		}
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) List(ctx context.Context, offset, limit int) ([]*models.Post, int64, error) {
	var posts []*models.Post
	var total int64

	// Получаем общее количество
	if err := r.db.WithContext(ctx).Model(&models.Post{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Получаем записи с пагинацией
	err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&posts).Error

	return posts, total, err
}

func (r *PostRepository) Update(ctx context.Context, post *models.Post) error {
	result := r.db.WithContext(ctx).Save(post)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("post not found")
	}
	return nil
}

func (r *PostRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.Post{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("post not found")
	}
	return nil
}
