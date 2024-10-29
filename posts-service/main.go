package main

import (
	"blog-system/posts-service/internal/models"
	"blog-system/posts-service/internal/repository/postgres"
	"blog-system/posts-service/internal/service"
	postsv1 "blog-system/proto/posts/v1"
	"net"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"blog-system/posts-service/internal/logger"
)

func main() {
	// Инициализация логгера
	logger.Init()
	log := logger.Get()
	defer log.Sync()

	// Настройка подключения к базе данных
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=blog_posts port=5432 sslmode=disable"
	}
	log.Info("connecting to database", zap.String("dsn", dsn))

	// Настройка GORM
	db, err := gorm.Open(pgdriver.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	})
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}

	// Автоматическая миграция
	if err := db.AutoMigrate(&models.Post{}); err != nil {
		log.Fatal("failed to migrate database", zap.Error(err))
	}

	// Инициализация репозитория и сервиса
	repo := postgres.NewPostRepository(db)
	postService := service.NewPostService(repo)

	// Настройка gRPC сервера
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	postsv1.RegisterPostServiceServer(grpcServer, postService)
	reflection.Register(grpcServer)

	// Запуск HTTP сервера для метрик
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":2112", nil); err != nil {
			log.Fatal("failed to start metrics server", zap.Error(err))
		}
	}()

	log.Info("starting gRPC server", zap.String("port", ":50051"))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to serve", zap.Error(err))
	}
}
