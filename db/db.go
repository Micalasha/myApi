package db

import (
	"context"
	"fmt"
	"log/slog"
	"myApi/config"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/ini.v1"
)

type Pool struct {
	pool        *pgxpool.Pool
	databaseURL string
	config      *pgxpool.Config
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	logger      *slog.Logger // ← добавили
}

func NewPool(ctx context.Context, dsn string, logger *slog.Logger) (*Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		logger.Warn("Initial database connection failed", "error", err)
		logger.Info("Service will start and attempt to reconnect in background")
	} else {
		if err = pool.Ping(ctx); err != nil {
			logger.Warn("Database ping failed", "error", err)
			pool.Close()
			pool = nil
		} else {
			logger.Info("Successfully connected to database")
		}
	}

	bgCtx, cancel := context.WithCancel(context.Background())

	p := &Pool{
		pool:        pool,
		databaseURL: dsn,
		config:      config,
		ctx:         bgCtx,
		cancel:      cancel,
		logger:      logger, // ← сохраняем
	}

	go p.healthCheck()

	return p, nil
}

func (p *Pool) healthCheck() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.mu.RLock()
			pool := p.pool
			p.mu.RUnlock()

			if pool == nil {
				p.logger.Warn("Database pool is nil, attempting to reconnect")
				p.reconnect()
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := pool.Ping(ctx)
			cancel()

			if err != nil {
				p.logger.Warn("Database health check failed", "error", err)
				p.reconnect()
			}
		}
	}
}

func (p *Pool) reconnect() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pool != nil {
		p.pool.Close()
		p.pool = nil
	}

	for i := 0; i < 5; i++ {
		p.logger.Info("Reconnection attempt", "attempt", i+1, "max", 5)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		pool, err := pgxpool.NewWithConfig(ctx, p.config)
		cancel()

		if err == nil {
			pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
			err = pool.Ping(pingCtx)
			pingCancel()

			if err == nil {
				p.pool = pool
				p.logger.Info("Successfully reconnected to database")
				return
			}
			pool.Close()
		}

		p.logger.Warn("Reconnection attempt failed", "attempt", i+1, "error", err)

		if i < 4 {
			waitTime := time.Duration(i+1) * 2 * time.Second
			time.Sleep(waitTime)
		}
	}

	p.logger.Error("Failed to reconnect after 5 attempts")
}

func (p *Pool) GetPool() *pgxpool.Pool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.pool
}

func (p *Pool) IsHealthy() bool {
	p.mu.RLock()
	pool := p.pool
	p.mu.RUnlock()

	if pool == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return pool.Ping(ctx) == nil
}

func (p *Pool) Close() {
	p.logger.Info("Closing database pool")
	p.cancel()

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pool != nil {
		p.pool.Close()
		p.pool = nil
	}
}

// Остальные функции без изменений
func BuildDSN() (dsn string, err error) {
	pgCfg, err := LoadConfig("./kis.ini")
	if err != nil {
		return "", fmt.Errorf("load config: %w", err)
	}
	return ParseConf(pgCfg), nil
}

func LoadConfig(path string) (*config.DatabaseConfig, error) {
	cfgFile, err := ini.Load(path)
	if err != nil {
		return nil, err
	}
	sec := cfgFile.Section("Postgresql")

	pg := &config.DatabaseConfig{
		Server:   sec.Key("server").MustString("localhost"),
		Port:     sec.Key("port").MustInt(5432),
		Database: sec.Key("database").MustString("mydatabase"),
		Username: sec.Key("user").MustString("myuser"),
		Password: sec.Key("password").MustString("1104"),
	}

	return pg, nil
}

func ParseConf(dc *config.DatabaseConfig) (dsn string) {
	dsn = fmt.Sprintf(
		"user=%s password=%s host=%s port=%d dbname=%s sslmode=disable",
		dc.Username,
		dc.Password,
		dc.Server,
		dc.Port,
		dc.Database,
	)
	return dsn
}

func (p *Pool) ExecQuery(query string, args ...any) ([]map[string]any, error) {
	rows, err := p.pool.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToMap)
}
