package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	BatchSize              int               `json:"batch_size"`
	MaxConcurrency         int               `json:"max_concurrency"`
	TimeoutPerVideo        time.Duration     `json:"timeout_per_video"`
	MaxRetries             int               `json:"max_retries"`
	BaseRetryDelay         time.Duration     `json:"base_retry_delay"`
	DefaultOutputDir       string            `json:"default_output_dir"`
	PlatformOutputDirs     map[string]string `json:"platform_output_dirs"`
	ResourceUrlsDir        string            `json:"resource_urls_dir"`
	CookieFile             string            `json:"cookie_file"`
	IndexFile              string            `json:"index_file"`
	RecordFile             string            `json:"record_file"`
	DefaultResolution      string            `json:"default_resolution"`
	DefaultDownloader      string            `json:"default_downloader"`
	GenerateMetaFile       bool              `json:"generate_meta_file"`
	OutputTemplate         string            `json:"output_template"`
	FilenameMaxLength      int               `json:"filename_max_length"`
	RecodeVideo            string            `json:"recode_video"`
	MaxConcurrentDownloads int               `json:"max_concurrent_downloads"`
	Proxy                  string            `json:"proxy"`
	LimitRate              string            `json:"limit_rate"`
	FfmpegPath             string            `json:"ffmpeg_path"`
}

// ConfigJSON 用于JSON序列化和反序列化的辅助结构体
type ConfigJSON struct {
	BatchSize              int               `json:"batch_size"`
	MaxConcurrency         int               `json:"max_concurrency"`
	TimeoutPerVideo        string            `json:"timeout_per_video"`
	MaxRetries             int               `json:"max_retries"`
	BaseRetryDelay         string            `json:"base_retry_delay"`
	DefaultOutputDir       string            `json:"default_output_dir"`
	PlatformOutputDirs     map[string]string `json:"platform_output_dirs"`
	ResourceUrlsDir        string            `json:"resource_urls_dir"`
	CookieFile             string            `json:"cookie_file"`
	IndexFile              string            `json:"index_file"`
	RecordFile             string            `json:"record_file"`
	DefaultResolution      string            `json:"default_resolution"`
	DefaultDownloader      string            `json:"default_downloader"`
	GenerateMetaFile       bool              `json:"generate_meta_file"`
	OutputTemplate         string            `json:"output_template"`
	FilenameMaxLength      int               `json:"filename_max_length"`
	RecodeVideo            string            `json:"recode_video"`
	MaxConcurrentDownloads int               `json:"max_concurrent_downloads"`
	Proxy                  string            `json:"proxy"`
	LimitRate              string            `json:"limit_rate"`
	FfmpegPath             string            `json:"ffmpeg_path"`
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
	c.PlatformOutputDirs = jsonCfg.PlatformOutputDirs
	c.ResourceUrlsDir = jsonCfg.ResourceUrlsDir
	c.CookieFile = jsonCfg.CookieFile
	c.IndexFile = jsonCfg.IndexFile
	c.RecordFile = jsonCfg.RecordFile
	c.DefaultResolution = jsonCfg.DefaultResolution
	c.DefaultDownloader = jsonCfg.DefaultDownloader
	c.OutputTemplate = jsonCfg.OutputTemplate
	c.FilenameMaxLength = jsonCfg.FilenameMaxLength
	c.GenerateMetaFile = jsonCfg.GenerateMetaFile
	c.RecodeVideo = jsonCfg.RecodeVideo
	c.MaxConcurrentDownloads = jsonCfg.MaxConcurrentDownloads
	c.Proxy = jsonCfg.Proxy
	c.LimitRate = jsonCfg.LimitRate
	c.FfmpegPath = jsonCfg.FfmpegPath

	// 解析时间字段
	var err error
	if jsonCfg.TimeoutPerVideo != "" {
		c.TimeoutPerVideo, err = time.ParseDuration(jsonCfg.TimeoutPerVideo)
		if err != nil {
			return fmt.Errorf("解析TimeoutPerVideo失败: %w", err)
		}
	}

	if jsonCfg.BaseRetryDelay != "" {
		c.BaseRetryDelay, err = time.ParseDuration(jsonCfg.BaseRetryDelay)
		if err != nil {
			return fmt.Errorf("解析BaseRetryDelay失败: %w", err)
		}
	}

	return nil
}

// MarshalJSON 实现自定义JSON序列化方法
func (c *Config) MarshalJSON() ([]byte, error) {
	jsonCfg := ConfigJSON{
		BatchSize:              c.BatchSize,
		MaxConcurrency:         c.MaxConcurrency,
		TimeoutPerVideo:        c.TimeoutPerVideo.String(),
		MaxRetries:             c.MaxRetries,
		BaseRetryDelay:         c.BaseRetryDelay.String(),
		DefaultOutputDir:       c.DefaultOutputDir,
		PlatformOutputDirs:     c.PlatformOutputDirs,
		ResourceUrlsDir:        c.ResourceUrlsDir,
		CookieFile:             c.CookieFile,
		IndexFile:              c.IndexFile,
		RecordFile:             c.RecordFile,
		DefaultResolution:      c.DefaultResolution,
		DefaultDownloader:      c.DefaultDownloader,
		OutputTemplate:         c.OutputTemplate,
		FilenameMaxLength:      c.FilenameMaxLength,
		GenerateMetaFile:       c.GenerateMetaFile,
		RecodeVideo:            c.RecodeVideo,
		MaxConcurrentDownloads: c.MaxConcurrentDownloads,
		Proxy:                  c.Proxy,
		LimitRate:              c.LimitRate,
		FfmpegPath:             c.FfmpegPath,
	}

	return json.MarshalIndent(jsonCfg, "", "  ")
}

func DefaultConfig() *Config {
	return &Config{
		BatchSize:        10,
		MaxConcurrency:   3,
		TimeoutPerVideo:  60 * time.Minute,
		MaxRetries:       3,
		BaseRetryDelay:   2 * time.Second,
		DefaultOutputDir: "output",
		PlatformOutputDirs: map[string]string{
			"youtube":  "output/youtube",
			"douyin":   "output/douyin",
			"bilibili": "output/bilibili",
			"tiktok":   "output/tiktok",
			"other":    "output/other",
		},
		ResourceUrlsDir:        "resource_urls",
		CookieFile:             "cookies.txt",
		IndexFile:              ".video_downloaded.index",
		RecordFile:             "下载记录.md",
		DefaultResolution:      "720",
		DefaultDownloader:      "auto",
		GenerateMetaFile:       true,
		OutputTemplate:         "%(platform)s_%(content_type)s_%(title)s_%(id)s_%(timestamp)s.%(ext)s",
		FilenameMaxLength:      0,
		RecodeVideo:            "",
		MaxConcurrentDownloads: 3,
		Proxy:                  "",
		LimitRate:              "",
		FfmpegPath:             "./deps/ffmpeg.exe",
	}
}

func LoadConfig(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	if configPath == "" {
		configPath = "config.json"
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，保存默认配置到文件
		if err := cfg.SaveConfig(configPath); err != nil {
			return nil, fmt.Errorf("保存默认配置文件失败: %w", err)
		}
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
	// 如果提供了 baseDir，使用它作为基础目录
	// 否则直接返回默认输出目录
	if baseDir != "" {
		return baseDir
	}
	return c.DefaultOutputDir
}

// GetPlatformOutputDir 根据平台获取对应的输出目录
func (c *Config) GetPlatformOutputDir(platform string) string {
	// 如果PlatformOutputDirs不为空且包含该平台的配置，返回对应的目录
	if c.PlatformOutputDirs != nil {
		if dir, ok := c.PlatformOutputDirs[platform]; ok {
			return dir
		}
		// 如果平台不存在但有other配置，返回other目录
		if dir, ok := c.PlatformOutputDirs["other"]; ok {
			return dir
		}
	}
	// 否则返回默认输出目录
	return c.DefaultOutputDir
}
