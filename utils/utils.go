package utils

import (
	"fmt"
	"os"
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
