package utils

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func GetWebsiteType(url string) string {
	if strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be") {
		return "youtube"
	}
	if strings.Contains(url, "douyin.com") {
		return "douyin"
	}
	if strings.Contains(url, "weibo.com") {
		return "weibo"
	}
	if strings.Contains(url, "bilibili.com") {
		return "bilibili"
	}
	if strings.Contains(url, "tiktok.com") {
		return "tiktok"
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
	return "unknown"
}

func GetQualityFormat(resolution string) string {
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
		return "bestvideo[height<=720]+bestaudio/best[height<=720]"
	}
}

func GetCurrentMonthDir() string {
	return time.Now().Format("2006-01")
}

func SanitizeFilename(filename string) string {
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

func FormatFileSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(1024), 0
	for n := bytes / 1024; n >= 1024; n /= 1024 {
		div *= 1024
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func CleanupZeroByteFiles(filePath string) error {
	if _, err := os.Stat(filePath); err == nil {
		info, err := os.Stat(filePath)
		if err != nil {
			return err
		}
		if info.Size() == 0 {
			return os.Remove(filePath)
		}
	}
	return nil
}

func GetOutputDir(baseDir, defaultDir string) string {
	if baseDir == "" {
		baseDir = defaultDir
	}
	return baseDir
}

func EnsureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

func TruncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}

// ValidateURL 验证URL是否有效
func ValidateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL不能为空")
	}

	// 检查URL格式
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("URL格式无效: %w", err)
	}

	// 检查URL是否包含协议
	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL缺少协议 (http:// 或 https://)")
	}

	// 检查URL是否包含主机名
	if parsedURL.Host == "" {
		return fmt.Errorf("URL缺少主机名")
	}

	return nil
}

// ValidateURLs 批量验证URL列表
func ValidateURLs(urls []string) ([]string, []error) {
	validURLs := []string{}
	errors := []error{}

	for i, urlStr := range urls {
		urlStr = strings.TrimSpace(urlStr)
		if urlStr == "" || strings.HasPrefix(urlStr, "#") {
			continue
		}

		if err := ValidateURL(urlStr); err != nil {
			errors = append(errors, fmt.Errorf("URL %d: %w", i+1, err))
		} else {
			validURLs = append(validURLs, urlStr)
		}
	}

	return validURLs, errors
}

// CleanupTempFiles 清理临时文件，特别是无后缀名的文件
func CleanupTempFiles(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("读取目录失败: %w", err)
	}

	removedCount := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		// 清理无后缀名的文件
		if filepath.Ext(filename) == "" {
			filePath := filepath.Join(dir, filename)
			if err := os.Remove(filePath); err != nil {
				// 忽略删除错误
				continue
			}
			removedCount++
		}
		// 清理其他常见临时文件
		if strings.HasSuffix(filename, ".part") || strings.HasSuffix(filename, ".ytdl") {
			filePath := filepath.Join(dir, filename)
			if err := os.Remove(filePath); err != nil {
				// 忽略删除错误
				continue
			}
			removedCount++
		}
	}

	if removedCount > 0 {
		fmt.Printf("清理了 %d 个临时文件\n", removedCount)
	}

	return nil
}

// CleanupTempFilesRecursive 递归清理目录及其子目录中的临时文件
func CleanupTempFilesRecursive(dir string) error {
	// 清理当前目录
	if err := CleanupTempFiles(dir); err != nil {
		return err
	}

	// 递归清理子目录
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("读取目录失败: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			subDir := filepath.Join(dir, file.Name())
			if err := CleanupTempFilesRecursive(subDir); err != nil {
				// 忽略子目录清理错误
				continue
			}
		}
	}

	return nil
}
