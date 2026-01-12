package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"batch_download_videos/config"
	"batch_download_videos/downloader"
	"batch_download_videos/indexer"
	"batch_download_videos/logger"
	"batch_download_videos/utils"
)

func main() {
	resolution := flag.String("r", "", "视频分辨率 (360/480/720/1080)")
	filePath := flag.String("f", "", "URL文件路径")
	downloaderType := flag.String("d", "", "下载器类型 (youtube/multi)")
	configPath := flag.String("c", "", "配置文件路径 (默认: config.json)")
	logDir := flag.String("log", "logs", "日志目录")
	logLevel := flag.String("log-level", "info", "日志级别 (debug/info/warn/error)")
	help := flag.Bool("help", false, "显示帮助信息")
	version := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	if *help {
		printHelp()
		return
	}

	if *version {
		printVersion()
		return
	}

	level := parseLogLevel(*logLevel)
	if _, err := logger.InitLogger(*logDir, level); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		return
	}
	defer logger.GetLogger().Close()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.GetLogger().Error("加载配置失败: %v", err)
		return
	}

	if *resolution != "" {
		cfg.DefaultResolution = *resolution
	}
	if *downloaderType != "" {
		cfg.DefaultDownloader = *downloaderType
	}

	logger.GetLogger().Info("使用分辨率: %sp", cfg.DefaultResolution)
	logger.GetLogger().Info("使用下载器: %s", cfg.DefaultDownloader)

	outputDir := "Output"
	if err := utils.EnsureDir(outputDir); err != nil {
		logger.GetLogger().Error("创建输出目录失败: %v", err)
		return
	}

	idx := indexer.NewIndexer(outputDir)
	if err := idx.Load(); err != nil {
		logger.GetLogger().Warn("初始化索引失败: %v", err)
	}

	var dl downloader.Downloader
	switch strings.ToLower(cfg.DefaultDownloader) {
	case "youtube", "yt":
		dl = downloader.NewYouTubeDownloader(cfg, idx)
		logger.GetLogger().Info("使用 YouTube 专用下载器（性能优化）")
	case "multi", "all":
		dl = downloader.NewMultiPlatformDownloader(cfg, idx)
		if err := dl.(*downloader.MultiPlatformDownloader).CheckYTDLP(); err != nil {
			logger.GetLogger().Error("检查 yt-dlp 失败: %v", err)
			return
		}
		logger.GetLogger().Info("使用多平台下载器（支持9+平台）")
	case "auto":
		ytDL := downloader.NewYouTubeDownloader(cfg, idx)
		multiDL := downloader.NewMultiPlatformDownloader(cfg, idx)
		if err := multiDL.CheckYTDLP(); err != nil {
			logger.GetLogger().Error("检查 yt-dlp 失败: %v", err)
			return
		}
		dl = downloader.NewSmartDownloader(ytDL, multiDL)
		logger.GetLogger().Info("使用智能下载器（自动检测平台，YouTube用专用，其他用multi）")
	default:
		logger.GetLogger().Error("不支持的下载器类型: %s (支持: youtube/multi/auto)", cfg.DefaultDownloader)
		return
	}

	if *filePath != "" {
		if err := processFromFile(*filePath, cfg.DefaultResolution, dl, idx, cfg.MaxConcurrency); err != nil {
			logger.GetLogger().Error("处理文件失败: %v", err)
			return
		}
	} else {
		if err := processFromDirectory(cfg.DefaultResolution, dl, idx, cfg.MaxConcurrency); err != nil {
			logger.GetLogger().Error("扫描目录失败: %v", err)
			return
		}
	}

	if err := idx.Save(); err != nil {
		logger.GetLogger().Error("保存索引失败: %v", err)
	}

	if err := updateDownloadRecord(outputDir); err != nil {
		logger.GetLogger().Error("更新下载记录失败: %v", err)
	}

	logger.GetLogger().Info("所有任务完成！")
}

func parseLogLevel(level string) logger.LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return logger.DEBUG
	case "info":
		return logger.INFO
	case "warn":
		return logger.WARN
	case "error":
		return logger.ERROR
	default:
		return logger.INFO
	}
}

func processFromFile(filePath, resolution string, dl downloader.Downloader, idx *indexer.Indexer, maxConcurrency int) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url == "" || strings.HasPrefix(url, "#") {
			continue
		}
		urls = append(urls, url)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	logger.GetLogger().Info("开始处理文件: %s (共 %d 个URL)", filePath, len(urls))

	return processURLs(urls, resolution, dl, maxConcurrency)
}

func processFromDirectory(resolution string, dl downloader.Downloader, idx *indexer.Indexer, maxConcurrency int) error {
	logger.GetLogger().Info("开始扫描 %s 目录...", "resource_urls")

	entries, err := os.ReadDir("resource_urls")
	if err != nil {
		return fmt.Errorf("读取目录失败: %w", err)
	}

	var urlFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if strings.HasSuffix(name, ".txt") || strings.HasSuffix(name, ".url") || strings.HasSuffix(name, ".list") {
			urlFiles = append(urlFiles, filepath.Join("resource_urls", entry.Name()))
		}
	}

	if len(urlFiles) == 0 {
		logger.GetLogger().Warn("未找到 URL 文件")
		return nil
	}

	logger.GetLogger().Info("找到 %d 个 URL 文件", len(urlFiles))

	for _, file := range urlFiles {
		if err := processFromFile(file, resolution, dl, idx, maxConcurrency); err != nil {
			logger.GetLogger().Error("处理文件 %s 失败: %v", file, err)
		}
	}

	return nil
}

