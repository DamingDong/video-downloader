package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kkdai/youtube/v2"
	"github.com/kkdai/youtube/v2/downloader"
)

// 配置项
const (
	batchSize         = 10                    // 每批次下载的视频数量
	defaultOutputDir  = "Output"              // 基础输出目录
	resourceURLsDir   = "resource_urls"       // URL资源目录
	qualityHighest    = "hd1080"              // 优先下载的最高分辨率
	fallbackQuality   = "hd720"               // 降级分辨率
	maxConcurrency    = 3                     // 单批次内并发下载数（避免被限流）
	timeoutPerVideo   = 60 * time.Minute     // 单个视频下载超时（大文件需要更长时间）
)

// 全局去重索引（视频ID -> 存储路径）
var (
	downloadedIndex = make(map[string]string)
	indexMutex      sync.RWMutex
	indexFile       = ".video_downloaded.index" // 持久化去重索引文件（通用，支持多平台）
	targetQuality   = ""                          // 用户指定的目标分辨率
)

// 从URL提取网站名称
func getWebsiteName(url string) string {
	if strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be") {
		return "youtube"
	}
	if strings.Contains(url, "tiktok.com") {
		return "tiktok"
	}
	if strings.Contains(url, "bilibili.com") {
		return "bilibili"
	}
	// 默认使用 unknown
	return "unknown"
}

// 获取当前月份目录（格式：2026-01）
func getCurrentMonthDir() string {
	return time.Now().Format("2006-01")
}

// 清理0字节文件
func cleanupZeroByteFiles(filePath string) error {
	if _, err := os.Stat(filePath); err == nil {
		// 文件存在，检查大小
		info, err := os.Stat(filePath)
		if err != nil {
			return err
		}
		if info.Size() == 0 {
			// 删除0字节文件
			log.Printf("发现0字节文件，删除: %s", filePath)
			return os.Remove(filePath)
		}
	}
	return nil
}

// 更新下载记录文档
func updateDownloadRecord(baseDir string) error {
	recordFile := filepath.Join(baseDir, "下载记录.md")
	
	// 计算统计信息
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
		// 获取文件信息
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		
		totalSize += info.Size()
		videoCount++
		
		// 从文件名提取标题
		filename := filepath.Base(path)
		title := strings.TrimSuffix(filename, filepath.Ext(filename))
		
		// 格式化文件大小
		size := formatFileSize(info.Size())
		
		// 格式化下载时间
		downloadTime := info.ModTime().Format("2006-01-02 15:04")
		
		// 获取相对路径
		relPath := strings.TrimPrefix(path, baseDir+"/")
		relPath = strings.TrimPrefix(relPath, "resource_urls/")
		// 确保使用相对路径
		if strings.HasPrefix(relPath, "/") || strings.Contains(relPath, ":") {
			// 如果是绝对路径，转换为相对路径
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
	
	// 按下载时间排序
	sort.Slice(videos, func(i, j int) bool {
		return videos[i].downloadTime < videos[j].downloadTime
	})
	
	// 生成视频列表表格
	var videoTable strings.Builder
	videoTable.WriteString("| 序号 | 视频标题 | 视频ID | 文件大小 | 下载时间 | 存储路径 | 状态 |\n")
	videoTable.WriteString("|------|----------|---------|----------|----------|-----------|------|\n")
	
	for i, v := range videos {
		videoTable.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s | %s | ✅ 已完成 |\n",
			i+1, v.title, v.videoID, v.size, v.downloadTime, v.path))
	}
	
	// 更新最后更新时间
	lastUpdate := time.Now().Format("2006-01-02 15:04:05")
	
	// 更新统计表格
	sizeGB := float64(totalSize) / (1024 * 1024 * 1024)
	
	// 完全重写文档
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
	
	// 写回文件
	if err := os.WriteFile(recordFile, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("写入下载记录失败: %w", err)
	}
	
	log.Printf("下载记录已更新: %s", recordFile)
	return nil
}

// 格式化文件大小
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

