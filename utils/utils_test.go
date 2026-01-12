package utils

import (
	"testing"
)

func TestGetWebsiteType(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"YouTube", "https://www.youtube.com/watch?v=test", "youtube"},
		{"YouTube Short", "https://youtu.be/test", "youtube"},
		{"Douyin", "https://www.douyin.com/video/test", "douyin"},
		{"Weibo", "https://weibo.com/tv/show/test", "weibo"},
		{"Bilibili", "https://www.bilibili.com/video/test", "bilibili"},
		{"TikTok", "https://www.tiktok.com/@user/video/test", "tiktok"},
		{"Vimeo", "https://vimeo.com/test", "vimeo"},
		{"Instagram", "https://www.instagram.com/p/test", "instagram"},
		{"Twitter", "https://twitter.com/user/status/test", "twitter"},
		{"X.com", "https://x.com/user/status/test", "twitter"},
		{"Facebook", "https://www.facebook.com/watch?v=test", "facebook"},
		{"Unknown", "https://unknown.com/video", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetWebsiteType(tt.url)
			if result != tt.expected {
				t.Errorf("GetWebsiteType(%q) = %q, want %q", tt.url, result, tt.expected)
			}
		})
	}
}

func TestGetQualityFormat(t *testing.T) {
	tests := []struct {
		name     string
		res      string
		expected string
	}{
		{"1080p", "1080", "bestvideo[height<=1080]+bestaudio/best[height<=1080]"},
		{"720p", "720", "bestvideo[height<=720]+bestaudio/best[height<=720]"},
		{"480p", "480", "bestvideo[height<=480]+bestaudio/best[height<=480]"},
		{"360p", "360", "bestvideo[height<=360]+bestaudio/best[height<=360]"},
		{"Unknown", "unknown", "bestvideo[height<=720]+bestaudio/best[height<=720]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetQualityFormat(tt.res)
			if result != tt.expected {
				t.Errorf("GetQualityFormat(%q) = %q, want %q", tt.res, result, tt.expected)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal", "video.mp4", "video.mp4"},
		{"With Slash", "video/test.mp4", "video_test.mp4"},
		{"With Backslash", "video\\test.mp4", "video_test.mp4"},
		{"With Colon", "video:test.mp4", "video_test.mp4"},
		{"With Asterisk", "video*.mp4", "video_.mp4"},
		{"With Question", "video?.mp4", "video_.mp4"},
		{"With Quote", "video\".mp4", "video_.mp4"},
		{"With Less Than", "video<.mp4", "video_.mp4"},
		{"With Greater Than", "video>.mp4", "video_.mp4"},
		{"With Pipe", "video|.mp4", "video_.mp4"},
		{"Multiple Special Chars", "video/:*?\"<>|.mp4", "video________.mp4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"Bytes", 512, "512 B"},
		{"KB", 1024, "1.0 KB"},
		{"MB", 1024 * 1024, "1.0 MB"},
		{"GB", 1024 * 1024 * 1024, "1.0 GB"},
		{"TB", int64(1024) * 1024 * 1024 * 1024, "1.0 TB"},
		{"Fractional", 1536, "1.5 KB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatFileSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatFileSize(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"Shorter than max", "hello", 10, "hello"},
		{"Equal to max", "hello", 5, "hello"},
		{"Longer than max", "hello world", 5, "he..."},
		{"Very long", "this is a very long string", 10, "this is..."},
		{"Max less than 3", "hello", 2, "he"},
		{"Empty string", "", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("TruncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}
