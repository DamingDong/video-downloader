package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	batchSize        = 10
	defaultOutputDir = "Output"
	resourceURLsDir  = "resource_urls"
	maxConcurrency   = 3
	timeoutPerVideo = 60 * time.Minute
	cookieFile       = "cookies.txt"
)

var (
	downloadedIndex = make(map[string]string)
	indexMutex     sync.RWMutex
)

type VideoInfo struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Duration int    `json:"duration"`
	Uploader string `json:"uploader"`
	WebpageURL string `json:"webpage_url"`
	Extractor string `json:"extractor_key"`
	Formats   []struct {
		FormatID string `json:"format_id"`
		Ext      string `json:"ext"`
		Height   int    `json:"height"`
		Width    int    `json:"width"`
		Filesize int64  `json:"filesize"`
		VCodec   string `json:"vcodec"`
		ACodec   string `json:"acodec"`
	} `json:"formats"`
}

func main() {
	var resolution string
	var filePath string

	if len(os.Args) > 1 {
		arg1 := strings.ToLower(os.Args[1])
		if arg1 == "720" || arg1 == "hd720" {
			resolution = "720"
			if len(os.Args) > 2 {
				filePath = os.Args[2]
			}
		} else if arg1 == "1080" || arg1 == "hd1080" {
			resolution = "1080"
			if len(os.Args) > 2 {
				filePath = os.Args[2]
			}
		} else if arg1 == "480" || arg1 == "medium" {
			resolution = "480"
			if len(os.Args) > 2 {
				filePath = os.Args[2]
			}
		} else if arg1 == "360" || arg1 == "small" {
			resolution = "360"
			if len(os.Args) > 2 {
				filePath = os.Args[2]
			}
		} else {
			filePath = os.Args[1]
			if len(os.Args) > 2 {
				arg2 := strings.ToLower(os.Args[2])
				if arg2 == "720" || arg2 == "hd720" {
					resolution = "720"
				} else if arg2 == "1080" || arg2 == "hd1080" {
					resolution = "1080"
				} else if arg2 == "480" || arg2 == "medium" {
					resolution = "480"
				} else if arg2 == "360" || arg2 == "small" {
					resolution = "360"
				}
			}
		}
	}

	if resolution == "" {
		resolution = "720"
	}

	log.Printf("使用分辨率: %sp", resolution)

	if err := checkYTDLP(); err != nil {
		log.Fatalf("检查 yt-dlp 失败: %v", err)
	}

	baseDir := filepath.Join(resourceURLsDir, defaultOutputDir)

	if err := initDownloadIndex(baseDir); err != nil {
		log.Printf("初始化索引失败: %v", err)
	}

	if filePath != "" {
		if err := processFromFile(filePath, resolution, baseDir); err != nil {
			log.Fatalf("处理文件失败: %v", err)
		}
	} else {
		if err := processFromDirectory(resolution, baseDir); err != nil {
			log.Fatalf("扫描目录失败: %v", err)
		}
	}
}

func getWebsiteType(url string) string {
	if strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be") {
		return "youtube"
	}
	if strings.Contains(url, "tiktok.com") {
		return "tiktok"
	}
	if strings.Contains(url, "douyin.com") {
		return "douyin"
	}
	if strings.Contains(url, "bilibili.com") {
		return "bilibili"
	}
	if strings.Contains(url, "vimeo.com") {
		return "vimeo"
	}
	if strings.Contains(url, "instagram.com") {
		return "instagram"
	}
	if strings.Contains(url, "twitter.com") || strings.Contains(url, "x.com") {
		return "twitter"
	}
	if strings.Contains(url, "facebook.com") {
		return "facebook"
	}
	if strings.Contains(url, "weibo.com") {
		return "weibo"
	}
	return "unknown"
}

func getCurrentMonthDir() string {
	return time.Now().Format("2006-01")
}

func getOutputDir(baseDir, website string) string {
	monthDir := getCurrentMonthDir()
	return filepath.Join(baseDir, website, monthDir)
}

