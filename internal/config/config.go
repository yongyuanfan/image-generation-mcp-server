package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr             string
	APIPrefix            string
	MCPPath              string
	ARKAPIKey            string
	ARKBaseURL           string
	ARKImageEndpointPath string
	ARKTextModel         string
	ARKImageModel        string
	RequestTimeout       time.Duration
	MCPServerName        string
	MCPServerVersion     string
}

func Load() (Config, error) {
	if err := loadDotEnvIfPresent(); err != nil {
		return Config{}, err
	}

	timeoutSeconds, err := intFromEnv("REQUEST_TIMEOUT_SECONDS", 120)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		HTTPAddr:             envOrDefault("HTTP_ADDR", ":9101"),
		APIPrefix:            normalizePath(envOrDefault("API_PREFIX", "/api/v1")),
		MCPPath:              normalizePath(envOrDefault("MCP_PATH", "/mcp")),
		ARKAPIKey:            strings.TrimSpace(os.Getenv("ARK_API_KEY")),
		ARKBaseURL:           strings.TrimRight(envOrDefault("ARK_BASE_URL", "https://ark.cn-beijing.volces.com/api/v3"), "/"),
		ARKImageEndpointPath: normalizePath(envOrDefault("ARK_IMAGE_ENDPOINT_PATH", "/images/generations")),
		ARKTextModel:         envOrDefault("ARK_MODEL_TEXT2IMAGE", "doubao-seedream-4-5-251128"),
		ARKImageModel:        envOrDefault("ARK_MODEL_IMAGE2IMAGE", "doubao-seedream-4-5-251128"),
		RequestTimeout:       time.Duration(timeoutSeconds) * time.Second,
		MCPServerName:        envOrDefault("MCP_SERVER_NAME", "seedream-image-server"),
		MCPServerVersion:     envOrDefault("MCP_SERVER_VERSION", "0.1.0"),
	}

	if cfg.ARKAPIKey == "" {
		return Config{}, fmt.Errorf("ARK_API_KEY is required")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func intFromEnv(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("invalid %s: must be greater than 0", key)
	}
	return parsed, nil
}

func normalizePath(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "/"
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	return strings.TrimRight(trimmed, "/")
}

func loadDotEnvIfPresent() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	file, err := os.Open(filepath.Join(wd, ".env"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		if key == "" || os.Getenv(key) != "" {
			continue
		}

		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return scanner.Err()
}
