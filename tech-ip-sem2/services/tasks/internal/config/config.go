package config

import (
	"fmt"
	"os"
	"strconv"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	Driver   string
}

type RedisConfig struct {
	Addr          string
	Password      string
	DB            int
	TTLSeconds    int
	JitterSeconds int
}

type Config struct {
	TasksPort    string
	AuthGRPCAddr string
	LogLevel     string
	InstanceID   string
	DB           DatabaseConfig
	Redis        RedisConfig
}

func Load() (*Config, error) {
	cfg := &Config{
		TasksPort:    getEnv("TASKS_PORT", "8082"),
		AuthGRPCAddr: getEnv("AUTH_GRPC_ADDR", "localhost:50051"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		InstanceID:   getEnv("INSTANCE_ID", "unknown"),
		DB: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "tasks_user"),
			Password: getEnv("DB_PASSWORD", "tasks_pass"),
			DBName:   getEnv("DB_NAME", "tasks_db"),
			Driver:   getEnv("DB_DRIVER", "postgres"),
		},
		Redis: RedisConfig{
			Addr:          getEnv("REDIS_ADDR", "localhost:6379"),
			Password:      getEnv("REDIS_PASSWORD", ""),
			DB:            getEnvAsInt("REDIS_DB", 0),
			TTLSeconds:    getEnvAsInt("CACHE_TTL_SECONDS", 120),
			JitterSeconds: getEnvAsInt("CACHE_TTL_JITTER_SECONDS", 30),
		},
	}
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func (db *DatabaseConfig) DSN() string {
	switch db.Driver {
	case "postgres":
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			db.Host, db.Port, db.User, db.Password, db.DBName)
	case "sqlite3":
		return fmt.Sprintf("%s.db", db.DBName)
	default:
		return ""
	}
}
