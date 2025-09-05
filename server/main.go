package main

import (
	"log"
	"myApi/db"

	"github.com/gin-gonic/gin"
)

func main() {

	dsn, err := db.Confini()
	if err != nil {
		log.Printf("Error parse dsn: %v", err)
	}
	conn, err := db.Connect(dsn)
	if err != nil {
		log.Printf("Error connect to db: %v", err)
	}
	defer conn.Close()
	dbrep := db.NewDatabaseRepository(conn)
	_, err = dbrep.ExecQuery("select * from mm.people limit 10")
	if err != nil {
		log.Printf("Error execute query: %v", err)
	}
	router := gin.Default()
	SetupRoutes(router)
	router.Run(":8080")

}
