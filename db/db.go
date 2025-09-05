package db

import (
	"database/sql"
	"fmt"
	"myApi/model"
	"time"

	_ "github.com/lib/pq"
	"gopkg.in/ini.v1"
)

type DatabaseRepository interface {
	ExecQuery(query string, args ...any) ([]map[string]any, error)
}
type DatabaseRepo struct {
	db *sql.DB
}

func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}
	db.SetMaxOpenConns(20)                 // Исходя из max_connections БД
	db.SetMaxIdleConns(5)                  // Не слишком много idle
	db.SetConnMaxLifetime(2 * time.Minute) // Обновляем соединения
	db.SetConnMaxIdleTime(15 * time.Second)
	return db, err
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
func (r *DatabaseRepo) ExecQuery(query string, args ...any) ([]map[string]any, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]any

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]any)
		for i, col := range columns {
			rowMap[col] = values[i]
		}

		results = append(results, rowMap)
	}

	return results, nil
}
func NewDatabaseRepository(db *sql.DB) DatabaseRepository {
	return &DatabaseRepo{db: db}
}