// 创建新的下载记录文档
func createDownloadRecord(baseDir string) error {
	recordFile := filepath.Join(baseDir, "下载记录.md")
	
	content := fmt.Sprintf("# 多平台视频下载记录\n\n> 最后更新：%s\n\n## 📊 统计信息\n\n| 指标 | 数值 |\n|------|------|\n| 总视频数 | 0 |\n| 总大小 | 0.00 GB |\n| 下载月份 | %s |\n| 视频来源 | 多平台 |\n\n## 📹 已下载视频列表\n\n| 序号 | 视频标题 | 视频ID | 文件大小 | 下载时间 | 存储路径 | 状态 |\n|------|----------|---------|----------|----------|-----------|------|\n\n## 📁 目录结构\n\n```\nOutput/\n├── youtube/\n├── bilibili/\n├── tiktok/\n├── vimeo/\n├── instagram/\n├── twitter/\n├── facebook/\n└── unknown/\n```\n\n## 📝 说明\n\n- ✅ **已完成**：视频下载成功，文件完整\n- ⏳ **下载中**：视频正在下载\n- ❌ **失败**：下载失败，需要重试\n- ⚠️ **不完整**：文件大小异常或下载中断\n\n---\n\n*此文档由批量下载工具自动生成，记录所有已下载的视频信息*\n", 
		time.Now().Format("2006-01-02 15:04:05"), getCurrentMonthDir())
	
	if err := os.WriteFile(recordFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("创建下载记录失败: %w", err)
	}
	
	log.Printf("下载记录已创建: %s", recordFile)
	return nil
}

// 初始化去重索引（从文件加载已下载的视频）
func initDownloadIndex(baseDir string) error {
	indexPath := filepath.Join(baseDir, indexFile)
	file, err := os.Open(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 首次运行无索引文件
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
		// 验证文件是否存在（防止索引文件残留但实际文件已删除）
		if _, err := os.Stat(storePath); err == nil {
			indexMutex.Lock()
			downloadedIndex[videoID] = storePath
			indexMutex.Unlock()
		} else {
			// 文件不存在，记录日志（可能是被移动或删除了）
			log.Printf("警告: 索引中的文件不存在，已移除记录: %s (ID: %s)", storePath, videoID)
		}
	}
	return scanner.Err()
}

// 保存去重索引到文件
func saveDownloadIndex(baseDir string) error {
	indexPath := filepath.Join(baseDir, indexFile)
	file, err := os.Create(indexPath)
	if err != nil {
		return fmt.Errorf("创建索引文件失败: %w", err)
	}
	defer file.Close()

	indexMutex.RLock()
	defer indexMutex.RUnlock()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// 写入注释
	_, err = writer.WriteString("# YouTube下载索引: 视频ID|存储路径\n")
	if err != nil {
		return err
	}

	// 写入已下载记录
	for vid, path := range downloadedIndex {
		_, err = writer.WriteString(fmt.Sprintf("%s|%s\n", vid, path))
		if err != nil {
			return err
		}
	}
	return nil
}

// 检查视频是否已下载
func isVideoDownloaded(videoID string) (bool, string) {
	indexMutex.RLock()
	defer indexMutex.RUnlock()
	path, exists := downloadedIndex[videoID]
	return exists, path
}

// 标记视频为已下载
func markVideoDownloaded(videoID, storePath string) {
	indexMutex.Lock()
	defer indexMutex.Unlock()
	downloadedIndex[videoID] = storePath
}

// 检查ffmpeg是否安装（高分辨率需要）
func checkFFmpeg() error {
	cmd := exec.Command("ffmpeg", "-version")
	if err := cmd.Run(); err != nil {
		return errors.New("ffmpeg未安装，请先安装并配置到PATH中（高分辨率下载依赖）")
	}
	return nil
}

// 提取视频ID并获取视频信息（带超时）
func getVideoInfo(client *youtube.Client, urlOrID string) (*youtube.Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 提取视频ID
	videoID, err := youtube.ExtractVideoID(urlOrID)
	if err != nil {
		return nil, fmt.Errorf("提取视频ID失败: %w", err)
	}

	// 检查是否已下载
	if exists, path := isVideoDownloaded(videoID); exists {
		return nil, fmt.Errorf("视频已下载: %s (路径: %s)", videoID, path)
	}

	// 获取视频信息
	video, err := client.GetVideoContext(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("获取视频信息失败: %w", err)
	}
	return video, nil
}

