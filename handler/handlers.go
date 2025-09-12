package handler

import (
	"myApi/model"
	"myApi/repository/postgresql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	taskRepo *postgresql.TaskRepository // Храним репозиторий
}

func NewHandler(taskRepo *postgresql.TaskRepository) *Handler {
	return &Handler{
		taskRepo: taskRepo,
	}
}

func (h *Handler) ListHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"list": model.Ltask})
}
func (h *Handler) ListNoteHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"list": nil})
}

/*
	func (h *Handler) CreateTaskHandler(c *gin.Context) {
		var newtask dto.CreateTaskRequest
		if err := c.ShouldBindJSON(&newtask); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return // Забыл return!
		}

		taskModel := dto.ToTaskModel(newtask)

		if taskModel.Priority == 0 {
			taskModel.Priority = 3
		}

		// Теперь правильно используем репозиторий через h.taskRepo
		createdTask, err := h.taskRepo.CreateTask(c.Request.Context(), *taskModel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
			return
		}

		c.JSON(http.StatusCreated, dto.ToTaskResponse(createdTask))
	}
*/
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
			tasks.GET("/list", h.ListHandler)
			/*tasks.POST("/create", h.CreateTaskHandler)*/
		}
		notes := api.Group("/notes")
		{
			notes.GET("/list", h.ListNoteHandler)
		}
	}
}

func authMiddlewareGroup(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Token") != token {
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
