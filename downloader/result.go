package downloader

import "time"

// DownloadStats 定义下载统计信息
type DownloadStats struct {
	TotalTasks     int        `json:"total_tasks"`
	CompletedTasks int        `json:"completed_tasks"`
	FailedTasks    int        `json:"failed_tasks"`
	TotalSize      int64      `json:"total_size"`
	TotalTime      int64      `json:"total_time"`
	AverageSpeed   float64    `json:"average_speed"` // 平均下载速度（字节/秒）
	StartTime      time.Time  `json:"start_time"`
	EndTime        *time.Time `json:"end_time,omitempty"`
}

// NewDownloadStats 创建新的下载统计信息
func NewDownloadStats() *DownloadStats {
	return &DownloadStats{
		StartTime: time.Now(),
	}
}

// UpdateWithResult 根据下载结果更新统计信息
func (ds *DownloadStats) UpdateWithResult(result *DownloadResult) {
	ds.TotalTasks++
	if result.Success {
		ds.CompletedTasks++
		ds.TotalSize += result.FileSize
	} else {
		ds.FailedTasks++
	}
}

// Complete 完成统计信息
func (ds *DownloadStats) Complete() {
	endTime := time.Now()
	ds.EndTime = &endTime
}
