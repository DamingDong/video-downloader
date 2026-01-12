package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	BatchSize         int           `json:"batch_size"`
	MaxConcurrency    int           `json:"max_concurrency"`
	TimeoutPerVideo   time.Duration `json:"timeout_per_video"`
	MaxRetries        int           `json:"max_retries"`
	BaseRetryDelay    time.Duration `json:"base_retry_delay"`
	DefaultOutputDir  string        `json:"default_output_dir"`
	ResourceURLsDir   string        `json:"resource_urls_dir"`
	CookieFile        string        `json:"cookie_file"`
	IndexFile         string        `json:"index_file"`
	RecordFile        string        `json:"record_file"`
	DefaultResolution string        `json:"default_resolution"`
	DefaultDownloader string        `json:"default_downloader"`
}

func DefaultConfig() *Config {
	return &Config{
		BatchSize:         10,
		MaxConcurrency:    3,
		TimeoutPerVideo:   60 * time.Minute,
		MaxRetries:        3,
		BaseRetryDelay:    2 * time.Second,
		DefaultOutputDir:  "Output",
		ResourceURLsDir:   "resource_urls",
		CookieFile:        "cookies.txt",
		IndexFile:         ".video_downloaded.index",
		RecordFile:        "下载记录.md",
		DefaultResolution: "720",
		DefaultDownloader: "auto",
	}
}

func LoadConfig(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	if configPath == "" {
		configPath = "config.json"
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return cfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return cfg, nil
}

func (c *Config) SaveConfig(configPath string) error {
	if configPath == "" {
		configPath = "config.json"
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

func (c *Config) GetOutputDir(baseDir string) string {
	if baseDir == "" {
		baseDir = c.ResourceURLsDir
	}
	return baseDir + "/" + c.DefaultOutputDir
}
