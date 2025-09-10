package handler

import (
	"context"
	"myApi/db"
	"myApi/db/entity"
	"myApi/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type TaskRepository struct {
	db *db.DbPg
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

func ListHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"list": model.Ltask})
}
func ListNoteHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"list": nil})
}
func CreateHandler(c *gin.Context) {
	var nwetask model.Task
	if err := c.ShouldBindJSON(&nwetask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"Task": model.NewTask})
	if model.NewTask.Status == "" {
		model.NewTask.Status = "pending"
	}

	// 4. Добавляем в список
	model.Ltask = append(model.Ltask, model.NewTask)
	c.JSON(http.StatusCreated, nwetask)
}
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now(),
	})
}
func SetupRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		api.GET("/health", HealthHandler)

		tasks := api.Group("/task")
		{
			tasks.GET("/list", ListHandler)
			tasks.POST("/create", CreateHandler)
		}
		notes := api.Group("/notes")
		{
			notes.GET("/list", ListNoteHandler)
		}
	}
}
