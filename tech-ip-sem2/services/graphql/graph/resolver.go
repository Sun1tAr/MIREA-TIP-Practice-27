package graph

import (
	"database/sql"

	"github.com/sirupsen/logrus"
	"github.com/sun1tar/MIREA-TIP-Practice-27/tech-ip-sem2/graphql/internal/repository"
)

// Resolver - основной тип с зависимостями
type Resolver struct {
	Repo   repository.TaskRepository
	Logger *logrus.Logger
}

func NewResolver(repo repository.TaskRepository, logger *logrus.Logger) *Resolver {
	return &Resolver{
		Repo:   repo,
		Logger: logger,
	}
}

// Вспомогательные функции для обработки ошибок
func (r *Resolver) handleError(err error, msg string) error {
	if err != nil && err != sql.ErrNoRows {
		r.Logger.WithError(err).Error(msg)
	}
	return err
}
