package db

import (
	"database/sql"
	"fmt"

	"udhaar-manager/config"

	_ "github.com/go-sql-driver/mysql"
)

func Connect(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(20)
	conn.SetMaxIdleConns(10)
	return conn, nil
}
