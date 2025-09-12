package postgresql

import (
	"context"
	"myApi/db"
	"myApi/db/entity"

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
	rows, err := t.db.Query(context.Background(), query)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	tasks, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.TaskEntity])
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
