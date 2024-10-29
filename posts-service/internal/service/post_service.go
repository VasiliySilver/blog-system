package service

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"blog-system/posts-service/internal/logger"
	"blog-system/posts-service/internal/metrics"
	"blog-system/posts-service/internal/models"
	"blog-system/posts-service/internal/repository/postgres"

	"go.uber.org/zap"

	postsv1 "blog-system/proto/posts/v1"

	"github.com/prometheus/client_golang/prometheus"
)

type PostService struct {
	postsv1.UnimplementedPostServiceServer
	repo *postgres.PostRepository
}

func NewPostService(repo *postgres.PostRepository) *PostService {
	return &PostService{repo: repo}
}

func (s *PostService) CreatePost(ctx context.Context, req *postsv1.CreatePostRequest) (*postsv1.CreatePostResponse, error) {
	timer := prometheus.NewTimer(metrics.RequestDuration.WithLabelValues("CreatePost"))
	defer timer.ObserveDuration()

	log := logger.Get()

	log.Info("creating post",
		zap.String("title", req.Title),
		zap.String("author_id", req.AuthorId))

	// Валидация входных данных
	if req.Title == "" {
		log.Warn("empty title in create post request")
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "content is required")
	}
	if req.AuthorId == "" {
		return nil, status.Error(codes.InvalidArgument, "author_id is required")
	}

	post := &models.Post{
		Title:     req.Title,
		Content:   req.Content,
		AuthorID:  req.AuthorId,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, post); err != nil {
		metrics.RequestsTotal.WithLabelValues("CreatePost", "error").Inc()
		metrics.DatabaseErrors.Inc()
		log.Error("failed to create post",
			zap.Error(err),
			zap.String("title", req.Title),
			zap.String("author_id", req.AuthorId))
		return nil, status.Error(codes.Internal, "failed to create post: "+err.Error())
	}

	metrics.RequestsTotal.WithLabelValues("CreatePost", "success").Inc()
	metrics.PostsTotal.Inc()

	log.Info("post created successfully",
		zap.String("id", post.ID),
		zap.String("title", post.Title))

	return &postsv1.CreatePostResponse{
		Post: &postsv1.Post{
			Id:        post.ID,
			Title:     post.Title,
			Content:   post.Content,
			AuthorId:  post.AuthorID,
			CreatedAt: post.CreatedAt.Format(time.RFC3339),
			UpdatedAt: post.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (s *PostService) GetPost(ctx context.Context, req *postsv1.GetPostRequest) (*postsv1.GetPostResponse, error) {
	post, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "post not found: %v", err)
	}

	return &postsv1.GetPostResponse{
		Post: convertPostToProto(post),
	}, nil
}

func (s *PostService) ListPosts(ctx context.Context, req *postsv1.ListPostsRequest) (*postsv1.ListPostsResponse, error) {
	timer := prometheus.NewTimer(metrics.RequestDuration.WithLabelValues("ListPosts"))
	defer timer.ObserveDuration()

	offset := int(req.Page * req.PageSize)
	posts, total, err := s.repo.List(ctx, offset, int(req.PageSize))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list posts: %v", err)
	}

	postProtos := make([]*postsv1.Post, len(posts))
	for i, post := range posts {
		postProtos[i] = convertPostToProto(post)
	}

	metrics.RequestsTotal.WithLabelValues("ListPosts", "success").Inc()
	return &postsv1.ListPostsResponse{
		Posts: postProtos,
		Total: int32(total),
	}, nil
}

func (s *PostService) UpdatePost(ctx context.Context, req *postsv1.UpdatePostRequest) (*postsv1.UpdatePostResponse, error) {
	post, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "post not found: %v", err)
	}

	if req.Title != "" {
		post.Title = req.Title
	}
	if req.Content != "" {
		post.Content = req.Content
	}
	post.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, post); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update post: %v", err)
	}

	return &postsv1.UpdatePostResponse{
		Post: convertPostToProto(post),
	}, nil
}

func (s *PostService) DeletePost(ctx context.Context, req *postsv1.DeletePostRequest) (*postsv1.DeletePostResponse, error) {
	if err := s.repo.Delete(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete post: %v", err)
	}

	return &postsv1.DeletePostResponse{
		Success: true,
	}, nil
}

// Вспомогательная функция для конвертации модели в proto
func convertPostToProto(post *models.Post) *postsv1.Post {
	return &postsv1.Post{
		Id:        post.ID,
		Title:     post.Title,
		Content:   post.Content,
		AuthorId:  post.AuthorID,
		CreatedAt: post.CreatedAt.Format(time.RFC3339),
		UpdatedAt: post.UpdatedAt.Format(time.RFC3339),
	}
}