func processURLs(urls []string, resolution string, dl downloader.Downloader, maxConcurrency int) error {
	if maxConcurrency <= 0 {
		maxConcurrency = 3
	}

	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var errMutex sync.Mutex
	var batchErr error
	successCount := 0
	failCount := 0
	skipCount := 0

	logger.GetLogger().BatchStart(len(urls), maxConcurrency)

	for i, url := range urls {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(url string, idx int) {
			defer wg.Done()
			defer func() { <-semaphore }()

			logger.GetLogger().Debug("[%d/%d] 开始下载: %s", idx+1, len(urls), url)

			result, err := dl.Download(url, "Output", resolution)
			if err != nil {
				errMutex.Lock()
				failCount++
				batchErr = err
				errMutex.Unlock()
				logger.GetLogger().DownloadFail("", url, err, 0)
			} else if result != nil {
				if result.Success {
					errMutex.Lock()
					successCount++
					errMutex.Unlock()
					logger.GetLogger().DownloadSuccess(result.VideoID, result.Title, result.RetryCount, result.FileSize)
				} else {
					if result.Error != nil && result.Error.Error() == "视频已下载" {
						errMutex.Lock()
						skipCount++
						errMutex.Unlock()
						logger.GetLogger().DownloadSkip(result.VideoID, result.Title)
					} else {
						errMutex.Lock()
						failCount++
						errMutex.Unlock()
						logger.GetLogger().DownloadFail(result.VideoID, result.Title, result.Error, result.RetryCount)
					}
				}
			}
		}(url, i)
	}

	wg.Wait()

	logger.GetLogger().BatchComplete(successCount, failCount, skipCount, len(urls))

	return batchErr
}

func updateDownloadRecord(baseDir string) error {
	recordFile := filepath.Join(baseDir, "下载记录.md")

	totalSize := int64(0)
	videoCount := 0
	var videos []struct {
		title        string
		videoID      string
		size         string
		downloadTime string
		path         string
	}

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".mp4" && ext != ".mkv" && ext != ".avi" && ext != ".mov" && ext != ".m4v" {
			return nil
		}

		totalSize += info.Size()
		videoCount++

		filename := filepath.Base(path)
		title := strings.TrimSuffix(filename, filepath.Ext(filename))

		size := utils.FormatFileSize(info.Size())
		downloadTime := info.ModTime().Format("2006-01-02 15:04")

		relPath := strings.TrimPrefix(path, baseDir+"/")
		relPath = strings.TrimPrefix(relPath, "resource_urls/")
		if strings.HasPrefix(relPath, "/") || strings.Contains(relPath, ":") {
			parts := strings.Split(path, "resource_urls/")
			if len(parts) > 1 {
				relPath = "resource_urls/" + parts[1]
			} else {
				relPath = filepath.Base(path)
			}
		}

		videos = append(videos, struct {
			title        string
			videoID      string
			size         string
			downloadTime string
			path         string
		}{
			title:        title,
			videoID:      "-",
			size:         size,
			downloadTime: downloadTime,
			path:         relPath,
		})
		return nil
	})

	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("扫描文件失败: %w", err)
	}

	sort.Slice(videos, func(i, j int) bool {
		return videos[i].downloadTime < videos[j].downloadTime
	})

	var videoTable strings.Builder
	videoTable.WriteString("| 序号 | 视频标题 | 视频ID | 文件大小 | 下载时间 | 存储路径 | 状态 |\n")
	videoTable.WriteString("|------|----------|---------|----------|----------|-----------|------|\n")

	for i, v := range videos {
		videoTable.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s | %s | ✅ 已完成 |\n",
			i+1, v.title, v.videoID, v.size, v.downloadTime, v.path))
	}

	lastUpdate := time.Now().Format("2006-01-02 15:04:05")
	sizeGB := float64(totalSize) / (1024 * 1024 * 1024)

	var content strings.Builder
	content.WriteString("# 多平台视频下载记录\n\n")
	content.WriteString(fmt.Sprintf("> 最后更新：%s\n\n", lastUpdate))
	content.WriteString("## 📊 统计信息\n\n")
	content.WriteString("| 指标 | 数值 |\n")
	content.WriteString("|------|------|\n")
	content.WriteString(fmt.Sprintf("| 总视频数 | %d |\n", videoCount))
	content.WriteString(fmt.Sprintf("| 总大小 | %.2f GB |\n", sizeGB))
	content.WriteString(fmt.Sprintf("| 下载月份 | %s |\n", utils.GetCurrentMonthDir()))
	content.WriteString("| 视频来源 | 多平台 |\n\n")
	content.WriteString("## 📹 已下载视频列表\n\n")
	content.WriteString(videoTable.String())
	content.WriteString("\n## 📁 目录结构\n\n")
	content.WriteString("```\n")
	content.WriteString("Output/\n")
	content.WriteString("├── youtube/\n")
	content.WriteString("├── douyin/\n")
	content.WriteString("├── weibo/\n")
	content.WriteString("├── bilibili/\n")
	content.WriteString("├── tiktok/\n")
	content.WriteString("├── vimeo/\n")
	content.WriteString("├── instagram/\n")
	content.WriteString("├── twitter/\n")
	content.WriteString("├── facebook/\n")
	content.WriteString("└── unknown/\n")
	content.WriteString("```\n\n")
	content.WriteString("## 📝 说明\n\n")
	content.WriteString("- ✅ **已完成**：视频下载成功，文件完整\n")
	content.WriteString("- ⏳ **下载中**：视频正在下载\n")
	content.WriteString("- ❌ **失败**：下载失败，需要重试\n")
	content.WriteString("- ⚠️ **不完整**：文件大小异常或下载中断\n\n")
	content.WriteString("---\n\n")
	content.WriteString("*此文档由批量下载工具自动生成，记录所有已下载的视频信息*\n")

	if err := os.WriteFile(recordFile, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("写入下载记录失败: %w", err)
	}

	logger.GetLogger().Info("下载记录已更新: %s", recordFile)
	return nil
}

