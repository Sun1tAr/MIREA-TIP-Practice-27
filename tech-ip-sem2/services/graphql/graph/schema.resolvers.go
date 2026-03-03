package graph

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/sun1tar/MIREA-TIP-Practice-27/tech-ip-sem2/graphql/internal/repository"
)

// CreateTask is the resolver for the createTask field.
func (r *mutationResolver) CreateTask(ctx context.Context, input CreateTaskInput) (*Task, error) {
	r.Resolver.Logger.WithField("title", input.Title).Info(">>> CreateTask CALLED")

	now := time.Now()
	dbTask := &repository.Task{
		ID:          "t_" + uuid.New().String(),
		Title:       input.Title,
		Description: "",
		Done:        false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if input.Description != nil {
		dbTask.Description = *input.Description
	}

	if err := r.Resolver.Repo.Create(ctx, dbTask); err != nil {
		r.Resolver.Logger.WithError(err).Error("CreateTask failed")
		return nil, err
	}

	task := &Task{
		ID:          dbTask.ID,
		Title:       dbTask.Title,
		Description: &dbTask.Description,
		Done:        dbTask.Done,
	}

	r.Resolver.Logger.WithField("task_id", task.ID).Info("<<< CreateTask success")
	return task, nil
}

// UpdateTask is the resolver for the updateTask field.
func (r *mutationResolver) UpdateTask(ctx context.Context, id string, input UpdateTaskInput) (*Task, error) {
	r.Resolver.Logger.WithField("task_id", id).Info(">>> UpdateTask CALLED")

	dbTask, err := r.Resolver.Repo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			r.Resolver.Logger.WithField("task_id", id).Warn("task not found")
			return nil, nil
		}
		return nil, err
	}
	if dbTask == nil {
		return nil, nil
	}

	if input.Title != nil {
		dbTask.Title = *input.Title
	}
	if input.Description != nil {
		dbTask.Description = *input.Description
	}
	if input.Done != nil {
		dbTask.Done = *input.Done
	}
	dbTask.UpdatedAt = time.Now()

	if err := r.Resolver.Repo.Update(ctx, dbTask); err != nil {
		return nil, err
	}

	task := &Task{
		ID:          dbTask.ID,
		Title:       dbTask.Title,
		Description: &dbTask.Description,
		Done:        dbTask.Done,
	}

	r.Resolver.Logger.WithField("task_id", id).Info("<<< UpdateTask success")
	return task, nil
}

// DeleteTask is the resolver for the deleteTask field.
func (r *mutationResolver) DeleteTask(ctx context.Context, id string) (bool, error) {
	r.Resolver.Logger.WithField("task_id", id).Info(">>> DeleteTask CALLED")

	err := r.Resolver.Repo.Delete(ctx, id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	r.Resolver.Logger.WithField("task_id", id).Info("<<< DeleteTask success")
	return true, nil
}

// Tasks is the resolver for the tasks field.
func (r *queryResolver) Tasks(ctx context.Context) ([]*Task, error) {
	r.Resolver.Logger.Info(">>> Tasks CALLED")

	dbTasks, err := r.Resolver.Repo.List(ctx)
	if err != nil {
		return nil, err
	}

	tasks := make([]*Task, len(dbTasks))
	for i, dbTask := range dbTasks {
		tasks[i] = &Task{
			ID:          dbTask.ID,
			Title:       dbTask.Title,
			Description: &dbTask.Description,
			Done:        dbTask.Done,
		}
	}

	r.Resolver.Logger.WithField("count", len(tasks)).Info("<<< Tasks success")
	return tasks, nil
}

// Task is the resolver for the task field.
func (r *queryResolver) Task(ctx context.Context, id string) (*Task, error) {
	r.Resolver.Logger.WithField("task_id", id).Info(">>> Task CALLED")

	dbTask, err := r.Resolver.Repo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if dbTask == nil {
		return nil, nil
	}

	task := &Task{
		ID:          dbTask.ID,
		Title:       dbTask.Title,
		Description: &dbTask.Description,
		Done:        dbTask.Done,
	}

	r.Resolver.Logger.WithField("task_id", id).Info("<<< Task success")
	return task, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }