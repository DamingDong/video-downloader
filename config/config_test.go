package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.json")

	content := `{
		"batch_size": 5,
		"max_concurrency": 2,
		"timeout_per_video": "30m0s",
		"max_retries": 2,
		"base_retry_delay": "1s",
		"default_output_dir": "TestOutput",
		"resource_urls_dir": "test_urls",
		"cookie_file": "test_cookies.txt",
		"index_file": ".test_index",
		"record_file": "test_record.md",
		"default_resolution": "480",
		"default_downloader": "youtube",
		"output_template": "%(title)s.%(ext)s",
		"filename_max_length": 150,
		"generate_meta_file": false,
		"recode_video": "webm",
		"max_concurrent_downloads": 2,
		"proxy": "http://proxy.example.com:8080",
		"limit_rate": "1M"
	}`

	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	cfg, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if cfg.BatchSize != 5 {
		t.Errorf("BatchSize = %d, want 5", cfg.BatchSize)
	}

	if cfg.MaxConcurrency != 2 {
		t.Errorf("MaxConcurrency = %d, want 2", cfg.MaxConcurrency)
	}

	if cfg.TimeoutPerVideo != 30*time.Minute {
		t.Errorf("TimeoutPerVideo = %v, want 30m0s", cfg.TimeoutPerVideo)
	}

	if cfg.MaxRetries != 2 {
		t.Errorf("MaxRetries = %d, want 2", cfg.MaxRetries)
	}

	if cfg.BaseRetryDelay != 1*time.Second {
		t.Errorf("BaseRetryDelay = %v, want 1s", cfg.BaseRetryDelay)
	}

	if cfg.DefaultOutputDir != "TestOutput" {
		t.Errorf("DefaultOutputDir = %q, want %q", cfg.DefaultOutputDir, "TestOutput")
	}

	if cfg.ResourceURLsDir != "test_urls" {
		t.Errorf("ResourceURLsDir = %q, want %q", cfg.ResourceURLsDir, "test_urls")
	}

	if cfg.CookieFile != "test_cookies.txt" {
		t.Errorf("CookieFile = %q, want %q", cfg.CookieFile, "test_cookies.txt")
	}

	if cfg.IndexFile != ".test_index" {
		t.Errorf("IndexFile = %q, want %q", cfg.IndexFile, ".test_index")
	}

	if cfg.RecordFile != "test_record.md" {
		t.Errorf("RecordFile = %q, want %q", cfg.RecordFile, "test_record.md")
	}

	if cfg.DefaultResolution != "480" {
		t.Errorf("DefaultResolution = %q, want %q", cfg.DefaultResolution, "480")
	}

	if cfg.DefaultDownloader != "youtube" {
		t.Errorf("DefaultDownloader = %q, want %q", cfg.DefaultDownloader, "youtube")
	}

	if cfg.OutputTemplate != "%(title)s.%(ext)s" {
		t.Errorf("OutputTemplate = %q, want %q", cfg.OutputTemplate, "%(title)s.%(ext)s")
	}

	if cfg.FilenameMaxLength != 150 {
		t.Errorf("FilenameMaxLength = %d, want 150", cfg.FilenameMaxLength)
	}

	if cfg.GenerateMetaFile != false {
		t.Errorf("GenerateMetaFile = %t, want false", cfg.GenerateMetaFile)
	}

	if cfg.RecodeVideo != "webm" {
		t.Errorf("RecodeVideo = %q, want %q", cfg.RecodeVideo, "webm")
	}

	if cfg.MaxConcurrentDownloads != 2 {
		t.Errorf("MaxConcurrentDownloads = %d, want 2", cfg.MaxConcurrentDownloads)
	}

	if cfg.Proxy != "http://proxy.example.com:8080" {
		t.Errorf("Proxy = %q, want %q", cfg.Proxy, "http://proxy.example.com:8080")
	}

	if cfg.LimitRate != "1M" {
		t.Errorf("LimitRate = %q, want %q", cfg.LimitRate, "1M")
	}
}

func TestLoadConfigNonExistent(t *testing.T) {
	cfg, err := LoadConfig("non_existent_config.json")
	if err != nil {
		t.Fatalf("LoadConfig() should not fail for non-existent file: %v", err)
	}

	// Check if default values are set
	if cfg.BatchSize != 10 {
		t.Errorf("Default BatchSize = %d, want 10", cfg.BatchSize)
	}

	if cfg.MaxConcurrency != 3 {
		t.Errorf("Default MaxConcurrency = %d, want 3", cfg.MaxConcurrency)
	}

	if cfg.TimeoutPerVideo != 1*time.Hour {
		t.Errorf("Default TimeoutPerVideo = %v, want 1h0m0s", cfg.TimeoutPerVideo)
	}
}

func TestSaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.json")

	cfg := &Config{
		BatchSize:              10,
		MaxConcurrency:         3,
		TimeoutPerVideo:        1 * time.Hour,
		MaxRetries:             3,
		BaseRetryDelay:         2 * time.Second,
		DefaultOutputDir:       "Output",
		ResourceURLsDir:        "resource_urls",
		CookieFile:             "cookies.txt",
		IndexFile:              ".video_downloaded.index",
		RecordFile:             "下载记录.md",
		DefaultResolution:      "720",
		DefaultDownloader:      "multi",
		OutputTemplate:         "%(upload_date)s_%(title)s.%(ext)s",
		FilenameMaxLength:      200,
		GenerateMetaFile:       true,
		RecodeVideo:            "mp4",
		MaxConcurrentDownloads: 3,
		Proxy:                  "",
		LimitRate:              "",
	}

	if err := cfg.SaveConfig(configFile); err != nil {
		t.Fatalf("SaveConfig() failed: %v", err)
	}

	// Verify the file was written
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Fatalf("Config file was not created: %v", err)
	}

	// Load it back and verify
	loadedCfg, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedCfg.BatchSize != cfg.BatchSize {
		t.Errorf("Saved BatchSize = %d, want %d", loadedCfg.BatchSize, cfg.BatchSize)
	}

	if loadedCfg.DefaultResolution != cfg.DefaultResolution {
		t.Errorf("Saved DefaultResolution = %q, want %q", loadedCfg.DefaultResolution, cfg.DefaultResolution)
	}
}

func TestGetOutputDir(t *testing.T) {
	cfg := &Config{
		DefaultOutputDir: "DefaultOutput",
	}

	// Test with empty baseDir
	outputDir := cfg.GetOutputDir("")
	if outputDir != "DefaultOutput" {
		t.Errorf("GetOutputDir() with empty baseDir = %q, want %q", outputDir, "DefaultOutput")
	}

	// Test with baseDir
	outputDir = cfg.GetOutputDir("subdir")
	if outputDir != "subdir" {
		t.Errorf("GetOutputDir() with baseDir = %q, want %q", outputDir, "subdir")
	}
}
