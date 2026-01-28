package downloader

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode"

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

	// 初始化随机数生成器
	rand.Seed(time.Now().UnixNano())

	// 设置代理
	if cfg.Proxy != "" {
		// 为 youtube.Client 设置代理
		// 注意：这里假设 youtube.Client 有一个 HTTPClient 字段，用于自定义 HTTP 客户端
		// 如果该字段不存在，可能需要修改底层库或使用其他方法设置代理
		// 由于无法直接查看外部库的代码，这里使用常见的实现方式
		client.HTTPClient = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(func() *url.URL {
					proxyURL, err := url.Parse(cfg.Proxy)
					if err != nil {
						log.Printf("代理URL解析失败: %v，将不使用代理", err)
						return nil
					}
					return proxyURL
				}()),
			},
		}
		log.Printf("已设置代理: %s", cfg.Proxy)
	}

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

	log.Printf("[YouTube下载器] 开始处理下载请求: %s", url)

	videoID, err := youtube.ExtractVideoID(url)
	if err != nil {
		log.Printf("[YouTube下载器] 提取视频ID失败: %v", err)
		return nil, fmt.Errorf("提取视频ID失败: %w", err)
	}
	log.Printf("[YouTube下载器] 提取到视频ID: %s", videoID)

	video, err := ytd.client.GetVideoContext(ctx, videoID)
	if err != nil {
		log.Printf("[YouTube下载器] 获取视频信息失败: %v", err)
		return nil, fmt.Errorf("获取视频失败: %w", err)
	}
	log.Printf("[YouTube下载器] 获取到视频信息: 标题='%s', 作者='%s', 时长=%d秒", video.Title, video.Author, int(video.Duration.Seconds()))

	if ytd.indexer.IsDownloaded(video.ID) {
		log.Printf("[YouTube下载器] 视频已下载，跳过: %s", video.Title)
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

	// 获取平台特定的输出目录
	platform := "youtube"
	platformOutputDir := ytd.config.GetPlatformOutputDir(platform)
	log.Printf("[YouTube下载器] 使用输出目录: %s", platformOutputDir)

	// 确保平台输出目录存在
	if err := os.MkdirAll(platformOutputDir, 0755); err != nil {
		log.Printf("[YouTube下载器] 创建输出目录失败: %v", err)
		return nil, fmt.Errorf("创建输出目录失败: %w", err)
	}
	log.Printf("[YouTube下载器] 输出目录已准备就绪: %s", platformOutputDir)

	// 生成符合OutputTemplate的文件名
	filename := ytd.generateFilename(video, ".mp4")
	outputPath := filepath.Join(platformOutputDir, filename)

	if err := utils.CleanupZeroByteFiles(outputPath); err != nil {
		log.Printf("清理0字节文件失败: %v", err)
	}

	log.Printf("开始下载: %s (ID: %s)", video.Title, video.ID)

	var lastErr error
	maxDelay := 30 * time.Second // 最大延迟上限

	for retry := 0; retry < ytd.config.MaxRetries; retry++ {
		if retry > 0 {
			// 指数退避算法：基础延迟 * (2^重试次数)
			baseDelay := ytd.config.BaseRetryDelay
			exponentialDelay := baseDelay * time.Duration(math.Pow(2, float64(retry-1)))

			// 添加随机抖动 (±20%)，避免多个请求同时重试导致的网络拥塞
			jitter := time.Duration(float64(exponentialDelay) * 0.2 * (2*rand.Float64() - 1))
			totalDelay := exponentialDelay + jitter

			// 确保延迟不超过最大值
			if totalDelay > maxDelay {
				totalDelay = maxDelay
			}

			// 确保延迟为正值
			if totalDelay < 0 {
				totalDelay = 0
			}

			log.Printf("重试 %d/%d，等待 %v 后重试...", retry, ytd.config.MaxRetries, totalDelay)
			time.Sleep(totalDelay)
		}

		err := ytd.downloadVideo(ctx, video, filename, platformOutputDir, resolution)
		if err != nil {
			lastErr = err
			log.Printf("下载失败 (尝试 %d/%d): %v", retry+1, ytd.config.MaxRetries, err)
			continue
		}

		ytd.indexer.MarkDownloaded(video.ID)

		// 处理Meta文件的生成
		if ytd.config.GenerateMetaFile {
			if err := ytd.generateMetaFile(video, platformOutputDir, filename); err != nil {
				log.Printf("生成Meta文件失败: %v", err)
			}
		}

		// 处理视频格式转换
		if ytd.config.RecodeVideo != "" {
			newFilename := strings.TrimSuffix(filename, filepath.Ext(filename)) + "." + ytd.config.RecodeVideo
			newOutputPath := filepath.Join(platformOutputDir, newFilename)
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
	// 创建进度条管理器，设置更高的刷新率
	p := mpb.New(
		mpb.WithWidth(80), // 增加进度条宽度
		mpb.WithRefreshRate(100*time.Millisecond), // 提高刷新率，使进度更实时
		mpb.WithOutput(os.Stderr),                 // 将进度条输出到标准错误
	)

	format := ytd.selectBestFormat(video, resolution)
	if format == nil {
		return fmt.Errorf("未找到合适的视频格式")
	}

	totalBytes := format.ContentLength
	if totalBytes == 0 {
		totalBytes = 100 * 1024 * 1024
	}

	// 创建进度条，添加更多装饰器
	bar := p.AddBar(totalBytes,
		mpb.PrependDecorators(
			decor.Name(fmt.Sprintf("%-40s", utils.TruncateString(video.Title, 40)), decor.WCSyncSpace),
			decor.CountersKiloByte("% .2f / % .2f"),
			decor.Name(" | "),
			decor.AverageSpeed(decor.UnitKiB, "% .2f"),
		),
		mpb.AppendDecorators(
			decor.Percentage(decor.WCSyncSpace),
			decor.Name(" ["),
			decor.EwmaETA(decor.ET_STYLE_GO, 60), // 使用更快的EWMA窗口，使ETA更准确
			decor.Name("]"),
		),
	)

	// 记录开始下载的日志
	log.Printf("开始下载视频: %s (ID: %s, 分辨率: %s, 格式: %s)",
		video.Title, video.ID, resolution, format.MimeType)

	stream, size, err := ytd.client.GetStreamContext(ctx, video, format)
	if err != nil {
		log.Printf("获取视频流失败: %v", err)
		return fmt.Errorf("获取视频流失败: %w", err)
	}
	defer stream.Close()

	outputPath := filepath.Join(outputDir, filename)
	file, err := os.Create(outputPath)
	if err != nil {
		log.Printf("创建文件失败: %v", err)
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 记录文件创建成功的日志
	log.Printf("创建输出文件: %s", outputPath)

	// 上次记录进度的时间和字节数
	lastLogTime := time.Now()
	lastLogBytes := int64(0)
	logInterval := 5 * time.Second // 日志记录间隔

	reader := &ProgressReader{
		Reader: stream,
		Total:  size,
		OnProgress: func(current, total int64) {
			// 更新进度条
			bar.SetCurrent(current)

			// 定期记录进度日志
			currentTime := time.Now()
			if currentTime.Sub(lastLogTime) >= logInterval {
				downloaded := current - lastLogBytes
				speed := float64(downloaded) / currentTime.Sub(lastLogTime).Seconds() / 1024 / 1024 // MB/s
				progress := float64(current) / float64(total) * 100

				log.Printf("下载进度: %s - %.2f%% (%.2f MB/%.2f MB, %.2f MB/s)",
					utils.TruncateString(video.Title, 30),
					progress,
					float64(current)/1024/1024,
					float64(total)/1024/1024,
					speed)

				lastLogTime = currentTime
				lastLogBytes = current
			}
		},
	}

	// 开始下载
	log.Printf("开始读取视频流，总大小: %.2f MB", float64(size)/1024/1024)

	if _, err := io.Copy(file, reader); err != nil {
		log.Printf("下载失败: %v", err)
		return fmt.Errorf("下载失败: %w", err)
	}

	// 完成进度条
	bar.SetTotal(totalBytes, true)
	p.Wait()

	// 记录下载完成的日志
	log.Printf("视频下载完成: %s (ID: %s, 保存路径: %s)",
		video.Title, video.ID, outputPath)

	return nil
}

func (ytd *YouTubeDownloader) selectBestFormat(video *youtube.Video, resolution string) *youtube.Format {
	// 解析目标分辨率
	targetHeight := 720
	resolutionMap := map[string]int{
		"2160": 2160, // 4K
		"1440": 1440, // 2K
		"1080": 1080, // Full HD
		"720":  720,  // HD
		"480":  480,  // SD
		"360":  360,  // Low SD
		"240":  240,  // Very Low
		"144":  144,  // Lowest
	}

	if h, ok := resolutionMap[resolution]; ok {
		targetHeight = h
	}

	// 按分辨率和质量排序格式
	formats := make([]*youtube.Format, 0, len(video.Formats))
	for i := range video.Formats {
		format := &video.Formats[i]
		// 只考虑包含音频的格式
		if format.AudioChannels == 0 {
			continue
		}
		// 确保高度有效
		if format.Height == 0 {
			continue
		}
		formats = append(formats, format)
	}

	// 如果没有找到合适的格式，返回nil
	if len(formats) == 0 {
		return nil
	}

	// 计算每个格式的得分
	type formatScore struct {
		format *youtube.Format
		score  int
	}

	var scoredFormats []formatScore
	for _, format := range formats {
		// 计算分辨率差异
		heightDiff := format.Height - targetHeight
		if heightDiff < 0 {
			heightDiff = -heightDiff
		}

		// 计算比特率得分（越高越好）
		bitrateScore := int(format.Bitrate / 1000) // 转换为kbps

		// 计算总得分：分辨率差异越小越好，比特率越高越好
		totalScore := bitrateScore - heightDiff*10

		scoredFormats = append(scoredFormats, formatScore{
			format: format,
			score:  totalScore,
		})
	}

	// 按得分排序，选择得分最高的格式
	bestFormat := scoredFormats[0].format
	bestScore := scoredFormats[0].score

	for _, sf := range scoredFormats {
		if sf.score > bestScore {
			bestScore = sf.score
			bestFormat = sf.format
		}
	}

	log.Printf("为视频 '%s' 选择最佳格式: 分辨率=%dp, 比特率=%dkbps, 格式=%s",
		utils.TruncateString(video.Title, 30),
		bestFormat.Height,
		bestFormat.Bitrate/1000,
		bestFormat.MimeType)

	return bestFormat
}

// generateFilename 根据配置文件中的OutputTemplate生成文件名
func (ytd *YouTubeDownloader) generateFilename(video *youtube.Video, ext string) string {
	// 如果配置文件中没有设置输出模板，则使用默认模板
	template := ytd.config.OutputTemplate
	if template == "" {
		template = "%(platform)s_%(content_type)s_%(title)s_%(id)s_%(timestamp)s.%(ext)s"
	}

	// 替换模板变量
	result := template

	// 替换平台信息
	platform := "youtube"
	result = strings.ReplaceAll(result, "%(platform)s", platform)

	// 替换内容类型
	contentType := "short"
	// 根据视频时长判断内容类型
	if video.Duration.Seconds() > 600 { // 10分钟以上为长视频
		contentType = "long"
	}
	result = strings.ReplaceAll(result, "%(content_type)s", contentType)

	// 替换标题
	title := utils.SanitizeFilename(video.Title)
	// 应用文件名最大长度限制
	if ytd.config.FilenameMaxLength > 0 {
		titleRunes := []rune(title)
		if len(titleRunes) > ytd.config.FilenameMaxLength {
			title = utils.TruncateString(title, ytd.config.FilenameMaxLength)
		}
	}
	result = strings.ReplaceAll(result, "%(title)s", title)

	// 替换播放列表名称（如果有）
	playlistName := ""
	result = strings.ReplaceAll(result, "%(playlist)s", playlistName)

	// 替换ID
	result = strings.ReplaceAll(result, "%(id)s", video.ID)

	// 替换作者
	author := utils.SanitizeFilename(video.Author)
	result = strings.ReplaceAll(result, "%(author)s", author)
	result = strings.ReplaceAll(result, "%(uploader)s", author)

	// 替换分辨率
	resolution := ""
	if video.Formats != nil && len(video.Formats) > 0 {
		bestFormat := ytd.selectBestFormat(video, ytd.config.DefaultResolution)
		if bestFormat != nil {
			resolution = fmt.Sprintf("%dp", bestFormat.Height)
			result = strings.ReplaceAll(result, "%(width)s", fmt.Sprintf("%d", bestFormat.Width))
			result = strings.ReplaceAll(result, "%(height)s", fmt.Sprintf("%d", bestFormat.Height))
		}
	}
	result = strings.ReplaceAll(result, "%(resolution)s", resolution)

	// 替换时长
	duration := fmt.Sprintf("%d", video.Duration)
	result = strings.ReplaceAll(result, "%(duration)s", duration)

	// 替换上传日期（使用当前日期作为替代）
	currentDate := time.Now()
	uploadDate := currentDate.Format("20060102")
	result = strings.ReplaceAll(result, "%(upload_date)s", uploadDate)

	// 替换下载日期和时间
	date := currentDate.Format("20060102")
	timeStr := currentDate.Format("150405")
	timestamp := currentDate.Format("20060102_150405")
	year := currentDate.Format("2006")
	month := currentDate.Format("01")
	day := currentDate.Format("02")

	result = strings.ReplaceAll(result, "%(date)s", date)
	result = strings.ReplaceAll(result, "%(time)s", timeStr)
	result = strings.ReplaceAll(result, "%(timestamp)s", timestamp)
	result = strings.ReplaceAll(result, "%(year)s", year)
	result = strings.ReplaceAll(result, "%(month)s", month)
	result = strings.ReplaceAll(result, "%(day)s", day)

	// 替换扩展名
	result = strings.ReplaceAll(result, "%(ext)s", strings.TrimPrefix(ext, "."))

	// 清理重复的下划线和点
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}
	for strings.Contains(result, "..") {
		result = strings.ReplaceAll(result, "..", ".")
	}

	// 对整个文件名应用长度限制
	if ytd.config.FilenameMaxLength > 0 {
		// 使用字符长度而不是字节长度
		runes := []rune(result)
		if len(runes) > ytd.config.FilenameMaxLength {
			// 直接对整个文件名进行截断，而不是只截断标题
			result = utils.TruncateString(result, ytd.config.FilenameMaxLength)
		}
	}

	return result
}

// generateMetaFile 生成包含视频标题、标签等信息的TXT文件
func (ytd *YouTubeDownloader) generateMetaFile(video *youtube.Video, outputDir, videoFilename string) error {
	// 生成Meta文件名
	metaFilename := strings.TrimSuffix(videoFilename, filepath.Ext(videoFilename)) + ".txt"
	metaPath := filepath.Join(outputDir, metaFilename)

	// 构建Meta文件内容
	var content strings.Builder

	// 添加标题
	content.WriteString(fmt.Sprintf("标题: %s\n", video.Title))

	// 添加作者
	content.WriteString(fmt.Sprintf("作者: %s\n", video.Author))

	// 添加视频ID
	content.WriteString(fmt.Sprintf("视频ID: %s\n", video.ID))

	// 添加视频URL
	content.WriteString(fmt.Sprintf("视频URL: https://www.youtube.com/watch?v=%s\n", video.ID))

	// 添加时长
	duration := time.Duration(video.Duration * time.Second)
	content.WriteString(fmt.Sprintf("时长: %s\n", duration))

	// 添加上传日期（如果有）
	// 注意：github.com/kkdai/youtube/v2库可能不直接提供上传日期
	// 这里使用当前日期作为替代
	content.WriteString(fmt.Sprintf("下载日期: %s\n", time.Now().Format("2006-01-02")))

	// 添加分辨率信息
	content.WriteString("分辨率: 自动选择最佳质量\n")

	// 添加标签
	content.WriteString("\n标签:\n")
	content.WriteString("#视频 #YouTube #" + strings.ReplaceAll(strings.ToLower(video.Author), " ", "#"))

	// 从标题中提取关键词作为标签
	titleWords := strings.Fields(video.Title)
	for _, word := range titleWords {
		// 只添加长度大于2的单词作为标签
		if len(word) > 2 {
			// 移除标点符号
			cleanWord := strings.Map(func(r rune) rune {
				if unicode.IsLetter(r) || unicode.IsNumber(r) {
					return r
				}
				return -1
			}, word)
			if len(cleanWord) > 2 {
				content.WriteString(" #" + strings.ToLower(cleanWord))
			}
		}
	}

	// 添加空行
	content.WriteString("\n\n")

	// 添加描述
	content.WriteString("描述:\n")
	content.WriteString(fmt.Sprintf("这是从YouTube下载的视频：%s\n", video.Title))
	content.WriteString(fmt.Sprintf("作者：%s\n", video.Author))
	content.WriteString(fmt.Sprintf("视频ID：%s\n", video.ID))

	// 写入Meta文件
	if err := os.WriteFile(metaPath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("写入Meta文件失败: %w", err)
	}

	log.Printf("生成Meta文件: %s", metaPath)
	return nil
}

// convertVideoFormat 使用ffmpeg进行视频格式转换
func (ytd *YouTubeDownloader) convertVideoFormat(inputPath, outputPath string) error {
	// 优先使用配置文件中的ffmpeg路径
	ffmpegPath := ytd.config.FfmpegPath
	if ffmpegPath == "" {
		// 尝试使用当前目录下的ffmpeg.exe
		ffmpegPath = "./ffmpeg.exe"
		if _, err := os.Stat(ffmpegPath); os.IsNotExist(err) {
			// 如果当前目录不存在，则尝试使用系统PATH中的ffmpeg
			ffmpegPath = "ffmpeg"
		}
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
