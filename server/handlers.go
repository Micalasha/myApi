package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func ListHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"list": Ltask})
}
func CreateHandler(c *gin.Context) {
	var nwetask Task
	if err := c.ShouldBindJSON(&nwetask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"Task": NewTask})
	if NewTask.Status == "" {
		NewTask.Status = "pending"
	}

	// 4. Добавляем в список
	Ltask = append(Ltask, NewTask)
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
	}
}
