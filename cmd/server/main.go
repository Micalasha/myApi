package main

import (
	"context"
	"io"
	"log"
	"log/slog"
	"myApi/db"
	"myApi/handler"
	"myApi/repository/postgresql"
	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {

	fileLogger := &lumberjack.Logger{
		Filename:   "./logs/app.log", // путь к файлу
		MaxSize:    10,               // MB - максимальный размер файла
		MaxBackups: 3,                // количество старых файлов
		MaxAge:     28,               // дней - максимальный возраст
		Compress:   true,             // сжимать старые файлы
	}
	multiWriter := io.MultiWriter(os.Stdout, fileLogger)
	logger := slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
	slog.Info("server starting", "port", 8080)

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
