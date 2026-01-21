package downloader

import (
	"net/url"
	"strings"
)

type SmartDownloader struct {
	youtubeDownloader *YouTubeDownloader
	multiDownloader   *MultiPlatformDownloader
}

func NewSmartDownloader(ytDownloader *YouTubeDownloader, multiDownloader *MultiPlatformDownloader) *SmartDownloader {
	return &SmartDownloader{
		youtubeDownloader: ytDownloader,
		multiDownloader:   multiDownloader,
	}
}

func (sd *SmartDownloader) Name() string {
	return "智能下载器（自动检测）"
}

func (sd *SmartDownloader) SupportedPlatforms() []string {
	return []string{"youtube", "youtube.com", "youtu.be", "douyin", "weibo", "bilibili", "tiktok", "vimeo", "instagram", "twitter", "facebook"}
}

func (sd *SmartDownloader) GetVideoInfo(urlStr string) (*VideoInfo, error) {
	dl := sd.selectDownloader(urlStr)
	return dl.GetVideoInfo(urlStr)
}

func (sd *SmartDownloader) Download(urlStr, outputDir, resolution string) (*DownloadResult, error) {
	dl := sd.selectDownloader(urlStr)
	return dl.Download(urlStr, outputDir, resolution)
}

func (sd *SmartDownloader) IsDownloaded(videoID string) bool {
	return sd.youtubeDownloader.IsDownloaded(videoID)
}

func (sd *SmartDownloader) MarkDownloaded(videoID string) error {
	return sd.youtubeDownloader.MarkDownloaded(videoID)
}

func (sd *SmartDownloader) CheckYTDLP() error {
	return sd.multiDownloader.CheckYTDLP()
}

func (sd *SmartDownloader) selectDownloader(urlStr string) Downloader {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return sd.multiDownloader
	}

	host := parsedURL.Hostname()

	if strings.Contains(host, "youtube.com") || strings.Contains(host, "youtu.be") {
		return sd.youtubeDownloader
	}

	return sd.multiDownloader
}
