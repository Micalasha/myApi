package dto

import (
	"myApi/model"
	"time"
)

type TaskResponse struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    int       `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func ToTaskResponse(task *model.Task) TaskResponse {
	return TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		Priority:    task.Priority,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

type CreateTaskRequest struct {
	Title       string `json:"title" binding:"required,min=2,max=200"`
	Description string `json:"description,omitempty"`
	Priority    int    `json:"priority,omitempty"`
}
type UpdateTaskRequest struct {
	ID          int    `json:"id" binding:"required"`
	Title       string `json:"title" binding:"required,max=200"`
	Description string `json:"description,omitempty"`
	Priority    int    `json:"priority,omitempty"`
}

func ToTaskModel(req CreateTaskRequest) *model.Task {
	return &model.Task{
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		Status:      model.StatusPending,
	}
}
