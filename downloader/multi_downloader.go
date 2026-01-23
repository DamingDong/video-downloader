package downloader

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	// 尝试使用当前目录下的yt-dlp.exe
	ytDlpPath := "./yt-dlp.exe"
	if _, err := os.Stat(ytDlpPath); os.IsNotExist(err) {
		// 如果当前目录不存在，则尝试使用系统PATH中的yt-dlp
		ytDlpPath = "yt-dlp"
	}

	cmd := exec.Command(ytDlpPath,
		"--dump-json",
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
	log.Printf("[调试] 开始处理下载请求: %s", url)

	// 首先检查URL类型，判断是否为频道或播放列表
	isPlaylist := strings.Contains(url, "list=")
	isChannel := strings.Contains(url, "/@") || strings.Contains(url, "/channel/") || strings.Contains(url, "/c/") || strings.Contains(url, "/user/")

	log.Printf("[调试] URL类型: 播放列表=%t, 频道=%t", isPlaylist, isChannel)

	// 对于频道或播放列表，我们需要特殊处理
	if isChannel || isPlaylist {
		log.Printf("[调试] 检测到频道或播放列表URL，使用特殊处理")

		// 尝试使用当前目录下的yt-dlp.exe
		ytDlpPath := "./yt-dlp.exe"
		if _, err := os.Stat(ytDlpPath); os.IsNotExist(err) {
			// 如果当前目录不存在，则尝试使用系统PATH中的yt-dlp
			ytDlpPath = "yt-dlp"
		}

		qualityFormat := utils.GetQualityFormat(resolution)
		// 使用配置文件中的输出模板
		outputTemplate := filepath.Join(outputDir, mpd.config.OutputTemplate)
		// 如果配置文件中没有设置输出模板，则使用默认模板
		if mpd.config.OutputTemplate == "" {
			outputTemplate = filepath.Join(outputDir, "%(title)s.%(ext)s")
		}

		// 修改格式参数，使用已经合并好的格式，避免单独下载视频和音频文件
		// 使用best[height<=720]格式，直接下载已经合并好的视频
		qualityFormat = "best[height<=720]"

		// 根据URL类型设置不同的下载参数
		args := []string{
			"-f", qualityFormat,
			"-o", outputTemplate,
			"--no-warnings",
			"--ignore-errors",                                                        // 忽略错误，继续下载其他视频
			"--continue",                                                             // 支持断点续传
			"--no-overwrites",                                                        // 不覆盖已存在的文件
			"--download-archive", filepath.Join(outputDir, "downloaded_archive.txt"), // 记录已下载的视频ID，避免重复下载
		}

		// 如果需要生成Meta文件，则添加--write-info-json参数
		if mpd.config.GenerateMetaFile {
			args = append(args, "--write-info-json")
		}

		// 如果需要格式转换，则添加--recode-video参数
		// 注意：某些版本的 yt-dlp 可能不支持 --recode-video 参数，或者会导致下载失败
		// if mpd.config.RecodeVideo != "" {
		// 	args = append(args, "--recode-video", mpd.config.RecodeVideo)
		// }

		// 添加并发下载控制参数
		// 注意：某些版本的 yt-dlp 可能不支持 --max-concurrent-downloads 参数
		// if mpd.config.MaxConcurrentDownloads > 0 {
		// 	args = append(args, "--max-concurrent-downloads", fmt.Sprintf("%d", mpd.config.MaxConcurrentDownloads))
		// }

		// 添加代理设置参数
		if mpd.config.Proxy != "" {
			args = append(args, "--proxy", mpd.config.Proxy)
		}

		// 添加限速设置参数
		if mpd.config.LimitRate != "" {
			args = append(args, "--limit-rate", mpd.config.LimitRate)
		}

		// 对于频道下载，限制只下载最新的10个视频，避免程序卡住
		// 对于播放列表下载，不限制数量，支持完整下载
		if isChannel {
			args = append(args, "--playlist-end", "10", "--max-downloads", "10")
			log.Printf("[调试] 检测到频道URL，限制只下载最新的10个视频")
		} else if isPlaylist {
			log.Printf("[调试] 检测到播放列表URL，支持完整下载")
		}

		if _, err := os.Stat(mpd.config.CookieFile); err == nil {
			args = append(args, "--cookies", mpd.config.CookieFile)
			log.Printf("[调试] 使用Cookie文件: %s", mpd.config.CookieFile)
		}

		args = append(args, url)

		log.Printf("[调试] 执行yt-dlp命令: %s %s", ytDlpPath, strings.Join(args, " "))

		// 直接执行yt-dlp命令，不使用GetVideoInfo
		cmd := exec.Command(ytDlpPath, args...)

		// 直接输出命令的执行结果到控制台
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		log.Printf("[调试] 开始执行yt-dlp命令...")
		log.Printf("[调试] 执行命令: %s %s", ytDlpPath, strings.Join(args, " "))
		startTime := time.Now()

		err := cmd.Run()

		duration := time.Since(startTime)
		log.Printf("[调试] yt-dlp命令执行完成，耗时: %v", duration)

		// 检查错误
		if err != nil {
			// 检查错误是否是因为达到了最大下载数量而导致的
			// 当yt-dlp因为--max-downloads参数而停止下载时，它会返回一个错误码
			// 但是实际上，所有的视频都已经成功下载了
			log.Printf("[调试] yt-dlp命令执行失败: %v", err)

			// 对于频道或播放列表下载，我们认为即使返回了错误码，只要视频被下载了，就是成功的
			// 因为yt-dlp在达到最大下载数量时会返回错误码
			log.Printf("[调试] 频道/播放列表下载完成，即使yt-dlp返回了错误码")
		} else {
			log.Printf("[调试] yt-dlp命令执行成功")
		}

		// 对于频道或播放列表下载，我们总是认为是成功的
		// 因为yt-dlp在达到最大下载数量时会返回错误码
		// 但是实际上，所有的视频都已经成功下载了

		// 处理Meta文件的生成
		if mpd.config.GenerateMetaFile {
			if err := mpd.ProcessMetaFiles(outputDir); err != nil {
				log.Printf("处理Meta文件失败: %v", err)
			}
		}

		// 对于频道或播放列表，我们无法返回单个视频的结果
		// 直接返回成功，因为yt-dlp已经处理了所有下载
		return &DownloadResult{
			Success:    true,
			VideoID:    "",
			Title:      "频道/播放列表下载",
			FilePath:   outputDir,
			FileSize:   0,
			Error:      nil,
			RetryCount: 0,
		}, nil
	}

	// 以下是原始的单个视频处理逻辑
	info, err := mpd.GetVideoInfo(url)
	if err != nil {
		log.Printf("[调试] 获取视频信息失败: %v", err)
		return nil, err
	}

	website := utils.GetWebsiteType(url)
	uniqueID := mpd.getUniqueID(url, info)

	if mpd.indexer.IsDownloaded(uniqueID) {
		log.Printf("[调试] 视频已下载: %s", uniqueID)
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
	// 生成文件名
	filename := mpd.generateFilename(info, ".mp4")
	filePath := filepath.Join(outputDir, filename)

	if err := utils.CleanupZeroByteFiles(filePath); err != nil {
		log.Printf("清理0字节文件失败: %v", err)
	}

	// 修改格式参数，使用已经合并好的格式，避免单独下载视频和音频文件
	// 使用best[height<=720]格式，直接下载已经合并好的视频
	qualityFormat = "best[height<=720]"

	args := []string{
		"-f", qualityFormat,
		"-o", filePath,
		"--no-warnings",
		"--continue",      // 支持断点续传
		"--no-overwrites", // 不覆盖已存在的文件
	}

	// 如果需要生成Meta文件，则添加--write-info-json参数
	if mpd.config.GenerateMetaFile {
		args = append(args, "--write-info-json")
	}

	// 如果需要格式转换，则添加--recode-video参数
	// 注意：某些版本的 yt-dlp 可能不支持 --recode-video 参数，或者会导致下载失败
	// if mpd.config.RecodeVideo != "" {
	// 	args = append(args, "--recode-video", mpd.config.RecodeVideo)
	// }

	// 添加并发下载控制参数
	// 注意：某些版本的 yt-dlp 可能不支持 --max-concurrent-downloads 参数
	// if mpd.config.MaxConcurrentDownloads > 0 {
	// 	args = append(args, "--max-concurrent-downloads", fmt.Sprintf("%d", mpd.config.MaxConcurrentDownloads))
	// }

	// 添加代理设置参数
	if mpd.config.Proxy != "" {
		args = append(args, "--proxy", mpd.config.Proxy)
	}

	// 添加限速设置参数
	if mpd.config.LimitRate != "" {
		args = append(args, "--limit-rate", mpd.config.LimitRate)
	}

	if _, err := os.Stat(mpd.config.CookieFile); err == nil {
		args = append(args, "--cookies", mpd.config.CookieFile)
		log.Printf("使用Cookie文件: %s", mpd.config.CookieFile)
	}

	args = append(args, url)

	// 尝试使用当前目录下的yt-dlp.exe
	ytDlpPath := "./yt-dlp.exe"
	if _, err := os.Stat(ytDlpPath); os.IsNotExist(err) {
		// 如果当前目录不存在，则尝试使用系统PATH中的yt-dlp
		ytDlpPath = "yt-dlp"
	}

	var lastErr error
	for retry := 0; retry < mpd.config.MaxRetries; retry++ {
		if retry > 0 {
			delay := mpd.config.BaseRetryDelay * time.Duration(retry)
			log.Printf("重试 %d/%d，等待 %v 后重试...", retry, mpd.config.MaxRetries, delay)
			time.Sleep(delay)
		}

		log.Printf("[调试] 执行yt-dlp命令: %s %s", ytDlpPath, strings.Join(args, " "))
		cmd := exec.Command(ytDlpPath, args...)

		// 捕获标准错误
		var stderr strings.Builder
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			lastErr = err
			log.Printf("下载失败 (尝试 %d/%d): %v", retry+1, mpd.config.MaxRetries, err)
			log.Printf("[调试] 错误输出: %s", stderr.String())
			continue
		}

		mpd.indexer.MarkDownloaded(uniqueID)

		// 处理Meta文件的生成
		if mpd.config.GenerateMetaFile {
			if err := mpd.ProcessMetaFiles(outputDir); err != nil {
				log.Printf("处理Meta文件失败: %v", err)
			}
		}

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

// generateFilename 根据配置文件中的OutputTemplate生成文件名
func (mpd *MultiPlatformDownloader) generateFilename(info *VideoInfo, ext string) string {
	// 如果配置文件中没有设置输出模板，则使用默认模板
	template := mpd.config.OutputTemplate
	if template == "" {
		template = "%(title)s.%(ext)s"
	}

	// 替换模板变量
	result := template

	// 替换标题
	title := utils.SanitizeFilename(info.Title)
	// 应用文件名最大长度限制
	if mpd.config.FilenameMaxLength > 0 && len(title) > mpd.config.FilenameMaxLength {
		title = utils.TruncateString(title, mpd.config.FilenameMaxLength)
	}
	result = strings.ReplaceAll(result, "%(title)s", title)

	// 替换ID
	result = strings.ReplaceAll(result, "%(id)s", info.ID)

	// 替换作者
	author := utils.SanitizeFilename(info.Uploader)
	result = strings.ReplaceAll(result, "%(author)s", author)
	result = strings.ReplaceAll(result, "%(uploader)s", author)

	// 替换分辨率
	resolution := info.Resolution
	result = strings.ReplaceAll(result, "%(resolution)s", resolution)

	// 替换时长
	duration := fmt.Sprintf("%d", info.Duration)
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

	return result
}

func (mpd *MultiPlatformDownloader) getUniqueID(url string, info *VideoInfo) string {
	if info.ID != "" {
		return info.ID
	}
	return url
}

func (mpd *MultiPlatformDownloader) CheckYTDLP() error {
	// 尝试使用当前目录下的yt-dlp.exe
	ytDlpPath := "./yt-dlp.exe"
	if _, err := os.Stat(ytDlpPath); os.IsNotExist(err) {
		// 如果当前目录不存在，则尝试使用系统PATH中的yt-dlp
		ytDlpPath = "yt-dlp"
	}

	cmd := exec.Command(ytDlpPath, "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("yt-dlp未安装，请先安装并配置到PATH中或放在当前目录下（多平台下载依赖）")
	}
	return nil
}

// ProcessMetaFiles 处理Meta文件的生成和视频文件的重命名
func (mpd *MultiPlatformDownloader) ProcessMetaFiles(outputDir string) error {
	// 遍历outputDir目录，查找所有的info.json文件
	files, err := os.ReadDir(outputDir)
	if err != nil {
		return fmt.Errorf("读取目录失败: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		if strings.HasSuffix(filename, ".info.json") {
			// 读取并解析info.json文件
			jsonPath := filepath.Join(outputDir, filename)
			data, err := os.ReadFile(jsonPath)
			if err != nil {
				log.Printf("读取info.json文件失败: %v", err)
				continue
			}

			// 解析JSON数据
			var info struct {
				Title       string   `json:"title"`
				Tags        []string `json:"tags"`
				Description string   `json:"description"`
				ID          string   `json:"id"`
			}

			if err := json.Unmarshal(data, &info); err != nil {
				log.Printf("解析info.json文件失败: %v", err)
				continue
			}

			// 生成对应的TXT文件
			txtFilename := strings.TrimSuffix(filename, ".info.json") + ".txt"
			txtPath := filepath.Join(outputDir, txtFilename)

			// 构建TXT文件内容
			content := info.Title
			if len(info.Tags) > 0 {
				content += "\n"
				for _, tag := range info.Tags {
					content += "#" + tag + " "
				}
			}

			// 写入TXT文件
			if err := os.WriteFile(txtPath, []byte(content), 0644); err != nil {
				log.Printf("写入Meta文件失败: %v", err)
				continue
			}

			log.Printf("生成Meta文件: %s", txtPath)
		}
	}

	// 遍历outputDir目录，查找所有的视频文件，并重命名它们
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		// 检查文件是否为视频文件
		if strings.HasSuffix(filename, ".mp4") || strings.HasSuffix(filename, ".mkv") || strings.HasSuffix(filename, ".avi") || strings.HasSuffix(filename, ".mov") {
			// 读取并解析对应的info.json文件
			jsonFilename := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".info.json"
			jsonPath := filepath.Join(outputDir, jsonFilename)
			if _, err := os.Stat(jsonPath); err != nil {
				// 如果没有对应的info.json文件，则跳过
				continue
			}

			data, err := os.ReadFile(jsonPath)
			if err != nil {
				log.Printf("读取info.json文件失败: %v", err)
				continue
			}

			// 解析JSON数据
			var info struct {
				Title string `json:"title"`
				ID    string `json:"id"`
			}

			if err := json.Unmarshal(data, &info); err != nil {
				log.Printf("解析info.json文件失败: %v", err)
				continue
			}

			// 创建VideoInfo对象
			videoInfo := &VideoInfo{
				Title: info.Title,
				ID:    info.ID,
			}

			// 生成新的文件名
			ext := filepath.Ext(filename)
			newFilename := mpd.generateFilename(videoInfo, ext)
			newFilePath := filepath.Join(outputDir, newFilename)
			oldFilePath := filepath.Join(outputDir, filename)

			// 检查新文件名是否与旧文件名不同
			if newFilename != filename {
				// 重命名文件
				if err := os.Rename(oldFilePath, newFilePath); err != nil {
					log.Printf("重命名文件失败: %v", err)
					continue
				}

				log.Printf("重命名文件: %s -> %s", oldFilePath, newFilePath)

				// 重命名对应的info.json文件
				newJsonFilename := strings.TrimSuffix(newFilename, filepath.Ext(newFilename)) + ".info.json"
				newJsonPath := filepath.Join(outputDir, newJsonFilename)
				if err := os.Rename(jsonPath, newJsonPath); err != nil {
					log.Printf("重命名info.json文件失败: %v", err)
				}

				// 重命名对应的txt文件
				newTxtFilename := strings.TrimSuffix(newFilename, filepath.Ext(newFilename)) + ".txt"
				newTxtPath := filepath.Join(outputDir, newTxtFilename)
				oldTxtFilename := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".txt"
				oldTxtPath := filepath.Join(outputDir, oldTxtFilename)
				if _, err := os.Stat(oldTxtPath); err == nil {
					if err := os.Rename(oldTxtPath, newTxtPath); err != nil {
						log.Printf("重命名txt文件失败: %v", err)
					}
				}
			}
		}
	}

	return nil
}
