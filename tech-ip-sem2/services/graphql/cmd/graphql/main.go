package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/sun1tar/MIREA-TIP-Practice-27/tech-ip-sem2/graphql/graph"
	"github.com/sun1tar/MIREA-TIP-Practice-27/tech-ip-sem2/graphql/internal/config"
	"github.com/sun1tar/MIREA-TIP-Practice-27/tech-ip-sem2/graphql/internal/logger"
	"github.com/sun1tar/MIREA-TIP-Practice-27/tech-ip-sem2/graphql/internal/repository"
)

func main() {
	logrusLogger := logger.Init("graphql")

	cfg, err := config.Load()
	if err != nil {
		logrusLogger.WithError(err).Fatal("failed to load config")
	}

	logrusLogger.WithFields(map[string]interface{}{
		"port":    cfg.GraphQLPort,
		"db_host": cfg.DB.Host,
		"db_name": cfg.DB.DBName,
	}).Info("starting GraphQL service")

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

	// Проверяем подключение к БД
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := repo.List(ctx); err != nil {
		logrusLogger.WithError(err).Warn("database not responding, but continuing")
	} else {
		logrusLogger.Info("database connected successfully")
	}

	// Создаём резолвер с репозиторием
	resolver := graph.NewResolver(repo, logrusLogger)

	// Настраиваем GraphQL сервер
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	// Настраиваем маршруты
	http.Handle("/", playground.Handler("GraphQL Playground", "/query"))
	http.Handle("/query", srv)

	addr := fmt.Sprintf(":%s", cfg.GraphQLPort)
	server := &http.Server{
		Addr:         addr,
		Handler:      nil,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Запускаем сервер в горутине
	go func() {
		logrusLogger.WithField("port", cfg.GraphQLPort).Info("GraphQL server started")
		logrusLogger.Info("Playground available at http://localhost:" + cfg.GraphQLPort + "/")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrusLogger.WithError(err).Fatal("server failed")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrusLogger.Info("Shutting down GraphQL server...")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logrusLogger.WithError(err).Fatal("server forced to shutdown")
	}

	logrusLogger.Info("server exited")
}
