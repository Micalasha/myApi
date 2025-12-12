package handler

import (
	"context"
	"errors"
	"log/slog"
	"myApi/db"
	"myApi/db/entity"
	"myApi/dto"
	"myApi/model"
	"myApi/repository/postgresql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type TaskRepo interface {
	GetAllTasks(ctx context.Context) ([]entity.TaskEntity, error)
	CreateTask(ctx context.Context, task model.Task) (entity.TaskEntity, error)
	UpdateTask(ctx context.Context, task dto.UpdateTaskRequest) (entity.TaskEntity, error)
	GetTaskById(ctx context.Context, id string) (entity.TaskEntity, error)
}
type Handler struct {
	taskRepo TaskRepo
	logger   *slog.Logger
}

func NewHandler(taskRepo TaskRepo, logger *slog.Logger) *Handler {
	return &Handler{
		taskRepo: taskRepo,
		logger:   logger,
	}
}

type HealthHandler struct {
	dbPool *db.Pool
}

func NewHealthHandler(dbPool *db.Pool) *HealthHandler {
	return &HealthHandler{dbPool: dbPool}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	status := gin.H{
		"status": "ok",
		"database": gin.H{
			"connected": h.dbPool.IsHealthy(),
		},
	}

	if !h.dbPool.IsHealthy() {
		c.JSON(http.StatusServiceUnavailable, status)
		return
	}

	c.JSON(http.StatusOK, status)
}

// TaskListHandler godoc
// @Summary      Get all tasks
// @Description  Get list of all tasks
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {array}   dto.TaskResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /task/list [get]
func (h *Handler) TaskListHandler(c *gin.Context) {
	tasks, err := h.taskRepo.GetAllTasks(c.Request.Context())

	if errors.Is(err, postgresql.ErrDatabaseUnavailable) {
		h.logger.Warn("Database unavailable during task list request")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Database temporarily unavailable",
		})
		return
	}

	if err != nil {
		h.logger.Error("Failed to get tasks", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Не удалось получить список задач",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"list": tasks})
}

// CreateTaskHandler godoc
// @Summary      Create a new task
// @Description  Create a new task with the input payload
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     Authorization
// @Param        task  body      dto.CreateTaskRequest  true  "Task data"
// @Success      201   {object}  dto.TaskResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      503   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /task/create [post]
func (h *Handler) CreateTaskHandler(c *gin.Context) {
	var newtask dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&newtask); err != nil {
		h.logger.Warn("Invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskModel := dto.ToTaskModel(newtask)
	if taskModel.Priority == 0 {
		taskModel.Priority = 3
	}

	createdTask, err := h.taskRepo.CreateTask(c.Request.Context(), *taskModel)

	if errors.Is(err, postgresql.ErrDatabaseUnavailable) {
		h.logger.Warn("Database unavailable during task creation")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Database temporarily unavailable",
			"message": "Please retry your request in a few moments",
		})
		return
	}

	if err != nil {
		h.logger.Error("Failed to create task", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, dto.ToTaskResponse(createdTask.ToModel()))
}

func (h *Handler) GetTaskByIdHandler(c *gin.Context) {
	id := c.Param("id")
	task, err := h.taskRepo.GetTaskById(c.Request.Context(), id)
	if errors.Is(err, postgresql.ErrDatabaseUnavailable) {

		h.logger.Warn("Database unavailable during task get")
	}

}

/*func (h *Handler) UpdateTaskHandler(c *gin.Context) {
	var updateTask dto.UpdateTaskRequest
	if err := c.ShouldBindJSON(&updateTask); err != nil {
		h.logger.Warn("Invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err, update := h.taskRepo.UpdateTask(c.Request.Context(), updateTask)

}*/

// ListNoteHandler godoc
// @Summary      Get all notes
// @Description  Get list of all notes
// @Tags         notes
// @Accept       json
// @Produce      json
// @Security     Authorization
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /notes/list [get]
func (h *Handler) ListNoteHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"list": nil})
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
		api.Use(authMiddlewareGroup("123"))
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