func printHelp() {
	fmt.Println("批量视频下载工具 - 混合架构")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  batch_download [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -r string")
	fmt.Println("        视频分辨率 (360/480/720/1080) (默认: 从配置文件读取)")
	fmt.Println("  -f string")
	fmt.Println("        URL文件路径")
	fmt.Println("  -d string")
	fmt.Println("        下载器类型 (youtube/multi/auto) (默认: auto)")
	fmt.Println("        youtube - YouTube 专用下载器（性能更好）")
	fmt.Println("        multi  - 多平台下载器（支持9+平台）")
	fmt.Println("        auto   - 自动检测（YouTube用专用，其他用multi）")
	fmt.Println("  -c string")
	fmt.Println("        配置文件路径 (默认: config.json)")
	fmt.Println("  -log string")
	fmt.Println("        日志目录 (默认: logs)")
	fmt.Println("  -log-level string")
	fmt.Println("        日志级别 (debug/info/warn/error) (默认: info)")
	fmt.Println("  -help")
	fmt.Println("        显示帮助信息")
	fmt.Println("  -version")
	fmt.Println("        显示版本信息")
	fmt.Println()
	fmt.Println("下载器说明:")
	fmt.Println("  youtube  - YouTube 专用下载器（使用 Go 库，性能更好，支持 YouTube Shorts）")
	fmt.Println("  multi    - 多平台下载器（使用 yt-dlp，支持9+平台）")
	fmt.Println("  auto     - 自动检测（YouTube用专用下载器，其他平台用multi）")
	fmt.Println()
	fmt.Println("配置文件 (config.json):")
	fmt.Println("  {")
	fmt.Println("    \"batch_size\": 10,")
	fmt.Println("    \"max_concurrency\": 3,")
	fmt.Println("    \"timeout_per_video\": \"1h0m0s\",")
	fmt.Println("    \"max_retries\": 3,")
	fmt.Println("    \"base_retry_delay\": \"2s\",")
	fmt.Println("    \"default_output_dir\": \"Output\",")
	fmt.Println("    \"resource_urls_dir\": \"resource_urls\",")
	fmt.Println("    \"cookie_file\": \"cookies.txt\",")
	fmt.Println("    \"index_file\": \".video_downloaded.index\",")
	fmt.Println("    \"record_file\": \"下载记录.md\",")
	fmt.Println("    \"default_resolution\": \"720\",")
	fmt.Println("    \"default_downloader\": \"auto\"")
	fmt.Println("  }")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  # 使用默认配置文件，扫描 resource_urls 目录")
	fmt.Println("  ./batch_download")
	fmt.Println()
	fmt.Println("  # 使用自定义配置文件")
	fmt.Println("  ./batch_download -c my_config.json")
	fmt.Println()
	fmt.Println("  # 命令行参数覆盖配置文件")
	fmt.Println("  ./batch_download -r 1080 -d youtube")
	fmt.Println()
	fmt.Println("  # 下载指定文件")
	fmt.Println("  ./batch_download -f resource_urls/example.txt")
	fmt.Println()
	fmt.Println("  # 启用调试日志")
	fmt.Println("  ./batch_download -log-level debug")
	fmt.Println()
	fmt.Println("支持的平台:")
	fmt.Println("  YouTube (含 Shorts), 抖音, 微博, Bilibili, TikTok, Vimeo, Instagram, Twitter, Facebook")
}

func printVersion() {
	fmt.Println("批量视频下载工具 v2.0.0")
	fmt.Println("混合架构 - 支持 YouTube 专用和多平台下载")
	fmt.Println("基于 yt-dlp 和 Go 实现")
}

var cfg *config.Config
