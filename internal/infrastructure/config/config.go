package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Bus      BusConfig
	Email    EmailConfig
	Purge    PurgeConfig
}

type ServerConfig struct {
	Host         string
	Port         string
	Environment  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type BusConfig struct {
	WorkerPoolSize  int
	BufferSize      int
	RetryAttempts   int
	RetryDelay      time.Duration
	RetryMaxDelay   time.Duration
	RetryMultiplier float64
}

type EmailConfig struct {
	FromAddress string
	FromName    string
	AppURL      string
	AppName     string
}

type PurgeConfig struct {
	Enabled                   bool
	Schedule                  string
	DryRun                    bool
	BatchSize                 int
	RetentionDaysUserPosts    int
	RetentionDaysPostComments int
}

// Load loads configuration from .env file and environment variables
func Load() (*Config, error) {
	// Load .env file (if exists)
	// Priority: .env.{environment} > .env.local > .env
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	// Try to load files in priority order
	envFiles := []string{
		fmt.Sprintf(".env.%s.local", env), // .env.development.local (highest priority)
		fmt.Sprintf(".env.%s", env),       // .env.development
		".env.local",                      // .env.local (not committed to git)
		".env",                            // .env (default)
	}

	// Load first found file
	for _, file := range envFiles {
		if err := godotenv.Load(file); err == nil {
			fmt.Printf("Loaded config from:  %s\n", file)
			break
		}
	}

	// If no file found - not an error, use environment variables
	return &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnv("SERVER_PORT", "8080"),
			Environment:  getEnv("ENVIRONMENT", "development"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			DBName:          getEnv("DB_NAME", "promenade"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			AccessTokenTTL:  time.Duration(getEnvAsInt("JWT_ACCESS_TTL_MINUTES", 15)) * time.Minute,
			RefreshTokenTTL: time.Duration(getEnvAsInt("JWT_REFRESH_TTL_HOURS", 168)) * time.Hour, // 7 days
		},
		Bus: BusConfig{
			WorkerPoolSize:  getEnvAsInt("BUS_WORKER_POOL_SIZE", 10),
			BufferSize:      getEnvAsInt("BUS_BUFFER_SIZE", 1000),
			RetryAttempts:   getEnvAsInt("BUS_RETRY_ATTEMPTS", 3),
			RetryDelay:      getDurationEnv("BUS_RETRY_DELAY", 1*time.Second),
			RetryMaxDelay:   getDurationEnv("BUS_RETRY_MAX_DELAY", 5*time.Second),
			RetryMultiplier: getEnvAsFloat("BUS_RETRY_MULTIPLIER", 2.0),
		},
		Email: EmailConfig{
			FromAddress: getEnv("EMAIL_FROM_ADDRESS", "noreply@promenade.com"),
			FromName:    getEnv("EMAIL_FROM_NAME", "Promenade Team"),
			AppURL:      getEnv("APP_URL", "http://localhost:8081"),
			AppName:     getEnv("APP_NAME", "Promenade"),
		},
		Purge: PurgeConfig{
			Enabled:                   getEnvAsBool("PURGE_ENABLED", true),
			Schedule:                  getEnv("PURGE_SCHEDULE", "0 2 * * *"), // 2 AM daily
			DryRun:                    getEnvAsBool("PURGE_DRY_RUN", false),
			BatchSize:                 getEnvAsInt("PURGE_BATCH_SIZE", 1000),
			RetentionDaysUserPosts:    getEnvAsInt("PURGE_RETENTION_USER_POSTS", 90),
			RetentionDaysPostComments: getEnvAsInt("PURGE_RETENTION_POST_COMMENTS", 30),
		},
	}, nil
}

// LoadFromFile loads configuration from specific file
func LoadFromFile(filename string) (*Config, error) {
	if err := godotenv.Load(filename); err != nil {
		return nil, fmt.Errorf("failed to load config from %s: %w", filename, err)
	}
	return Load()
}

// Validate checks configuration correctness
func (c *Config) Validate() error {
	if c.JWT.Secret == "your-secret-key-change-in-production" {
		return fmt.Errorf("JWT_SECRET must be changed in production")
	}
	if c.Server.Environment == "production" && c.Database.SSLMode == "disable" {
		return fmt.Errorf("SSL should be enabled in production")
	}
	return nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
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
func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
