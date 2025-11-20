package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Security SecurityConfig
	CORS     CORSConfig
    Storage  StorageConfig
}

type AppConfig struct {
	Name  string
	Env   string
	Port  string
	Debug bool
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	TimeZone string
}

type JWTConfig struct {
	Secret             string
	AccessTokenExpiry  int
	RefreshTokenExpiry int
}

type SecurityConfig struct {
	APIKey     string
	BcryptCost int
	RateLimit  int
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

type StorageConfig struct {
    Driver    string // "local" or "s3"
    BasePath  string // "./uploads" for local
    BaseURL   string // "http://localhost:5000/uploads"
    MaxSizeMB int    // Max file size in MB
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	config := &Config{
        App: AppConfig{
            Name:  getEnv("APP_NAME", "BlogPost API"),
            Env:   getEnv("APP_ENV", "development"),
            Port:  getEnv("APP_PORT", "8080"),
            Debug: getEnvBool("APP_DEBUG", true),
        },
        Database: DatabaseConfig{
            Host:     getEnv("DB_HOST", "localhost"),
            Port:     getEnv("DB_PORT", "5432"),
            User:     getEnv("DB_USER", "postgres"),
            Password: getEnv("DB_PASSWORD", "postgres"),
            DBName:   getEnv("DB_NAME", "blogpost_db"),
            SSLMode:  getEnv("DB_SSLMODE", "disable"),
            TimeZone: getEnv("DB_TIMEZONE", "Asia/Jakarta"),
        },
        JWT: JWTConfig{
            Secret:             getEnv("JWT_SECRET", "your-secret-key"),
            AccessTokenExpiry:  getEnvInt("JWT_ACCESS_TOKEN_EXPIRY", 3600),
            RefreshTokenExpiry: getEnvInt("JWT_REFRESH_TOKEN_EXPIRY", 604800),
        },
        Security: SecurityConfig{
            APIKey:     getEnv("API_KEY", "your-api-key"),
            BcryptCost: getEnvInt("BCRYPT_COST", 10),
            RateLimit:  getEnvInt("RATE_LIMIT", 60),
        },
        CORS: CORSConfig{
            AllowedOrigins:   strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"), ","),
            AllowedMethods:   strings.Split(getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS"), ","),
            AllowedHeaders:   strings.Split(getEnv("CORS_ALLOWED_HEADERS", "Origin,Content-Type,Accept,Authorization,X-API-Key"), ","),
            AllowCredentials: getEnvBool("CORS_ALLOW_CREDENTIALS", true),
        },
        Storage: StorageConfig{
            Driver:    getEnv("STORAGE_DRIVER", "local"),
            BasePath:  getEnv("STORAGE_BASE_PATH", "./uploads"),
            BaseURL:   getEnv("STORAGE_BASE_URL", "http://localhost:5000/uploads"),
            MaxSizeMB: getEnvInt("STORAGE_MAX_SIZE_MB", 2),
        },
    }

	if err := config.Validate(); err != nil {
        return nil, err
    }

    return config, nil
}

func (c *Config) Validate() error {
    if c.JWT.Secret == "your-secret-key" && c.App.Env == "production" {
        return fmt.Errorf("JWT_SECRET must be set in production")
    }
    if c.Security.APIKey == "your-api-key" && c.App.Env == "production" {
        return fmt.Errorf("API_KEY must be set in production")
    }
    return nil
}

func (c *DatabaseConfig) DSN() string {
    return fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
        c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode, c.TimeZone,
    )
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        if boolValue, err := strconv.ParseBool(value); err == nil {
            return boolValue
        }
    }
    return defaultValue
}