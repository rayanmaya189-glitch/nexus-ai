package config

import (
	"time"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
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
	return "host=" + d.Host + " port=" + itoa(d.Port) + " user=" + d.User +
		" password=" + d.Password + " dbname=" + d.DBName + " sslmode=" + d.SSLMode
}

type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
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
	return DefaultConfig(), nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}
