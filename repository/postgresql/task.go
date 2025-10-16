package postgresql

import (
	"context"
	"errors"
	"fmt"
	"myApi/db"
	"myApi/db/entity"
	"myApi/model"
	"time"

	"github.com/jackc/pgx/v5"
)

type TaskRepository struct {
	db *db.DbPg // Использует подключение из db.go
}

func NewTaskRepository(db *db.DbPg) *TaskRepository {
	return &TaskRepository{db: db}
}
func (t *TaskRepository) GetAllTasks() ([]entity.TaskEntity, error) {
	query := "SELECT * FROM tasks"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*90)
	defer cancel()
	rows, err := t.db.Query(ctx, query)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("база данных не ответила за 90 секунд: %w", err)
		}
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer rows.Close()
	tasks, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.TaskEntity])
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
func (t *TaskRepository) CreateTask(ctx context.Context, task model.Task) (entity.TaskEntity, error) {
	query := `
        INSERT INTO tasks (title, description, status, priority)
        VALUES (?, ?, ?, ?)
        returning id, title, description, status, priority, created_at, updated_at`
	var Tentity entity.TaskEntity
	err := t.db.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.Status).Scan(&Tentity.ID, &Tentity.Title, &Tentity.Description, &Tentity.Status)
	if err != nil {
		return entity.TaskEntity{}, err
	}

	return Tentity, nil
}
