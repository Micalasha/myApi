package main

import (
	"context"
	"log"
	"myApi/db"
	"myApi/handler"

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
	if err != nil {
		log.Printf("Error execute query: %v", err)
	}
	router := gin.Default()
	handler.SetupRoutes(router)

	router.Run(":8080")

}
