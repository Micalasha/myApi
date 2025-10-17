package handler

import (
	"log/slog"
	"myApi/dto"
	"myApi/repository/postgresql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	taskRepo *postgresql.TaskRepository // Храним репозиторий
}

func NewHandler(taskRepo *postgresql.TaskRepository) *Handler {
	return &Handler{
		taskRepo: taskRepo,
	}
}

func (h *Handler) TaskListHandler(c *gin.Context) {
	tasks, err := h.taskRepo.GetAllTasks()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"handler": "TaskListHandler",
			"error":   err.Error(),
		}).Error("Failed to get tasks")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Не удалось получить список задач",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"list": tasks})
}
func (h *Handler) ListNoteHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"list": nil})
}

func (h *Handler) CreateTaskHandler(c *gin.Context) {
	var newtask dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&newtask); err != nil {
		slog.Warn("invalid request body",
			"error", err,
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskModel := dto.ToTaskModel(newtask)

	if taskModel.Priority == 0 {
		taskModel.Priority = 3
	}

	createdTask, err := h.taskRepo.CreateTask(c.Request.Context(), *taskModel)
	if err != nil {
		slog.Error("failed to create task",
			"error", err,
			"task_title", taskModel.Title,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}
	slog.Info("task created successfully",
		"task_id", createdTask.ID,
		"title", createdTask.Title,
	)
	c.JSON(http.StatusCreated, dto.ToTaskResponse(createdTask.ToModel()))
}

func (h *Handler) HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now(),
	})
}
func (h *Handler) SetupRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		api.Use(authMiddlewareGroup("123214"))
		api.GET("/health", h.HealthHandler)
		tasks := api.Group("/task")
		{
			tasks.GET("/list", h.TaskListHandler)
			tasks.POST("/create", h.CreateTaskHandler)
		}
		notes := api.Group("/notes")
		{
			notes.GET("/list", h.ListNoteHandler)
		}
	}
}

func authMiddlewareGroup(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") != token {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		c.Next()
	}
}

func authMiddleware(token string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Token") != token {
				http.Error(w, "Forbidden", 403)
				return
			}
			next(w, r)
		}
	}
}
