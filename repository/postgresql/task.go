package postgresql

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"myApi/db"
	"myApi/db/entity"
	"myApi/model"
)

var ErrDatabaseUnavailable = errors.New("database connection not available")

type TaskRepository struct {
	dbPool *db.Pool
	logger *slog.Logger
}

func NewTaskRepository(dbPool *db.Pool, logger *slog.Logger) *TaskRepository {
	return &TaskRepository{
		dbPool: dbPool,
		logger: logger,
	}
}

func (t *TaskRepository) GetAllTasks(ctx context.Context) ([]entity.TaskEntity, error) {
	pool := t.dbPool.GetPool()
	if pool == nil {
		t.logger.Warn("Attempted to get tasks but database is unavailable")
		return nil, ErrDatabaseUnavailable
	}

	query := `
		SELECT id, title, description, status, priority, created_at, updated_at
		FROM tasks
		ORDER BY created_at DESC
	`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		t.logger.Error("Failed to query tasks", "error", err)
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}
	defer rows.Close()

	var tasks []entity.TaskEntity
	for rows.Next() {
		var task entity.TaskEntity
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Priority,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			t.logger.Error("Failed to scan task", "error", err)
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	t.logger.Info("Retrieved tasks", "count", len(tasks))
	return tasks, rows.Err()
}

func (t *TaskRepository) CreateTask(ctx context.Context, task model.Task) (entity.TaskEntity, error) {
	pool := t.dbPool.GetPool()
	if pool == nil {
		t.logger.Warn("Attempted to create task but database is unavailable",
			"title", task.Title,
		)
		return entity.TaskEntity{}, ErrDatabaseUnavailable
	}

	query := `
		INSERT INTO tasks (title, description, status, priority)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, description, status, priority, created_at, updated_at
	`

	var taskEntity entity.TaskEntity
	err := pool.QueryRow(ctx, query,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
	).Scan(
		&taskEntity.ID,
		&taskEntity.Title,
		&taskEntity.Description,
		&taskEntity.Status,
		&taskEntity.Priority,
		&taskEntity.CreatedAt,
		&taskEntity.UpdatedAt,
	)

	if err != nil {
		t.logger.Error("Failed to create task",
			"error", err,
			"title", task.Title,
		)
		return entity.TaskEntity{}, fmt.Errorf("failed to create task: %w", err)
	}

	t.logger.Info("Task created successfully",
		"task_id", taskEntity.ID,
		"title", taskEntity.Title,
	)

	return taskEntity, nil
}
