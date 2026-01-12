package indexer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Indexer struct {
	index      map[string]bool
	indexMutex sync.RWMutex
	baseDir    string
	indexFile  string
}

func NewIndexer(baseDir string) *Indexer {
	return &Indexer{
		index:     make(map[string]bool),
		baseDir:   baseDir,
		indexFile: filepath.Join(baseDir, ".video_downloaded.index"),
	}
}

func (idx *Indexer) Load() error {
	file, err := os.Open(idx.indexFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("打开索引文件失败: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || len(line) == 0 {
			continue
		}
		if line[0] == '#' {
			continue
		}

		videoID := line
		idx.indexMutex.Lock()
		idx.index[videoID] = true
		idx.indexMutex.Unlock()
	}

	return scanner.Err()
}

func (idx *Indexer) Save() error {
	file, err := os.Create(idx.indexFile)
	if err != nil {
		return fmt.Errorf("创建索引文件失败: %w", err)
	}
	defer file.Close()

	idx.indexMutex.RLock()
	defer idx.indexMutex.RUnlock()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	_, err = writer.WriteString("# 视频下载索引\n")
	if err != nil {
		return err
	}

	for vid := range idx.index {
		_, err = writer.WriteString(fmt.Sprintf("%s\n", vid))
		if err != nil {
			return err
		}
	}

	return nil
}

func (idx *Indexer) IsDownloaded(videoID string) bool {
	idx.indexMutex.RLock()
	defer idx.indexMutex.RUnlock()

	_, exists := idx.index[videoID]
	return exists
}

func (idx *Indexer) MarkDownloaded(videoID string) {
	idx.indexMutex.Lock()
	defer idx.indexMutex.Unlock()

	idx.index[videoID] = true
}

func (idx *Indexer) GetCount() int {
	idx.indexMutex.RLock()
	defer idx.indexMutex.RUnlock()

	return len(idx.index)
}

func (idx *Indexer) Clear() {
	idx.indexMutex.Lock()
	defer idx.indexMutex.Unlock()

	idx.index = make(map[string]bool)
}
