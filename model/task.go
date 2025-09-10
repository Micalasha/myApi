package model

import "time"

type Task struct {
	ID          int
	Title       string
	Description string
	Status      TaskStatus
	Priority    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

const (
	StatusPending    TaskStatus = "pending"
	StatusInProgress TaskStatus = "in_progress"
	StatusCompleted  TaskStatus = "completed"
)

type TaskStatus string

var Ltask = []Task{
	{
		ID:          1,
		Title:       "Изучить горутины в Go",
		Description: "Прочитать главу про concurrency, написать пример с воркер пулом",
		Status:      "in_progress",
	},
	{
		ID:          2,
		Title:       "Настроить Docker",
		Description: "Создать Dockerfile для проекта и docker-compose с PostgreSQL",
		Status:      "pending",
	},
	{
		ID:          3,
		Title:       "Написать тесты",
		Description: "Покрыть тестами handlers и основную бизнес-логику",
		Status:      "pending",
	},
	{
		ID:          4,
		Title:       "Сходить в магазин",
		Description: "Купить молоко, хлеб, кофе",
		Status:      "completed",
	},
	{
		ID:          5,
		Title:       "Код-ревью PR коллеги",
		Description: "Посмотреть pull request #142 по новой фиче авторизации",
		Status:      "in_progress",
	},
	{
		ID:          6,
		Title:       "Исправить баг с CORS",
		Description: "Фронтенд не может достучаться до API с localhost:3000",
		Status:      "completed",
	},
	{
		ID:          7,
		Title:       "Обновить README",
		Description: "Добавить инструкцию по запуску и описание API endpoints",
		Status:      "pending",
	},
}
var NewTask = Task{}
