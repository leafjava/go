package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	LLM      LLMConfig
	LogLevel string
}

type ServerConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type LLMConfig struct {
	BaseURL        string
	APIKey         string
	MockMode       bool
	Model          string
	MaxTokens      int
	Temperature    float64
	StreamChunk    time.Duration
	RequestTimeout time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Addr:         getEnv("SERVER_ADDR", ":8080"),
			ReadTimeout:  getDuration("SERVER_READ_TIMEOUT", "10s"),
			WriteTimeout: getDuration("SERVER_WRITE_TIMEOUT", "60s"), // SSE 长连接要放宽
		},
		LLM: LLMConfig{
			BaseURL:        getEnv("LLM_BASE_URL", "https://api.openai.com/v1"),
			APIKey:         getEnv("LLM_API_KEY", ""),
			MockMode:       getBool("LLM_MOCK_MODE", true),
			Model:          getEnv("LLM_MODEL", "gpt-4o-mini"),
			MaxTokens:      getInt("LLM_MAX_TOKENS", 1024),
			Temperature:    getFloat("LLM_TEMPERATURE", 0.7),
			StreamChunk:    getDuration("LLM_MOCK_STREAM_CHUNK", "30ms"),
			RequestTimeout: getDuration("LLM_REQUEST_TIMEOUT", "30s"),
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func getFloat(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func getBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func getDuration(key, def string) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
		if i, err := strconv.Atoi(v); err == nil {
			return time.Duration(i) * time.Second
		}
	}
	d, _ := time.ParseDuration(def)
	return d
}