func getQualityFormat(resolution string) string {
	switch resolution {
	case "1080", "hd1080":
		return "bestvideo[height<=1080]+bestaudio/best[height<=1080]"
	case "720", "hd720":
		return "bestvideo[height<=720]+bestaudio/best[height<=720]"
	case "480", "medium":
		return "bestvideo[height<=480]+bestaudio/best[height<=480]"
	case "360", "small":
		return "bestvideo[height<=360]+bestaudio/best[height<=360]"
	default:
		return "bestvideo+bestaudio/best"
	}
}

func checkYTDLP() error {
	cmd := exec.Command("yt-dlp", "--version")
	if err := cmd.Run(); err != nil {
		return errors.New("yt-dlp未安装，请先安装并配置到PATH中（多平台下载依赖）")
	}
	return nil
}

func cleanupZeroByteFiles(filePath string) error {
	if _, err := os.Stat(filePath); err == nil {
		info, err := os.Stat(filePath)
		if err != nil {
			return err
		}
		if info.Size() == 0 {
			log.Printf("发现0字节文件，删除: %s", filePath)
			return os.Remove(filePath)
		}
	}
	return nil
}

func parseVideoInfo(url string) (*VideoInfo, error) {
	cmd := exec.Command("yt-dlp",
		"--dump-json",
		"--no-playlist",
		url)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取视频信息失败: %w", err)
	}

	var info VideoInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("解析视频信息失败: %w", err)
	}

	return &info, nil
}

func downloadVideo(url, outputDir, resolution string) error {
	info, err := parseVideoInfo(url)
	if err != nil {
		return err
	}

	website := getWebsiteType(url)
	uniqueID := getUniqueID(url, info)

	indexMutex.RLock()
	_, exists := downloadedIndex[uniqueID]
	indexMutex.RUnlock()

	if exists {
		return fmt.Errorf("视频已下载: %s (ID: %s)", info.Title, uniqueID)
	}

	log.Printf("开始下载: %s (ID: %s, 网站: %s, 分辨率: %s)", info.Title, uniqueID, website, resolution)

	qualityFormat := getQualityFormat(resolution)
	outputTemplate := filepath.Join(outputDir, "%(title)s.%(ext)s")

	filename := sanitizeFilename(info.Title) + ".mp4"
	filePath := filepath.Join(outputDir, filename)

	if err := cleanupZeroByteFiles(filePath); err != nil {
		log.Printf("清理0字节文件失败: %v", err)
	}

	args := []string{
		"-f", qualityFormat,
		"-o", outputTemplate,
		"--no-playlist",
		"--no-warnings",
	}

	if _, err := os.Stat(cookieFile); err == nil {
		args = append(args, "--cookies", cookieFile)
		log.Printf("使用Cookie文件: %s", cookieFile)
	}

	args = append(args, url)

	cmd := exec.Command("yt-dlp", args...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	markVideoDownloaded(uniqueID, filePath)

	log.Printf("下载完成: %s (ID: %s)", info.Title, uniqueID)
	return nil
}

func getUniqueID(url string, info *VideoInfo) string {
	if info.ID != "" {
		return info.ID
	}
	return url
}

func sanitizeFilename(filename string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(filename)
}

func processFromFile(filePath, resolution, baseDir string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url == "" {
			continue
		}
		urls = append(urls, url)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	log.Printf("开始处理文件: %s (共 %d 个URL)", filePath, len(urls))

	return processURLs(urls, resolution, baseDir)
}

func processFromDirectory(resolution, baseDir string) error {
	log.Printf("开始扫描 %s 目录...", resourceURLsDir)

	entries, err := os.ReadDir(resourceURLsDir)
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
			urlFiles = append(urlFiles, filepath.Join(resourceURLsDir, entry.Name()))
		}
	}

	if len(urlFiles) == 0 {
		log.Println("未找到 URL 文件")
		return nil
	}

	log.Printf("找到 %d 个URL文件，开始批量处理...", len(urlFiles))

	for _, file := range urlFiles {
		if err := processFromFile(file, resolution, baseDir); err != nil {
			log.Printf("处理文件 %s 失败: %v", file, err)
		}
	}

	return nil
}

