package main

import (
	"context"
	"log"
	"myApi/db"

	"github.com/gin-gonic/gin"
)

func main() {

	dsn, err := db.Confini()
	if err != nil {
		log.Printf("Error parse dsn: %v", err)
	}
	pool, err := db.NewConnection(context.Background(), dsn)
	if err != nil {
		log.Printf("Error connect to db: %v", err)
	}
	defer pool.Close()
	if err != nil {
		log.Printf("Error execute query: %v", err)
	}
	pool.ExecQuery("select * from people limit 1")
	router := gin.Default()
	SetupRoutes(router)

	router.Run(":8080")

}
