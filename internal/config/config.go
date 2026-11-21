package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"

	"mybili/internal/version"
)

type Config struct {
	HTTPAddr              string
	Timezone              string
	Version               string
	Debug                 bool
	SQLLogLevel           string // GORM SQL 日志级别：silent、error、warn、info
	AutoMigrate           bool
	DBPath                string
	DataDir               string
	HumanReadableDir      string
	RedisHost             string
	RedisPort             string
	RedisPassword         string
	RedisDB               int
	FavPageSize           int
	DownloadConcurrency   int
	LimitDownloadVideoJob int
	IgnoreCookies         bool
	SettingReadOnly       bool
}

func Load() Config {
	_ = godotenv.Load()
	return Config{
		HTTPAddr:              env("HTTP_ADDR", ":8080"),
		Timezone:              env("APP_TIMEZONE", "Asia/Shanghai"),
		Version:               version.Version,
		Debug:                 envBool("APP_DEBUG", false),
		SQLLogLevel:           env("SQL_LOG_LEVEL", "warn"),
		AutoMigrate:           envBool("DB_AUTO_MIGRATE", false),
		DBPath:                env("DB_DATABASE", "database/database.sqlite"),
		DataDir:               env("DATA_DIR", "storage/app"),
		HumanReadableDir:      os.Getenv("HUMAN_READABLE_DIR"),
		RedisHost:             env("REDIS_HOST", "127.0.0.1"),
		RedisPort:             env("REDIS_PORT", "6379"),
		RedisPassword:         os.Getenv("REDIS_PASSWORD"),
		RedisDB:               envInt("REDIS_DB", 0),
		FavPageSize:           envInt("BILIBILI_FAV_VIDEOS_PAGE_SIZE", 40),
		DownloadConcurrency:   envInt("BILIBILI_DOWNLOAD_CONCURRENCY", 3),
		LimitDownloadVideoJob: envInt("BILIBILI_LIMIT_DOWNLOAD_VIDEO_JOB", 2),
		IgnoreCookies:         envBool("BILIBILI_IGNORE_COOKIES", false),
		SettingReadOnly:       envBool("BILIBILI_SETTING_READ_ONLY", false),
	}
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "on", "yes":
		return true
	case "0", "false", "off", "no":
		return false
	default:
		return fallback
	}
}
