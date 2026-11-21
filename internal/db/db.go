package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"mybili/internal/dbtime"
)

func Open(path string, sqlLogLevel string) (*gorm.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return nil, err
	}
	level := SQLLogLevel(sqlLogLevel)
	dsn := path + "?_pragma=busy_timeout(60000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(0)&_txlock=immediate"
	conn, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger:  logger.Default.LogMode(level),
		NowFunc: dbtime.Now,
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	sqlDB, err := conn.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(4)
	return conn, nil
}

// SQLLogLevel 把 SQL_LOG_LEVEL 转成 GORM 级别。空值和无法识别的值都用 warn。
func SQLLogLevel(value string) logger.LogLevel {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "warn":
		return logger.Warn
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "info":
		return logger.Info
	default:
		log.Printf("未知的 SQL_LOG_LEVEL %q，已改用 warn", value)
		return logger.Warn
	}
}
