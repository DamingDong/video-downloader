package downloader

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"
	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"

	"batch_download_videos/config"
	"batch_download_videos/indexer"
	"batch_download_videos/utils"
)

type YouTubeDownloader struct {
	client    *youtube.Client
	config    *config.Config
	indexer   *indexer.Indexer
	outputDir string
}

func NewYouTubeDownloader(cfg *config.Config, idx *indexer.Indexer) *YouTubeDownloader {
	client := youtube.Client{}

	return &YouTubeDownloader{
		client:    &client,
		config:    cfg,
		indexer:   idx,
		outputDir: cfg.GetOutputDir(""),
	}
}

func (ytd *YouTubeDownloader) Name() string {
	return "YouTube专用下载器"
}

func (ytd *YouTubeDownloader) SupportedPlatforms() []string {
	return []string{"youtube", "youtu.be"}
}

func (ytd *YouTubeDownloader) GetVideoInfo(url string) (*VideoInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	videoID, err := youtube.ExtractVideoID(url)
	if err != nil {
		return nil, fmt.Errorf("提取视频ID失败: %w", err)
	}

	video, err := ytd.client.GetVideoContext(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("获取视频失败: %w", err)
	}

	return &VideoInfo{
		ID:         video.ID,
		Title:      video.Title,
		Duration:   int(video.Duration.Seconds()),
		Uploader:   video.Author,
		WebpageURL: url,
		Extractor:  "youtube",
		Resolution: "",
		FileSize:   0,
	}, nil
}

func (ytd *YouTubeDownloader) Download(url, outputDir, resolution string) (*DownloadResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ytd.config.TimeoutPerVideo)
	defer cancel()

	videoID, err := youtube.ExtractVideoID(url)
	if err != nil {
		return nil, fmt.Errorf("提取视频ID失败: %w", err)
	}

	video, err := ytd.client.GetVideoContext(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("获取视频失败: %w", err)
	}

	if ytd.indexer.IsDownloaded(video.ID) {
		return &DownloadResult{
			Success:    false,
			VideoID:    video.ID,
			Title:      video.Title,
			FilePath:   "",
			FileSize:   0,
			Error:      fmt.Errorf("视频已下载"),
			RetryCount: 0,
		}, nil
	}

	// 生成符合OutputTemplate的文件名
	filename := ytd.generateFilename(video, ".mp4")
	outputPath := filepath.Join(outputDir, filename)

	if err := utils.CleanupZeroByteFiles(outputPath); err != nil {
		log.Printf("清理0字节文件失败: %v", err)
	}

	log.Printf("开始下载: %s (ID: %s)", video.Title, video.ID)

	var lastErr error
	for retry := 0; retry < ytd.config.MaxRetries; retry++ {
		if retry > 0 {
			delay := ytd.config.BaseRetryDelay * time.Duration(retry)
			log.Printf("重试 %d/%d，等待 %v 后重试...", retry, ytd.config.MaxRetries, delay)
			time.Sleep(delay)
		}

		err := ytd.downloadVideo(ctx, video, filename, outputDir, resolution)
		if err != nil {
			lastErr = err
			log.Printf("下载失败 (尝试 %d/%d): %v", retry+1, ytd.config.MaxRetries, err)
			continue
		}

		ytd.indexer.MarkDownloaded(video.ID)

		// 处理Meta文件的生成
		if ytd.config.GenerateMetaFile {
			if err := ytd.generateMetaFile(video, outputDir, filename); err != nil {
				log.Printf("生成Meta文件失败: %v", err)
			}
		}

		// 处理视频格式转换
		if ytd.config.RecodeVideo != "" {
			newFilename := strings.TrimSuffix(filename, filepath.Ext(filename)) + "." + ytd.config.RecodeVideo
			newOutputPath := filepath.Join(outputDir, newFilename)
			if err := ytd.convertVideoFormat(outputPath, newOutputPath); err != nil {
				log.Printf("视频格式转换失败: %v", err)
			} else {
				// 更新输出路径和文件名
				outputPath = newOutputPath
				filename = newFilename
			}
		}

		info, _ := os.Stat(outputPath)
		fileSize := int64(0)
		if info != nil {
			fileSize = info.Size()
		}

		log.Printf("下载完成: %s (ID: %s)", video.Title, video.ID)
		return &DownloadResult{
			Success:    true,
			VideoID:    video.ID,
			Title:      video.Title,
			FilePath:   outputPath,
			FileSize:   fileSize,
			Error:      nil,
			RetryCount: retry,
		}, nil
	}

	return nil, fmt.Errorf("下载失败: %w (尝试 %d 次后放弃)", lastErr, ytd.config.MaxRetries)
}

func (ytd *YouTubeDownloader) IsDownloaded(videoID string) bool {
	return ytd.indexer.IsDownloaded(videoID)
}

func (ytd *YouTubeDownloader) MarkDownloaded(videoID string) error {
	ytd.indexer.MarkDownloaded(videoID)
	return nil
}

