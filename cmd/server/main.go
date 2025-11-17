package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"myApi/db"
	_ "myApi/docs"
	"myApi/handler"
	"myApi/repository/postgresql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gopkg.in/natefinch/lumberjack.v2"
)

// @title           Task API
// @version         1.0
// @description     API for managing tasks and notes
// @contact.name   API Support
// @contact.email  support@example.com
// @host      localhost:8080
// @BasePath  /api
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	// 1. Настраиваем логгер
	logger := setupLogger()
	slog.SetDefault(logger) // устанавливаем как глобальный

	logger.Info("Starting application")

	// 2. Загружаем конфиг
	ctx := context.Background()
	dsn, err := db.BuildDSN()
	if err != nil {
		logger.Error("Failed to build DSN", "error", err)
		os.Exit(1)
	}

	// 3. Создаем pool с логгером
	dbPool, err := db.NewPool(ctx, dsn, logger)
	if err != nil {
		logger.Error("Failed to initialize database pool", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// 4. Создаем репозитории с логгером
	taskRepo := postgresql.NewTaskRepository(dbPool, logger)

	// 5. Создаем handlers с логгером
	h := handler.NewHandler(taskRepo, logger)
	healthHandler := handler.NewHealthHandler(dbPool)

	// 6. Настраиваем router
	gin.SetMode(gin.DebugMode)
	router := gin.New()

	// Middleware
	router.Use(ginLogger(logger))
	router.Use(gin.Recovery())

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check (без auth)
	router.GET("/health", healthHandler.HealthCheck)

	// API routes
	h.SetupRoutes(router)

	// 7. Запуск сервера
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		logger.Info("Server starting", "address", ":8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// 8. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited gracefully")
}

func setupLogger() *slog.Logger {
	// Создаем директорию
	if err := os.MkdirAll("./logs", 0755); err != nil {
		panic(fmt.Sprintf("Failed to create logs directory: %v", err))
	}

	// Rotation для файла
	logFile := &lumberjack.Logger{
		Filename:   "./logs/app.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}

	env := os.Getenv("ENV")

	if env == "production" {
		// Production: JSON в оба места
		multiWriter := io.MultiWriter(os.Stdout, logFile)
		handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: true,
		})
		return slog.New(handler)
	}

	// Development: используем tee handler (разные форматы)
	consoleHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: false,
	})

	fileHandler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})

	// Комбинируем оба handler'а
	return slog.New(&teeHandler{
		handlers: []slog.Handler{consoleHandler, fileHandler},
	})
}

// teeHandler отправляет логи в несколько handler'ов одновременно
type teeHandler struct {
	handlers []slog.Handler
}

func (t *teeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range t.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (t *teeHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, h := range t.handlers {
		if err := h.Handle(ctx, record.Clone()); err != nil {
			return err
		}
	}
	return nil
}

func (t *teeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(t.handlers))
	for i, h := range t.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return &teeHandler{handlers: handlers}
}

func (t *teeHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(t.handlers))
	for i, h := range t.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return &teeHandler{handlers: handlers}
}

// Middleware для логирования HTTP запросов
func ginLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)

		logger.Info("HTTP Request",
			"status", c.Writer.Status(),
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"ip", c.ClientIP(),
			"latency_ms", latency.Milliseconds(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}
