package downloader

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"batch_download_videos/config"
	"batch_download_videos/indexer"
	"batch_download_videos/utils"
)

type MultiPlatformDownloader struct {
	config    *config.Config
	indexer   *indexer.Indexer
	outputDir string
}

func NewMultiPlatformDownloader(cfg *config.Config, idx *indexer.Indexer) *MultiPlatformDownloader {
	return &MultiPlatformDownloader{
		config:    cfg,
		indexer:   idx,
		outputDir: cfg.GetOutputDir(""),
	}
}

func (mpd *MultiPlatformDownloader) Name() string {
	return "多平台下载器"
}

func (mpd *MultiPlatformDownloader) SupportedPlatforms() []string {
	return []string{
		"youtube", "youtu.be",
		"douyin",
		"weibo",
		"bilibili",
		"tiktok",
		"vimeo",
		"instagram",
		"twitter", "x.com",
		"facebook",
	}
}

func (mpd *MultiPlatformDownloader) GetVideoInfo(url string) (*VideoInfo, error) {
	cmd := exec.Command("yt-dlp",
		"--dump-json",
		"--no-playlist",
		url)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取视频信息失败: %w", err)
	}

	var info struct {
		ID         string `json:"id"`
		Title      string `json:"title"`
		Duration   int    `json:"duration"`
		Uploader   string `json:"uploader"`
		WebpageURL string `json:"webpage_url"`
		Extractor  string `json:"extractor_key"`
	}

	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("解析视频信息失败: %w", err)
	}

	return &VideoInfo{
		ID:         info.ID,
		Title:      info.Title,
		Duration:   info.Duration,
		Uploader:   info.Uploader,
		WebpageURL: url,
		Extractor:  info.Extractor,
		Resolution: "",
		FileSize:   0,
	}, nil
}

func (mpd *MultiPlatformDownloader) Download(url, outputDir, resolution string) (*DownloadResult, error) {
	info, err := mpd.GetVideoInfo(url)
	if err != nil {
		return nil, err
	}

	website := utils.GetWebsiteType(url)
	uniqueID := mpd.getUniqueID(url, info)

	if mpd.indexer.IsDownloaded(uniqueID) {
		return &DownloadResult{
			Success:    false,
			VideoID:    uniqueID,
			Title:      info.Title,
			FilePath:   "",
			FileSize:   0,
			Error:      fmt.Errorf("视频已下载"),
			RetryCount: 0,
		}, nil
	}

	log.Printf("开始下载: %s (ID: %s, 网站: %s, 分辨率: %s)", info.Title, uniqueID, website, resolution)

	qualityFormat := utils.GetQualityFormat(resolution)
	outputTemplate := filepath.Join(outputDir, "%(title)s.%(ext)s")

	filename := utils.SanitizeFilename(info.Title) + ".mp4"
	filePath := filepath.Join(outputDir, filename)

	if err := utils.CleanupZeroByteFiles(filePath); err != nil {
		log.Printf("清理0字节文件失败: %v", err)
	}

	args := []string{
		"-f", qualityFormat,
		"-o", outputTemplate,
		"--no-playlist",
		"--no-warnings",
	}

	if _, err := os.Stat(mpd.config.CookieFile); err == nil {
		args = append(args, "--cookies", mpd.config.CookieFile)
		log.Printf("使用Cookie文件: %s", mpd.config.CookieFile)
	}

	args = append(args, url)

	var lastErr error
	for retry := 0; retry < mpd.config.MaxRetries; retry++ {
		if retry > 0 {
			delay := mpd.config.BaseRetryDelay * time.Duration(retry)
			log.Printf("重试 %d/%d，等待 %v 后重试...", retry, mpd.config.MaxRetries, delay)
			time.Sleep(delay)
		}

		cmd := exec.Command("yt-dlp", args...)
		if err := cmd.Run(); err != nil {
			lastErr = err
			log.Printf("下载失败 (尝试 %d/%d): %v", retry+1, mpd.config.MaxRetries, err)
			continue
		}

		mpd.indexer.MarkDownloaded(uniqueID)

		fileInfo, _ := os.Stat(filePath)
		fileSize := int64(0)
		if fileInfo != nil {
			fileSize = fileInfo.Size()
		}

		log.Printf("下载完成: %s (ID: %s)", info.Title, uniqueID)
		return &DownloadResult{
			Success:    true,
			VideoID:    uniqueID,
			Title:      info.Title,
			FilePath:   filePath,
			FileSize:   fileSize,
			Error:      nil,
			RetryCount: retry,
		}, nil
	}

	return nil, fmt.Errorf("下载失败: %w (尝试 %d 次后放弃)", lastErr, mpd.config.MaxRetries)
}

func (mpd *MultiPlatformDownloader) IsDownloaded(videoID string) bool {
	return mpd.indexer.IsDownloaded(videoID)
}

func (mpd *MultiPlatformDownloader) MarkDownloaded(videoID string) error {
	mpd.indexer.MarkDownloaded(videoID)
	return nil
}

func (mpd *MultiPlatformDownloader) getUniqueID(url string, info *VideoInfo) string {
	if info.ID != "" {
		return info.ID
	}
	return url
}

func (mpd *MultiPlatformDownloader) CheckYTDLP() error {
	cmd := exec.Command("yt-dlp", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("yt-dlp未安装，请先安装并配置到PATH中（多平台下载依赖）")
	}
	return nil
}
