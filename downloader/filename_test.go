package downloader

import (
	"batch_download_videos/config"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestFilenameTemplateVariables 测试文件名模板变量替换
func TestFilenameTemplateVariables(t *testing.T) {
	// 模拟视频信息
	videoInfo := &VideoInfo{
		Title:     "测试视频标题",
		ID:        "test123",
		Duration:  300, // 5分钟，短视频
		Extractor: "youtube",
	}

	// 测试模板
	template := "%(platform)s_%(content_type)s_%(title)s_%(id)s_%(timestamp)s.%(ext)s"

	// 模拟多平台下载器
	mpd := &MultiPlatformDownloader{
		config: &config.Config{
			OutputTemplate:    template,
			FilenameMaxLength: 100,
		},
	}

	// 生成文件名
	timestamp := time.Now().Format("20060102_150405")
	filename := mpd.generateFilename(videoInfo, "mp4")

	// 验证文件名包含所有必要信息
	assert.Contains(t, filename, "youtube")     // 平台信息
	assert.Contains(t, filename, "short")       // 内容类型
	assert.Contains(t, filename, "测试视频标题")      // 标题
	assert.Contains(t, filename, "test123")     // ID
	assert.Contains(t, filename, timestamp[:8]) // 日期部分
	assert.Contains(t, filename, ".mp4")        // 扩展名
}

// TestContentTypeDetection 测试内容类型判断逻辑
func TestContentTypeDetection(t *testing.T) {
	// 测试短视频（5分钟）
	shortVideo := &VideoInfo{
		Duration:  300, // 5分钟
		Extractor: "youtube",
	}

	// 测试长视频（15分钟）
	longVideo := &VideoInfo{
		Duration:  900, // 15分钟
		Extractor: "youtube",
	}

	// 模拟多平台下载器
	template := "%(platform)s_%(content_type)s_%(title)s_%(id)s_%(timestamp)s.%(ext)s"
	mpd := &MultiPlatformDownloader{
		config: &config.Config{
			OutputTemplate:    template,
			FilenameMaxLength: 100,
		},
	}

	// 生成短视频文件名
	shortFilename := mpd.generateFilename(shortVideo, "mp4")
	assert.Contains(t, shortFilename, "short")

	// 生成长视频文件名
	longFilename := mpd.generateFilename(longVideo, "mp4")
	assert.Contains(t, longFilename, "long")
}

// TestPlatformInfoExtraction 测试平台信息提取
func TestPlatformInfoExtraction(t *testing.T) {
	// 测试不同平台的视频
	platforms := []string{"youtube", "bilibili", "twitter", "tiktok"}

	// 模拟多平台下载器
	template := "%(platform)s_%(content_type)s_%(title)s_%(id)s_%(timestamp)s.%(ext)s"
	mpd := &MultiPlatformDownloader{
		config: &config.Config{
			OutputTemplate:    template,
			FilenameMaxLength: 100,
		},
	}

	for _, platform := range platforms {
		videoInfo := &VideoInfo{
			Title:     "测试视频",
			ID:        "test123",
			Duration:  300,
			Extractor: platform,
		}

		filename := mpd.generateFilename(videoInfo, "mp4")
		assert.Contains(t, filename, platform)
	}
}

// TestExtendedPlatformInfoExtraction 测试其他平台的信息提取
func TestExtendedPlatformInfoExtraction(t *testing.T) {
	// 测试其他平台的视频
	extendedPlatforms := []string{"douyin", "weibo", "vimeo", "instagram", "facebook"}

	// 模拟多平台下载器
	template := "%(platform)s_%(content_type)s_%(title)s_%(id)s_%(timestamp)s.%(ext)s"
	mpd := &MultiPlatformDownloader{
		config: &config.Config{
			OutputTemplate:    template,
			FilenameMaxLength: 100,
		},
	}

	for _, platform := range extendedPlatforms {
		videoInfo := &VideoInfo{
			Title:     "测试视频",
			ID:        "test123",
			Duration:  300,
			Extractor: platform,
		}

		filename := mpd.generateFilename(videoInfo, "mp4")
		assert.Contains(t, filename, platform)
	}
}

// TestFilenameSanitization 测试文件名清理功能
func TestFilenameSanitization(t *testing.T) {
	// 测试包含非法字符的标题
	videoInfo := &VideoInfo{
		Title:     "测试视频 / 标题 * 包含 ? 非法 | 字符",
		ID:        "test123",
		Duration:  300,
		Extractor: "youtube",
	}

	// 模拟多平台下载器
	template := "%(platform)s_%(content_type)s_%(title)s_%(id)s_%(timestamp)s.%(ext)s"
	mpd := &MultiPlatformDownloader{
		config: &config.Config{
			OutputTemplate:    template,
			FilenameMaxLength: 100,
		},
	}

	// 生成文件名
	filename := mpd.generateFilename(videoInfo, "mp4")

	// 验证文件名不包含非法字符
	assert.NotContains(t, filename, "/")
	assert.NotContains(t, filename, "*")
	assert.NotContains(t, filename, "?")
	assert.NotContains(t, filename, "|")
}

// TestFilenameMaxLength 测试文件名最大长度限制
func TestFilenameMaxLength(t *testing.T) {
	// 测试长标题
	longTitle := "这是一个非常长的视频标题，用于测试文件名长度限制功能，确保文件名不会超过系统限制"
	videoInfo := &VideoInfo{
		Title:     longTitle,
		ID:        "test123",
		Duration:  300,
		Extractor: "youtube",
	}

	// 模拟多平台下载器，设置较短的最大长度
	template := "%(platform)s_%(content_type)s_%(title)s_%(id)s_%(timestamp)s.%(ext)s"
	mpd := &MultiPlatformDownloader{
		config: &config.Config{
			OutputTemplate:    template,
			FilenameMaxLength: 50, // 设置较短的最大长度
		},
	}

	// 生成文件名
	filename := mpd.generateFilename(videoInfo, "mp4")

	// 打印调试信息
	t.Logf("FilenameMaxLength: %d", mpd.config.FilenameMaxLength)
	t.Logf("Generated filename: %s", filename)
	filenameRunes := []rune(filename)
	t.Logf("Filename length (bytes): %d", len(filename))
	t.Logf("Filename length (characters): %d", len(filenameRunes))

	// 验证文件名长度不超过限制
	assert.LessOrEqual(t, len(filenameRunes), 50)
}