func (ytd *YouTubeDownloader) downloadVideo(ctx context.Context, video *youtube.Video, filename, outputDir, resolution string) error {
	p := mpb.New(mpb.WithWidth(60), mpb.WithRefreshRate(180*time.Millisecond))

	format := ytd.selectBestFormat(video, resolution)
	if format == nil {
		return fmt.Errorf("未找到合适的视频格式")
	}

	totalBytes := format.ContentLength
	if totalBytes == 0 {
		totalBytes = 100 * 1024 * 1024
	}

	bar := p.AddBar(totalBytes,
		mpb.PrependDecorators(
			decor.Name(fmt.Sprintf("%-30s", utils.TruncateString(video.Title, 30))),
			decor.CountersKiloByte("% .2f / % .2f"),
		),
		mpb.AppendDecorators(
			decor.Percentage(decor.WCSyncSpace),
			decor.Name(" ["),
			decor.EwmaETA(decor.ET_STYLE_GO, 90),
			decor.Name("]"),
		),
	)

	stream, size, err := ytd.client.GetStreamContext(ctx, video, format)
	if err != nil {
		return fmt.Errorf("获取视频流失败: %w", err)
	}
	defer stream.Close()

	outputPath := filepath.Join(outputDir, filename)
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	reader := &ProgressReader{
		Reader: stream,
		Total:  size,
		OnProgress: func(current, total int64) {
			bar.SetCurrent(current)
		},
	}

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	bar.SetTotal(totalBytes, true)
	p.Wait()

	return nil
}

func (ytd *YouTubeDownloader) selectBestFormat(video *youtube.Video, resolution string) *youtube.Format {
	targetHeight := 720
	switch resolution {
	case "1080":
		targetHeight = 1080
	case "720":
		targetHeight = 720
	case "480":
		targetHeight = 480
	case "360":
		targetHeight = 360
	}

	var bestFormat *youtube.Format
	minDiff := int(^uint(0) >> 1)

	for _, format := range video.Formats {
		if format.AudioChannels == 0 {
			continue
		}

		height := format.Height
		if height == 0 {
			continue
		}

		diff := height - targetHeight
		if diff < 0 {
			diff = -diff
		}

		if diff < minDiff {
			minDiff = diff
			bestFormat = &format
		}
	}

	return bestFormat
}

// generateFilename 根据配置文件中的OutputTemplate生成文件名
func (ytd *YouTubeDownloader) generateFilename(video *youtube.Video, ext string) string {
	// 如果配置文件中没有设置输出模板，则使用默认模板
	template := ytd.config.OutputTemplate
	if template == "" {
		template = "%(title)s.%(ext)s"
	}

	// 替换模板变量
	result := template

	// 替换标题
	result = strings.ReplaceAll(result, "%(title)s", utils.SanitizeFilename(video.Title))

	// 替换ID
	result = strings.ReplaceAll(result, "%(id)s", video.ID)

	// 替换上传日期（使用当前日期作为替代）
	currentDate := time.Now().Format("20060102")
	result = strings.ReplaceAll(result, "%(upload_date)s", currentDate)

	// 替换扩展名
	result = strings.ReplaceAll(result, "%(ext)s", strings.TrimPrefix(ext, "."))

	return result
}

// generateMetaFile 生成包含视频标题、标签等信息的TXT文件
func (ytd *YouTubeDownloader) generateMetaFile(video *youtube.Video, outputDir, videoFilename string) error {
	// 生成Meta文件名
	metaFilename := strings.TrimSuffix(videoFilename, filepath.Ext(videoFilename)) + ".txt"
	metaPath := filepath.Join(outputDir, metaFilename)

	// 构建Meta文件内容
	content := video.Title

	// 添加标签（如果有）
	// 注意：github.com/kkdai/youtube/v2库可能不直接提供标签信息
	// 这里我们使用视频标题中的关键词作为标签的替代
	// 实际项目中可能需要使用其他方式获取标签信息
	content += "\n"
	content += "#视频 #YouTube"

	// 写入Meta文件
	if err := os.WriteFile(metaPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入Meta文件失败: %w", err)
	}

	log.Printf("生成Meta文件: %s", metaPath)
	return nil
}

// convertVideoFormat 使用ffmpeg进行视频格式转换
func (ytd *YouTubeDownloader) convertVideoFormat(inputPath, outputPath string) error {
	// 尝试使用当前目录下的ffmpeg.exe
	ffmpegPath := "./ffmpeg.exe"
	if _, err := os.Stat(ffmpegPath); os.IsNotExist(err) {
		// 如果当前目录不存在，则尝试使用系统PATH中的ffmpeg
		ffmpegPath = "ffmpeg"
	}

	// 构建ffmpeg命令
	cmd := exec.Command(ffmpegPath, "-i", inputPath, "-c", "copy", outputPath)

	// 执行命令
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg格式转换失败: %w", err)
	}

	log.Printf("视频格式转换完成: %s -> %s", inputPath, outputPath)
	return nil
}
