package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
	NATS     NATSConfig     `json:"nats"`
	JWT      JWTConfig      `json:"jwt"`
	Ollama   OllamaConfig   `json:"ollama"`
}

type ServerConfig struct {
	HTTPPort     int           `json:"http_port"`
	GRPCPort     int           `json:"grpc_port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	Debug        bool          `json:"debug"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

func (d *DatabaseConfig) ConnectionString() string {
	return "host=" + d.Host + " port=" + strconv.Itoa(d.Port) + " user=" + d.User +
		" password=" + d.Password + " dbname=" + d.DBName + " sslmode=" + d.SSLMode
}

type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

type NATSConfig struct {
	URL            string `json:"url"`
	MaxReconnect   int    `json:"max_reconnect"`
	ReconnectWait  time.Duration `json:"reconnect_wait"`
}

type JWTConfig struct {
	Secret          string        `json:"secret"`
	Issuer          string        `json:"issuer"`
	AccessTokenTTL  time.Duration `json:"access_token_ttl"`
	RefreshTokenTTL time.Duration `json:"refresh_token_ttl"`
}

type OllamaConfig struct {
	BaseURL string        `json:"base_url"`
	Timeout time.Duration `json:"timeout"`
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			HTTPPort:     8080,
			GRPCPort:     50050,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			Debug:        false,
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "nexus",
			Password: "nexus_secret_2026",
			DBName:   "nexus_main",
			SSLMode:  "disable",
		},
		Redis: RedisConfig{
			Addr: "localhost:6379",
			DB:   0,
		},
		NATS: NATSConfig{
			URL:           "nats://localhost:4222",
			MaxReconnect:  10,
			ReconnectWait: 2 * time.Second,
		},
		JWT: JWTConfig{
			Secret:          "nexus-jwt-secret-2026",
			Issuer:          "aeroxe-nexus-ai",
			AccessTokenTTL:  1 * time.Hour,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
		Ollama: OllamaConfig{
			BaseURL: "http://localhost:11434",
			Timeout: 120 * time.Second,
		},
	}
}

func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path != "" {
		if err := loadFromFile(path, cfg); err != nil {
			return nil, err
		}
	}

	applyEnvOverrides(cfg)

	return cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	// Server
	if v := os.Getenv("HTTP_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Server.HTTPPort = n
		}
	}
	if v := os.Getenv("GRPC_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Server.GRPCPort = n
		}
	}
	if v := os.Getenv("PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Server.HTTPPort = n
		}
	}
	if v := os.Getenv("DEBUG"); v != "" {
		cfg.Server.Debug = v == "true" || v == "1"
	}
	if v := os.Getenv("READ_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Server.ReadTimeout = d
		}
	}
	if v := os.Getenv("WRITE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Server.WriteTimeout = d
		}
	}

	// Database
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Database.Port = n
		}
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.Database.DBName = v
	}
	if v := os.Getenv("DB_SSLMODE"); v != "" {
		cfg.Database.SSLMode = v
	}

	// Redis
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		cfg.Redis.Addr = v
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}
	if v := os.Getenv("REDIS_DB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Redis.DB = n
		}
	}

	// NATS
	if v := os.Getenv("NATS_URL"); v != "" {
		cfg.NATS.URL = v
	}
	if v := os.Getenv("NATS_MAX_RECONNECT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.NATS.MaxReconnect = n
		}
	}

	// JWT
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}
	if v := os.Getenv("JWT_ISSUER"); v != "" {
		cfg.JWT.Issuer = v
	}
	if v := os.Getenv("JWT_ACCESS_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.JWT.AccessTokenTTL = d
		}
	}
	if v := os.Getenv("JWT_REFRESH_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.JWT.RefreshTokenTTL = d
		}
	}

	// Ollama
	if v := os.Getenv("OLLAMA_BASE_URL"); v != "" {
		cfg.Ollama.BaseURL = v
	}
	if v := os.Getenv("OLLAMA_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Ollama.Timeout = d
		}
	}
}

func loadFromFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return parseJSON(data, cfg)
}

func parseJSON(data []byte, cfg *Config) error {
	if len(data) == 0 {
		return nil
	}
	// Use encoding/json via a helper to avoid importing here
	return jsonUnmarshal(data, cfg)
}
