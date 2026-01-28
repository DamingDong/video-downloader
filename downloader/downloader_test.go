package downloader

import (
	"testing"
)

func TestYouTubeDownloader_Name(t *testing.T) {
	ytDownloader := &YouTubeDownloader{}
	name := ytDownloader.Name()
	if name != "YouTube专用下载器" {
		t.Errorf("YouTubeDownloader.Name() = %q, want %q", name, "YouTube专用下载器")
	}
}

func TestMultiPlatformDownloader_Name(t *testing.T) {
	multiDownloader := &MultiPlatformDownloader{}
	name := multiDownloader.Name()
	if name != "多平台下载器" {
		t.Errorf("MultiPlatformDownloader.Name() = %q, want %q", name, "多平台下载器")
	}
}

func TestMultiPlatformDownloader_SupportedPlatforms(t *testing.T) {
	multiDownloader := &MultiPlatformDownloader{}
	platforms := multiDownloader.SupportedPlatforms()

	if len(platforms) == 0 {
		t.Fatal("MultiPlatformDownloader.SupportedPlatforms() returned empty slice")
	}

	// Check if some common platforms are included
	expectedPlatforms := []string{"youtube", "douyin", "bilibili", "tiktok"}
	for _, expected := range expectedPlatforms {
		found := false
		for _, platform := range platforms {
			if platform == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected platform %q not found in supported platforms", expected)
		}
	}
}
