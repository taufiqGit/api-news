package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	S3        S3Config
	Scheduler SchedulerConfig
	AI        AIConfig
	News      NewsConfig
	AppURL    string
}

type ServerConfig struct {
	Port         string
	Environment  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	MaxConns int32
	MinConns int32
}

type JWTConfig struct {
	Secret    string
	ExpiresIn time.Duration
}

type S3Config struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
	PublicURL string // URL publik untuk akses file (opsional)
}

// SchedulerConfig mengontrol runtime in-process scheduler berita.
type SchedulerConfig struct {
	Enabled         bool
	IntervalMinutes int
	MaxConcurrency  int
	DefaultStatus   string
}

// AIConfig adalah konfigurasi provider AI [OI]-compatible.
type AIConfig struct {
	Provider    string
	APIKey      string
	BaseURL     string
	Model       string
	MaxTokens   int
	Temperature float64
}

// NewsConfig mengontrol sumber & jumlah berita per eksekusi.
type NewsConfig struct {
	ItemsPerWebsite int
}

// Load membaca konfigurasi dari .env dan environment variable
func Load() (*Config, error) {
	// .env optional — env variable sistem lebih diutamakan
	_ = godotenv.Load()

	cfg := &Config{
		AppURL: getEnv("APP_URL", "http://localhost:8080"),
		Server: ServerConfig{
			Port:         getEnv("APP_PORT", "8080"),
			Environment:  getEnv("APP_ENV", "development"),
			ReadTimeout:  time.Duration(getEnvInt("APP_READ_TIMEOUT", 10)) * time.Second,
			WriteTimeout: time.Duration(getEnvInt("APP_WRITE_TIMEOUT", 10)) * time.Second,
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "news_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			MaxConns: int32(getEnvInt("DB_MAX_CONNS", 10)),
			MinConns: int32(getEnvInt("DB_MIN_CONNS", 2)),
		},
		JWT: JWTConfig{
			Secret:    getEnv("JWT_SECRET", "change-me-in-production"),
			ExpiresIn: time.Duration(getEnvInt("JWT_EXPIRES_IN_HOURS", 24)) * time.Hour,
		},
		S3: S3Config{
			Endpoint:  getEnv("S3_ENDPOINT", "s3.amazonaws.com"),
			Region:    getEnv("S3_REGION", "ap-southeast-1"),
			Bucket:    getEnv("S3_BUCKET", "news-api"),
			AccessKey: getEnv("S3_ACCESS_KEY", ""),
			SecretKey: getEnv("S3_SECRET_KEY", ""),
			UseSSL:    getEnv("S3_USE_SSL", "true") == "true",
			PublicURL: getEnv("S3_PUBLIC_URL", ""),
		},
		Scheduler: SchedulerConfig{
			Enabled:         getEnv("SCHEDULER_ENABLED", "false") == "true",
			IntervalMinutes: getEnvInt("SCHEDULER_INTERVAL_MINUTES", 10),
			MaxConcurrency:  getEnvInt("SCHEDULER_MAX_CONCURRENCY", 3),
			DefaultStatus:   getEnv("SCHEDULER_DEFAULT_STATUS", "published"),
		},
		AI: AIConfig{
			Provider:    getEnv("AI_PROVIDER", "openai"),
			APIKey:      getEnv("AI_API_KEY", ""),
			BaseURL:     getEnv("AI_BASE_URL", "https://api.openai.com/v1"),
			Model:       getEnv("AI_MODEL", "gpt-4o-mini"),
			MaxTokens:   getEnvInt("AI_MAX_TOKENS", 0),
			Temperature: getEnvFloat("AI_TEMPERATURE", 0.7),
		},
		News: NewsConfig{
			ItemsPerWebsite: getEnvInt("NEWS_ITEMS_PER_WEBSITE", 5),
		},
	}

	return cfg, nil
}

// DSN mengembalikan connection string Postgres
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if v, ok := os.LookupEnv(key); ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}
