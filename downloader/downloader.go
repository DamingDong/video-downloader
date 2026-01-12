package downloader

import (
	"io"
	"time"
)

type VideoInfo struct {
	ID         string
	Title      string
	Duration   int
	Uploader   string
	WebpageURL string
	Extractor  string
	Resolution string
	FileSize   int64
}

type DownloadResult struct {
	Success    bool
	VideoID    string
	Title      string
	FilePath   string
	FileSize   int64
	Error      error
	RetryCount int
}

type Downloader interface {
	Name() string
	SupportedPlatforms() []string
	GetVideoInfo(url string) (*VideoInfo, error)
	Download(url, outputDir, resolution string) (*DownloadResult, error)
	IsDownloaded(videoID string) bool
	MarkDownloaded(videoID string) error
}

type DownloadConfig struct {
	Resolution     string
	OutputDir      string
	MaxRetries     int
	RetryDelay     time.Duration
	MaxConcurrency int
	BatchSize      int
	CookieFile     string
}

type DownloadProgress struct {
	VideoID  string
	Title    string
	Progress float64
	Speed    string
	ETA      string
}

type ProgressReader struct {
	io.Reader
	Total      int64
	Current    int64
	OnProgress func(current, total int64)
}

func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.Reader.Read(p)
	if err == nil {
		pr.Current += int64(n)
		if pr.OnProgress != nil {
			pr.OnProgress(pr.Current, pr.Total)
		}
	}
	return
}
