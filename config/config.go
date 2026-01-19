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

// ConfigJSON 用于JSON序列化和反序列化的辅助结构体
type ConfigJSON struct {
	BatchSize         int    `json:"batch_size"`
	MaxConcurrency    int    `json:"max_concurrency"`
	TimeoutPerVideo   string `json:"timeout_per_video"`
	MaxRetries        int    `json:"max_retries"`
	BaseRetryDelay    string `json:"base_retry_delay"`
	DefaultOutputDir  string `json:"default_output_dir"`
	ResourceURLsDir   string `json:"resource_urls_dir"`
	CookieFile        string `json:"cookie_file"`
	IndexFile         string `json:"index_file"`
	RecordFile        string `json:"record_file"`
	DefaultResolution string `json:"default_resolution"`
	DefaultDownloader string `json:"default_downloader"`
}

// UnmarshalJSON 实现自定义JSON反序列化方法
func (c *Config) UnmarshalJSON(data []byte) error {
	var jsonCfg ConfigJSON
	if err := json.Unmarshal(data, &jsonCfg); err != nil {
		return err
	}

	// 设置基本字段
	c.BatchSize = jsonCfg.BatchSize
	c.MaxConcurrency = jsonCfg.MaxConcurrency
	c.MaxRetries = jsonCfg.MaxRetries
	c.DefaultOutputDir = jsonCfg.DefaultOutputDir
	c.ResourceURLsDir = jsonCfg.ResourceURLsDir
	c.CookieFile = jsonCfg.CookieFile
	c.IndexFile = jsonCfg.IndexFile
	c.RecordFile = jsonCfg.RecordFile
	c.DefaultResolution = jsonCfg.DefaultResolution
	c.DefaultDownloader = jsonCfg.DefaultDownloader

	// 解析时间字段
	var err error
	if jsonCfg.TimeoutPerVideo != "" {
		c.TimeoutPerVideo, err = time.ParseDuration(jsonCfg.TimeoutPerVideo)
		if err != nil {
			return fmt.Errorf("解析timeout_per_video失败: %w", err)
		}
	}
	
	if jsonCfg.BaseRetryDelay != "" {
		c.BaseRetryDelay, err = time.ParseDuration(jsonCfg.BaseRetryDelay)
		if err != nil {
			return fmt.Errorf("解析base_retry_delay失败: %w", err)
		}
	}

	return nil
}

// MarshalJSON 实现自定义JSON序列化方法
func (c *Config) MarshalJSON() ([]byte, error) {
	jsonCfg := ConfigJSON{
		BatchSize:         c.BatchSize,
		MaxConcurrency:    c.MaxConcurrency,
		TimeoutPerVideo:   c.TimeoutPerVideo.String(),
		MaxRetries:        c.MaxRetries,
		BaseRetryDelay:    c.BaseRetryDelay.String(),
		DefaultOutputDir:  c.DefaultOutputDir,
		ResourceURLsDir:   c.ResourceURLsDir,
		CookieFile:        c.CookieFile,
		IndexFile:         c.IndexFile,
		RecordFile:        c.RecordFile,
		DefaultResolution: c.DefaultResolution,
		DefaultDownloader: c.DefaultDownloader,
	}

	return json.MarshalIndent(jsonCfg, "", "  ")
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