// 获取最高可用分辨率的格式（优化排序逻辑，优先选mp4编码）
func getHighestQualityFormat(video *youtube.Video) (*youtube.Format, string) {
	// 如果用户指定了目标分辨率，优先使用
	if targetQuality != "" {
		// 尝试获取指定分辨率的mp4格式
		mp4Formats := video.Formats.Quality(targetQuality).Type("mp4")
		if len(mp4Formats) > 0 {
			mp4Formats.Sort()
			return &mp4Formats[0], targetQuality
		}
		// 如果指定分辨率没有mp4，尝试其他格式
		formats := video.Formats.Quality(targetQuality)
		if len(formats) > 0 {
			formats.Sort()
			return &formats[0], targetQuality
		}
		log.Printf("警告: 未找到指定分辨率 %s，降级到其他分辨率", targetQuality)
	}

	// 优先筛选mp4编码的高分辨率
	mp4Formats1080 := video.Formats.Quality(qualityHighest).Type("mp4")
	if len(mp4Formats1080) > 0 {
		mp4Formats1080.Sort()
		return &mp4Formats1080[0], qualityHighest
	}

	// 降级到mp4编码的720p
	mp4Formats720 := video.Formats.Quality(fallbackQuality).Type("mp4")
	if len(mp4Formats720) > 0 {
		mp4Formats720.Sort()
		return &mp4Formats720[0], fallbackQuality
	}

	// 无mp4则选最高比特率的格式
	video.Formats.Sort()
	return &video.Formats[0], "auto"
}

// 从 MIME 类型获取文件扩展名
func getExtensionFromMime(mimeType string) string {
	if strings.Contains(mimeType, "mp4") {
		return ".mp4"
	}
	if strings.Contains(mimeType, "webm") {
		return ".webm"
	}
	if strings.Contains(mimeType, "3gp") {
		return ".3gp"
	}
	return ".mp4"
}

// 下载单个视频（带超时、去重、错误重试）
func downloadVideo(dl *downloader.Downloader, video *youtube.Video, outputDir string, url string) error {
	// 二次校验去重（防止并发场景下重复下载）
	if exists, path := isVideoDownloaded(video.ID); exists {
		log.Printf("跳过已下载视频: %s (ID: %s, 路径: %s)", video.Title, video.ID, path)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutPerVideo)
	defer cancel()

	format, quality := getHighestQualityFormat(video)

	// 清理文件名（兼容多系统）
	filename := downloader.SanitizeFilename(video.Title) + getExtensionFromMime(format.MimeType)
	outputPath := filepath.Join(outputDir, filename)

	// 下载前清理0字节文件
	if err := cleanupZeroByteFiles(outputPath); err != nil {
		log.Printf("清理0字节文件失败: %v", err)
	}

	log.Printf("开始下载: %s (ID: %s, 分辨率: %s)", video.Title, video.ID, quality)

	var err error
	// 高分辨率需要合并音视频
	if strings.HasPrefix(quality, "hd") {
		err = dl.DownloadComposite(ctx, filename, video, quality, "mp4", "")
	} else {
		err = dl.Download(ctx, video, format, filename)
	}

	if err != nil {
		// 重试一次（处理临时网络波动）
		log.Printf("首次下载失败，重试一次: %s, 错误: %v", video.Title, err)
		time.Sleep(2 * time.Second)
		if strings.HasPrefix(quality, "hd") {
			err = dl.DownloadComposite(ctx, filename, video, quality, "mp4", "")
		} else {
			err = dl.Download(ctx, video, format, filename)
		}
		if err != nil {
			return fmt.Errorf("下载失败: %w", err)
		}
	}

	// 标记为已下载并更新索引
	markVideoDownloaded(video.ID, outputPath)
	log.Printf("下载完成: %s (ID: %s)", video.Title, video.ID)
	return nil
}

