package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sun1tar/MIREA-TIP-Practice-27/tech-ip-sem2/shared/logger"
	"github.com/sun1tar/MIREA-TIP-Practice-27/tech-ip-sem2/shared/middleware"
	"github.com/sun1tar/MIREA-TIP-Practice-27/tech-ip-sem2/tasks/internal/cache"
	"github.com/sun1tar/MIREA-TIP-Practice-27/tech-ip-sem2/tasks/internal/client/authclient"
	"github.com/sun1tar/MIREA-TIP-Practice-27/tech-ip-sem2/tasks/internal/config"
	handlers "github.com/sun1tar/MIREA-TIP-Practice-27/tech-ip-sem2/tasks/internal/http"
	customMiddleware "github.com/sun1tar/MIREA-TIP-Practice-27/tech-ip-sem2/tasks/internal/middleware"
	metricsMiddleware "github.com/sun1tar/MIREA-TIP-Practice-27/tech-ip-sem2/tasks/internal/middleware"
	"github.com/sun1tar/MIREA-TIP-Practice-27/tech-ip-sem2/tasks/internal/repository"
	"github.com/sun1tar/MIREA-TIP-Practice-27/tech-ip-sem2/tasks/internal/service"
)

func main() {
	logrusLogger := logger.Init("tasks")

	cfg, err := config.Load()
	if err != nil {
		logrusLogger.WithError(err).Fatal("failed to load config")
	}

	// Добавляем instance_id в логи
	logrusLogger = logrusLogger.WithField("instance_id", cfg.InstanceID).Logger

	logrusLogger.WithFields(map[string]interface{}{
		"port":        cfg.TasksPort,
		"instance_id": cfg.InstanceID,
		"auth_addr":   cfg.AuthGRPCAddr,
		"redis_addr":  cfg.Redis.Addr,
		"db_host":     cfg.DB.Host,
	}).Info("starting tasks service")

	// Инициализация репозитория (БД)
	var repo repository.TaskRepository
	if cfg.DB.Driver == "postgres" {
		postgresRepo, err := repository.NewPostgresTaskRepository(cfg.DB.DSN())
		if err != nil {
			logrusLogger.WithError(err).Fatal("failed to connect to database")
		}
		defer postgresRepo.Close()
		repo = postgresRepo
	} else {
		logrusLogger.Fatal("unsupported database driver: " + cfg.DB.Driver)
	}

	// Инициализация клиента Auth (gRPC)
	authClient, err := authclient.NewClient(cfg.AuthGRPCAddr, 2*time.Second, logrusLogger)
	if err != nil {
		logrusLogger.WithError(err).Fatal("failed to create auth client")
	}
	defer authClient.Close()

	// Инициализация клиента Redis (кэш)
	cacheClient := cache.NewClient(
		cfg.Redis.Addr,
		cfg.Redis.Password,
		cfg.Redis.DB,
		cfg.Redis.TTLSeconds,
		cfg.Redis.JitterSeconds,
		logrusLogger,
	)
	defer cacheClient.Close()

	// Проверяем подключение к Redis
	if err := cacheClient.Ping(context.Background()); err != nil {
		logrusLogger.WithError(err).Warn("redis not available, continuing without cache")
	} else {
		logrusLogger.Info("redis connected successfully")
	}

	// Инициализация сервиса с кэшем
	taskService := service.NewTaskService(repo, cacheClient, logrusLogger)

	// Инициализация хендлера
	taskHandler := handlers.NewTaskHandler(taskService, authClient, logrusLogger)

	// Настройка роутера
	mux := http.NewServeMux()

	// Health endpoint (не требует аутентификации)
	mux.HandleFunc("GET /health", taskHandler.HealthHandler)

	// Основные endpoints
	mux.HandleFunc("POST /v1/tasks", taskHandler.CreateTask)
	mux.HandleFunc("GET /v1/tasks", taskHandler.ListTasks)
	mux.HandleFunc("GET /v1/tasks/{id}", taskHandler.GetTask)
	mux.HandleFunc("PATCH /v1/tasks/{id}", taskHandler.UpdateTask)
	mux.HandleFunc("DELETE /v1/tasks/{id}", taskHandler.DeleteTask)
	mux.HandleFunc("GET /v1/tasks/search", taskHandler.SearchTasks)
	mux.Handle("GET /metrics", metricsMiddleware.MetricsHandler())

	// Цепочка middleware
	handler := middleware.RequestIDMiddleware(mux)
	handler = customMiddleware.SecurityHeadersMiddleware(handler)
	handler = customMiddleware.CSRFMiddleware(handler)
	handler = metricsMiddleware.MetricsMiddleware(handler)
	handler = middleware.LoggingMiddleware(handler)

	addr := fmt.Sprintf(":%s", cfg.TasksPort)
	logrusLogger.WithField("port", cfg.TasksPort).Info("tasks service started")
	if err := http.ListenAndServe(addr, handler); err != nil {
		logrusLogger.WithError(err).Fatal("server failed")
	}
}
