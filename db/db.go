package db

import (
	"context"
	"fmt"
	"myApi/model"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"gopkg.in/ini.v1"
)

type DbPg struct {
	*pgxpool.Pool
}

func NewConnection(ctx context.Context, dsn string) (*DbPg, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	config.MaxConns = 10               // максимум соединений
	config.MinConns = 2                // минимум соединений
	config.MaxConnLifetime = time.Hour // время жизни соединения
	config.MaxConnIdleTime = time.Minute * 30
	config.HealthCheckPeriod = time.Minute
	dbpool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}
	if err := dbpool.Ping(ctx); err != nil {
		dbpool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &DbPg{dbpool}, err
}

func Confini() (dsn string, err error) {
	pgCfg, err := LoadConfig("./kis.ini")
	if err != nil {
		return "", fmt.Errorf("load config: %w", err)
	}
	return ParseConf(pgCfg), nil
}
func LoadConfig(path string) (*model.DatabaseConfig, error) {
	cfgFile, err := ini.Load(path)
	if err != nil {
		return nil, err
	}
	sec := cfgFile.Section("Postgresql")

	pg := &model.DatabaseConfig{
		Server:   sec.Key("server").MustString("localhost"),
		Port:     sec.Key("port").MustInt(5432),
		Database: sec.Key("database").MustString(""),
		Username: sec.Key("user").MustString(""),
		Password: sec.Key("password").MustString(""),
	}

	return pg, nil
}
func ParseConf(dc *model.DatabaseConfig) (dsn string) {
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
func (r *DbPg) ExecQuery(query string, args ...any) ([]map[string]any, error) {
	rows, err := r.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToMap)
}
func (db *DbPg) Close() {
	db.Close()
}