func processURLs(urls []string, resolution, baseDir string) error {
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var batchErr error
	errMutex := sync.Mutex{}
	batchNum := 1

	for i := 0; i < len(urls); i += batchSize {
		end := i + batchSize
		if end > len(urls) {
			end = len(urls)
		}
		batchURLs := urls[i:end]

		log.Printf("开始处理第 %d 批次 (共 %d 个视频，已加载%d个已下载索引)", batchNum, len(batchURLs), len(downloadedIndex))

		for idx := 0; idx < len(batchURLs); idx++ {
			url := strings.TrimSpace(batchURLs[idx])
			if url == "" {
				continue
			}

			website := getWebsiteType(url)
			outputDir := getOutputDir(baseDir, website)

			if err := os.MkdirAll(outputDir, 0755); err != nil {
				log.Printf("创建目录失败: %v", err)
				continue
			}

			sem <- struct{}{}
			wg.Add(1)
			go func(urlIdx int, url string) {
				defer func() {
					<-sem
					wg.Done()
				}()

				log.Printf("处理第 %d/%d 个URL: %s", urlIdx+1, len(batchURLs), url)

				if err := downloadVideo(url, outputDir, resolution); err != nil {
					if !strings.Contains(err.Error(), "视频已下载") {
						errMutex.Lock()
						if batchErr == nil {
							batchErr = fmt.Errorf("下载URL %s失败: %v", url, err)
						}
						errMutex.Unlock()
						log.Printf("下载失败 %s: %v", url, err)
					}
				}
			}(idx, url)
		}

		wg.Wait()

		if err := saveDownloadIndex(baseDir); err != nil {
			log.Printf("保存索引失败: %v", err)
		}

		if err := updateDownloadRecord(baseDir); err != nil {
			log.Printf("更新下载记录失败: %v", err)
		}

		if batchNum%5 == 0 {
			log.Printf("批次间隔休眠30秒（防限流）")
			time.Sleep(30 * time.Second)
		}
		batchNum++
	}

	return batchErr
}

func initDownloadIndex(baseDir string) error {
	indexPath := filepath.Join(baseDir, ".video_downloaded.index")
	file, err := os.Open(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("打开索引文件失败: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			continue
		}
		videoID := parts[0]
		storePath := parts[1]
		if _, err := os.Stat(storePath); err == nil {
			indexMutex.Lock()
			downloadedIndex[videoID] = storePath
			indexMutex.Unlock()
		}
	}
	return scanner.Err()
}

func saveDownloadIndex(baseDir string) error {
	indexPath := filepath.Join(baseDir, ".video_downloaded.index")
	file, err := os.Create(indexPath)
	if err != nil {
		return fmt.Errorf("创建索引文件失败: %w", err)
	}
	defer file.Close()

	indexMutex.RLock()
	defer indexMutex.RUnlock()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	_, err = writer.WriteString("# 视频下载索引: 视频ID|存储路径\n")
	if err != nil {
		return err
	}

	for vid, path := range downloadedIndex {
		_, err = writer.WriteString(fmt.Sprintf("%s|%s\n", vid, path))
		if err != nil {
			return err
		}
	}

	return nil
}

func markVideoDownloaded(videoID, storePath string) {
	indexMutex.Lock()
	defer indexMutex.Unlock()
	downloadedIndex[videoID] = storePath
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func updateDownloadRecord(baseDir string) error {
	recordFile := filepath.Join(baseDir, "下载记录.md")
	
	totalSize := int64(0)
	videoCount := 0
	var videos []struct {
		title     string
		videoID   string
		size      string
		downloadTime string
		path      string
	}
	
	indexMutex.RLock()
	for videoID, path := range downloadedIndex {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		
		totalSize += info.Size()
		videoCount++
		
		filename := filepath.Base(path)
		title := strings.TrimSuffix(filename, filepath.Ext(filename))
		
		size := formatFileSize(info.Size())
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
			title     string
			videoID   string
			size      string
			downloadTime string
			path      string
		}{
			title:     title,
			videoID:   videoID,
			size:      size,
			downloadTime: downloadTime,
			path:      relPath,
		})
	}
	indexMutex.RUnlock()
	
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
	content.WriteString(fmt.Sprintf("| 下载月份 | %s |\n", getCurrentMonthDir()))
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
	
	log.Printf("下载记录已更新: %s", recordFile)
	return nil
}
