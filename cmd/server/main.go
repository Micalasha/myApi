package main

import (
	"context"
	"log"
	"myApi/db"
	"myApi/handler"
	"myApi/repository/postgresql"

	"github.com/gin-gonic/gin"
)

func main() {

	dsn, err := db.Confini()
	if err != nil {
		log.Printf("Error parse dsn: %v", err)
	}

	pool, err := db.NewConnection(context.Background(), dsn)
	if err != nil {
		log.Printf("Error connect to postgres: %v", err)
	}
	defer pool.Close()

	taskRepo := postgresql.NewTaskRepository(pool)

	h := handler.NewHandler(taskRepo)

	router := gin.Default()
	h.SetupRoutes(router)

	router.Run(":8080")

}