// 批量处理URL列表（支持并发+去重）
func processBatch(urls []string, batchNum int, baseDir string) error {
	// 初始化客户端和下载器（自定义HTTP配置防限流）
	client := &youtube.Client{
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     90 * time.Second,
				DisableCompression:  false,
				TLSHandshakeTimeout: 20 * time.Second,
				ResponseHeaderTimeout: 60 * time.Second,
			},
		},
	}

	// 控制单批次内的并发数
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var batchErr error
	errMutex := sync.Mutex{}

	for idx, url := range urls {
		url = strings.TrimSpace(url)
		if url == "" {
			continue
		}

		sem <- struct{}{}
		wg.Add(1)
		go func(idx int, url string) {
			defer func() {
				<-sem
				wg.Done()
			}()

			log.Printf("处理第 %d/%d 个URL: %s", idx+1, len(urls), url)
			video, err := getVideoInfo(client, url)
			if err != nil {
				// 区分"已下载"和"真错误"
				if strings.Contains(err.Error(), "视频已下载") {
					log.Printf("跳过: %s", err.Error())
					return
				}
				errMutex.Lock()
				if batchErr == nil {
					batchErr = fmt.Errorf("处理URL %s失败: %v", url, err)
				}
				errMutex.Unlock()
				log.Printf("跳过URL %s: %v", url, err)
				return
			}

			// 根据URL创建输出目录（网站/月份）
			websiteName := getWebsiteName(url)
			monthDir := getCurrentMonthDir()
			outputDir := filepath.Join(baseDir, websiteName, monthDir)
			
			// 创建目录
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				errMutex.Lock()
				if batchErr == nil {
					batchErr = fmt.Errorf("创建输出目录失败: %w", err)
				}
				errMutex.Unlock()
				log.Printf("创建目录失败: %v", err)
				return
			}

			// 初始化下载器（每个URL使用不同的输出目录）
			dl := &downloader.Downloader{
				Client:    *client,
				OutputDir: outputDir,
			}

			if err := downloadVideo(dl, video, outputDir, url); err != nil {
				errMutex.Lock()
				if batchErr == nil {
					batchErr = fmt.Errorf("下载URL %s失败: %v", url, err)
				}
				errMutex.Unlock()
				log.Printf("下载失败 %s: %v", url, err)
			}
		}(idx, url)
	}

	wg.Wait()
	
	// 更新下载记录文档
	if err := updateDownloadRecord(baseDir); err != nil {
		log.Printf("更新下载记录失败: %v", err)
	}
	
	return batchErr
}

