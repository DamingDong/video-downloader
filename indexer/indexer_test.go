package indexer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewIndexer(t *testing.T) {
	tempDir := t.TempDir()
	idx := NewIndexer(tempDir)

	if idx == nil {
		t.Fatal("NewIndexer() returned nil")
	}

	if idx.indexFile != filepath.Join(tempDir, ".video_downloaded.index") {
		t.Errorf("indexFile = %q, want %q", idx.indexFile, filepath.Join(tempDir, ".video_downloaded.index"))
	}

	if len(idx.index) != 0 {
		t.Errorf("index length = %d, want 0", len(idx.index))
	}
}

func TestIndexerLoad(t *testing.T) {
	tempDir := t.TempDir()
	idx := NewIndexer(tempDir)

	indexFile := filepath.Join(tempDir, ".video_downloaded.index")
	content := "# Video download index\nvideo1\nvideo2\nvideo3\n"
	if err := os.WriteFile(indexFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test index file: %v", err)
	}

	if err := idx.Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if len(idx.index) != 3 {
		t.Errorf("index length = %d, want 3", len(idx.index))
	}

	if !idx.IsDownloaded("video1") {
		t.Error("video1 should be marked as downloaded")
	}

	if !idx.IsDownloaded("video2") {
		t.Error("video2 should be marked as downloaded")
	}

	if !idx.IsDownloaded("video3") {
		t.Error("video3 should be marked as downloaded")
	}
}

func TestIndexerLoadNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	idx := NewIndexer(tempDir)

	if err := idx.Load(); err != nil {
		t.Errorf("Load() on non-existent file should not error, got: %v", err)
	}
}

func TestIndexerSave(t *testing.T) {
	tempDir := t.TempDir()
	idx := NewIndexer(tempDir)

	idx.MarkDownloaded("video1")
	idx.MarkDownloaded("video2")

	if err := idx.Save(); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	indexFile := filepath.Join(tempDir, ".video_downloaded.index")
	content, err := os.ReadFile(indexFile)
	if err != nil {
		t.Fatalf("Failed to read saved index file: %v", err)
	}

	expectedContent := "# 视频下载索引\nvideo1\nvideo2\n"
	if string(content) != expectedContent {
		t.Errorf("Saved content = %q, want %q", string(content), expectedContent)
	}
}

func TestIndexerIsDownloaded(t *testing.T) {
	tempDir := t.TempDir()
	idx := NewIndexer(tempDir)

	if idx.IsDownloaded("video1") {
		t.Error("video1 should not be marked as downloaded initially")
	}

	idx.MarkDownloaded("video1")

	if !idx.IsDownloaded("video1") {
		t.Error("video1 should be marked as downloaded after MarkDownloaded")
	}
}

func TestIndexerMarkDownloaded(t *testing.T) {
	tempDir := t.TempDir()
	idx := NewIndexer(tempDir)

	idx.MarkDownloaded("video1")
	idx.MarkDownloaded("video2")

	if len(idx.index) != 2 {
		t.Errorf("index length = %d, want 2", len(idx.index))
	}

	if !idx.IsDownloaded("video1") {
		t.Error("video1 should be marked as downloaded")
	}

	if !idx.IsDownloaded("video2") {
		t.Error("video2 should be marked as downloaded")
	}
}

func TestIndexerGetCount(t *testing.T) {
	tempDir := t.TempDir()
	idx := NewIndexer(tempDir)

	if count := idx.GetCount(); count != 0 {
		t.Errorf("GetCount() = %d, want 0", count)
	}

	idx.MarkDownloaded("video1")
	idx.MarkDownloaded("video2")
	idx.MarkDownloaded("video3")

	if count := idx.GetCount(); count != 3 {
		t.Errorf("GetCount() = %d, want 3", count)
	}
}

func TestIndexerClear(t *testing.T) {
	tempDir := t.TempDir()
	idx := NewIndexer(tempDir)

	idx.MarkDownloaded("video1")
	idx.MarkDownloaded("video2")

	if len(idx.index) != 2 {
		t.Errorf("index length = %d, want 2 before Clear", len(idx.index))
	}

	idx.Clear()

	if len(idx.index) != 0 {
		t.Errorf("index length = %d, want 0 after Clear", len(idx.index))
	}

	if idx.IsDownloaded("video1") {
		t.Error("video1 should not be marked as downloaded after Clear")
	}
}
