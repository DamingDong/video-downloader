package downloader

import (
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
	// 对于智能下载器，我们需要同时检查两个下载器的索引
	return sd.youtubeDownloader.IsDownloaded(videoID) || sd.multiDownloader.IsDownloaded(videoID)
}

func (sd *SmartDownloader) MarkDownloaded(videoID string) error {
	// 对于智能下载器，我们需要同时标记两个下载器的索引
	if err := sd.youtubeDownloader.MarkDownloaded(videoID); err != nil {
		return err
	}
	return sd.multiDownloader.MarkDownloaded(videoID)
}

func (sd *SmartDownloader) CheckYTDLP() error {
	return sd.multiDownloader.CheckYTDLP()
}

func (sd *SmartDownloader) selectDownloader(urlStr string) Downloader {
	// 检查是否为 YouTube URL
	if strings.Contains(urlStr, "youtube.com") || strings.Contains(urlStr, "youtu.be") {
		// 检查是否为播放列表：包含 list 参数
		if strings.Contains(urlStr, "list=") {
			return sd.multiDownloader
		}

		// 检查是否为频道
		if strings.Contains(urlStr, "/channel/") || strings.Contains(urlStr, "/c/") || strings.Contains(urlStr, "/user/") {
			return sd.multiDownloader
		}

		// 单个视频：使用 YouTube 专用下载器
		return sd.youtubeDownloader
	}

	// 非 YouTube URL：使用多平台下载器
	return sd.multiDownloader
}