// 从文本文件读取URL并分批次处理
func processFromFile(filePath string, skipInit bool) error {
	// 检查文件是否存在
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 读取所有URL并去重（同一文件内重复的URL）
	urlSet := make(map[string]bool)
	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url == "" || !strings.Contains(url, "youtube") {
			continue
		}
		if !urlSet[url] {
			urlSet[url] = true
			urls = append(urls, url)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	if len(urls) == 0 {
		return errors.New("文件中未找到有效的YouTube URL")
	}

	// 确定输出基础目录（与文本文件同目录）
	fileDir := filepath.Dir(filePath)
	outputBaseDir := filepath.Join(fileDir, defaultOutputDir)
	if err := os.MkdirAll(outputBaseDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 初始化去重索引（如果需要）
	if !skipInit {
		if err := initDownloadIndex(outputBaseDir); err != nil {
			log.Printf("警告: 加载去重索引失败（仍可继续，但可能重复下载）: %v", err)
		}
		// 检查ffmpeg
		if err := checkFFmpeg(); err != nil {
			log.Printf("警告: %s (仅影响hd1080/hd720分辨率下载)", err)
		}
	}

	// 分批次处理
	batchNum := 1
	for i := 0; i < len(urls); i += batchSize {
		end := i + batchSize
		if end > len(urls) {
			end = len(urls)
		}
		batchURLs := urls[i:end]

		log.Printf("开始处理第 %d 批次 (共 %d 个视频，已加载%d个已下载索引)", batchNum, len(batchURLs), len(downloadedIndex))
		if err := processBatch(batchURLs, batchNum, outputBaseDir); err != nil {
			log.Printf("第 %d 批次处理存在错误: %v", batchNum, err)
		}

		// 每批次结束后保存索引（防止程序崩溃丢失）
		if err := saveDownloadIndex(outputBaseDir); err != nil {
			log.Printf("警告: 保存去重索引失败: %v", err)
		}

		// 批次间休眠（避免高频请求被YouTube限流）
		if batchNum%5 == 0 {
			log.Printf("批次间隔休眠30秒（防限流）")
			time.Sleep(30 * time.Second)
		}
		batchNum++
	}

	// 最终保存索引
	if !skipInit {
		if err := saveDownloadIndex(outputBaseDir); err != nil {
			log.Printf("最终保存索引失败: %v", err)
		}
	}

	log.Printf("文件 %s 处理完成！本文件总计已下载 %d 个视频", filepath.Base(filePath), len(downloadedIndex))
	return nil
}

// 扫描resource_urls目录中的所有URL文件
func scanResourceURLsDir(baseDir string) ([]string, error) {
	resourceDir := filepath.Join(baseDir, resourceURLsDir)
	
	// 检查目录是否存在
	if _, err := os.Stat(resourceDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("资源目录不存在: %s", resourceDir)
	}

	// 读取目录内容
	entries, err := os.ReadDir(resourceDir)
	if err != nil {
		return nil, fmt.Errorf("读取资源目录失败: %w", err)
	}

	var urlFiles []string
	for _, entry := range entries {
		// 跳过目录
		if entry.IsDir() {
			continue
		}
		
		// 检查文件扩展名
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".txt" || ext == ".url" || ext == ".list" {
			urlFiles = append(urlFiles, filepath.Join(resourceDir, name))
		}
	}

	if len(urlFiles) == 0 {
		return nil, fmt.Errorf("资源目录中未找到URL文件 (.txt/.url/.list)")
	}

	return urlFiles, nil
}

func main() {
	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("获取工作目录失败: %v", err)
	}

	// 解析命令行参数
	var filePath string
	if len(os.Args) >= 2 {
		// 检查第一个参数是否是分辨率（如 720, 1080, hd720, hd1080）
		arg1 := strings.ToLower(os.Args[1])
		if arg1 == "720" || arg1 == "hd720" {
			targetQuality = "hd720"
			log.Printf("使用分辨率: 720p")
		} else if arg1 == "1080" || arg1 == "hd1080" {
			targetQuality = "hd1080"
			log.Printf("使用分辨率: 1080p")
		} else if arg1 == "480" || arg1 == "medium" {
			targetQuality = "medium"
			log.Printf("使用分辨率: 480p")
		} else if arg1 == "360" || arg1 == "small" {
			targetQuality = "small"
			log.Printf("使用分辨率: 360p")
		} else {
			// 不是分辨率参数，可能是文件路径
			filePath = os.Args[1]
		}
	}

	// 检查第二个参数是否是分辨率
	if len(os.Args) >= 3 && filePath != "" {
		arg2 := strings.ToLower(os.Args[2])
		if arg2 == "720" || arg2 == "hd720" {
			targetQuality = "hd720"
			log.Printf("使用分辨率: 720p")
		} else if arg2 == "1080" || arg2 == "hd1080" {
			targetQuality = "hd1080"
			log.Printf("使用分辨率: 1080p")
		} else if arg2 == "480" || arg2 == "medium" {
			targetQuality = "medium"
			log.Printf("使用分辨率: 480p")
		} else if arg2 == "360" || arg2 == "small" {
			targetQuality = "small"
			log.Printf("使用分辨率: 360p")
		}
	}

	// 如果没有指定分辨率，显示默认分辨率
	if targetQuality == "" {
		log.Printf("使用默认分辨率: 1080p")
	}

	// 如果指定了文件路径，处理单个文件
	if filePath != "" {
		if err := processFromFile(filePath, false); err != nil {
			log.Fatalf("批量处理失败: %v", err)
		}
		return
	}

	// 无参数时，自动扫描resource_urls目录
	log.Printf("开始扫描 %s 目录...", resourceURLsDir)
	urlFiles, err := scanResourceURLsDir(workDir)
	if err != nil {
		log.Fatalf("扫描资源目录失败: %v", err)
	}

	log.Printf("找到 %d 个URL文件，开始批量处理...", len(urlFiles))

	// 确定输出基础目录
	outputBaseDir := filepath.Join(workDir, defaultOutputDir)
	if err := os.MkdirAll(outputBaseDir, 0755); err != nil {
		log.Fatalf("创建输出目录失败: %v", err)
	}

	// 初始化去重索引（全局初始化一次）
	if err := initDownloadIndex(outputBaseDir); err != nil {
		log.Printf("警告: 加载去重索引失败（仍可继续，但可能重复下载）: %v", err)
	}

	// 检查ffmpeg
	if err := checkFFmpeg(); err != nil {
		log.Printf("警告: %s (仅影响hd1080/hd720分辨率下载)", err)
	}

	// 处理每个URL文件
	successCount := 0
	failCount := 0
	for idx, urlFile := range urlFiles {
		log.Printf("\n========== 处理文件 %d/%d: %s ==========", idx+1, len(urlFiles), filepath.Base(urlFile))
		
		// 第一个文件需要初始化，后续文件跳过初始化
		skipInit := (idx > 0)
		if err := processFromFile(urlFile, skipInit); err != nil {
			log.Printf("处理文件失败: %v", err)
			failCount++
		} else {
			successCount++
		}
	}

	// 最终保存索引
	if err := saveDownloadIndex(outputBaseDir); err != nil {
		log.Printf("最终保存索引失败: %v", err)
	}

	log.Printf("\n========== 批量处理完成 ==========")
	log.Printf("成功处理: %d 个文件", successCount)
	log.Printf("失败处理: %d 个文件", failCount)
	log.Printf("总计已下载视频: %d 个", len(downloadedIndex))
}