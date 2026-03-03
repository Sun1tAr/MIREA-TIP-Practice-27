package repository

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

type Task struct {
	ID          string
	Title       string
	Description string
	Done        bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, id string) (*Task, error)
	List(ctx context.Context) ([]*Task, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id string) error
	SearchByTitle(ctx context.Context, titleSubstring string) ([]*Task, error)
}

type PostgresTaskRepository struct {
	db *sql.DB
}

func NewPostgresTaskRepository(dsn string) (*PostgresTaskRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresTaskRepository{db: db}, nil
}

func (r *PostgresTaskRepository) Close() error {
	return r.db.Close()
}

func (r *PostgresTaskRepository) Create(ctx context.Context, task *Task) error {
	query := `INSERT INTO tasks (id, title, description, done, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query,
		task.ID, task.Title, task.Description, task.Done, task.CreatedAt, task.UpdatedAt)
	return err
}

func (r *PostgresTaskRepository) GetByID(ctx context.Context, id string) (*Task, error) {
	query := `SELECT id, title, description, done, created_at, updated_at FROM tasks WHERE id = $1`
	task := &Task{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.Done, &task.CreatedAt, &task.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (r *PostgresTaskRepository) List(ctx context.Context) ([]*Task, error) {
	query := `SELECT id, title, description, done, created_at, updated_at FROM tasks ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Done, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (r *PostgresTaskRepository) Update(ctx context.Context, task *Task) error {
	query := `UPDATE tasks SET title = $1, description = $2, done = $3, updated_at = NOW() WHERE id = $4`
	result, err := r.db.ExecContext(ctx, query, task.Title, task.Description, task.Done, task.ID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *PostgresTaskRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tasks WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *PostgresTaskRepository) SearchByTitle(ctx context.Context, titleSubstring string) ([]*Task, error) {
	query := `SELECT id, title, description, done, created_at, updated_at FROM tasks WHERE title ILIKE $1`
	rows, err := r.db.QueryContext(ctx, query, "%"+titleSubstring+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Done, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}