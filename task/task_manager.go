package task

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"batch_download_videos/downloader"
	"batch_download_videos/logger"
)

// TaskStatus 定义任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"   // 等待中
	TaskStatusDownloading TaskStatus = "downloading" // 下载中
	TaskStatusPaused    TaskStatus = "paused"    // 暂停
	TaskStatusCompleted TaskStatus = "completed" // 完成
	TaskStatusFailed    TaskStatus = "failed"    // 失败
	TaskStatusCanceled  TaskStatus = "canceled"  // 取消
)

// DownloadTask 定义下载任务
type DownloadTask struct {
	ID          string          `json:"id"`
	URL         string          `json:"url"`
	OutputDir   string          `json:"output_dir"`
	Resolution  string          `json:"resolution"`
	Status      TaskStatus      `json:"status"`
	Error       string          `json:"error"`
	Progress    float64         `json:"progress"`
	Speed       string          `json:"speed"`
	ETA         string          `json:"eta"`
	FileSize    int64           `json:"file_size"`
	RetryCount  int             `json:"retry_count"`
	CreatedAt   time.Time       `json:"created_at"`
	StartedAt   *time.Time      `json:"started_at"`
	CompletedAt *time.Time      `json:"completed_at"`
	Ctx         context.Context `json:"-"`
	CancelFunc  context.CancelFunc `json:"-"`
	Result      *downloader.DownloadResult `json:"-"`
	Mutex       sync.Mutex      `json:"-"`
}

// TaskManager 定义任务管理器
type TaskManager struct {
	Tasks      map[string]*DownloadTask `json:"tasks"`
	TaskQueue  []string                 `json:"task_queue"`
	Processing []string                 `json:"processing"`
	MaxConcurrent int                   `json:"max_concurrent"`
	Mutex      sync.RWMutex             `json:"-"`
	PersistFile string                  `json:"-"`
}

// NewTaskManager 创建新的任务管理器
func NewTaskManager(maxConcurrent int, persistFile string) *TaskManager {
	manager := &TaskManager{
		Tasks:      make(map[string]*DownloadTask),
		TaskQueue:  make([]string, 0),
		Processing: make([]string, 0),
		MaxConcurrent: maxConcurrent,
		PersistFile: persistFile,
	}
	
	// 尝试加载持久化的任务状态
	manager.Load()
	
	return manager
}

// AddTask 添加新的下载任务
func (tm *TaskManager) AddTask(url, outputDir, resolution string) *DownloadTask {
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()
	
	// 生成任务ID
	taskID := fmt.Sprintf("%s_%d", time.Now().Format("20060102_150405"), len(tm.Tasks))
	
	// 创建任务
	task := &DownloadTask{
		ID:         taskID,
		URL:        url,
		OutputDir:  outputDir,
		Resolution: resolution,
		Status:     TaskStatusPending,
		Progress:   0,
		CreatedAt:  time.Now(),
	}
	
	// 添加到任务映射和队列
	tm.Tasks[taskID] = task
	tm.TaskQueue = append(tm.TaskQueue, taskID)
	
	// 持久化任务状态
	tm.Save()
	
	logger.GetLogger().Info("添加下载任务: %s (URL: %s)", taskID, url)
	
	return task
}

// StartTask 开始执行任务
func (tm *TaskManager) StartTask(taskID string) error {
	tm.Mutex.Lock()
	
	// 检查任务是否存在
	task, exists := tm.Tasks[taskID]
	if !exists {
		tm.Mutex.Unlock()
		return fmt.Errorf("任务不存在: %s", taskID)
	}
	
	// 检查任务状态
	if task.Status != TaskStatusPending && task.Status != TaskStatusPaused {
		tm.Mutex.Unlock()
		return fmt.Errorf("任务状态不允许开始: %s", task.Status)
	}
	
	// 检查并发数限制
	if len(tm.Processing) >= tm.MaxConcurrent {
		tm.Mutex.Unlock()
		return fmt.Errorf("达到最大并发数限制: %d", tm.MaxConcurrent)
	}
	
	// 更新任务状态
	task.Mutex.Lock()
	task.Status = TaskStatusDownloading
	now := time.Now()
	task.StartedAt = &now
	task.Mutex.Unlock()
	
	// 将任务从队列移到处理中
	for i, id := range tm.TaskQueue {
		if id == taskID {
			tm.TaskQueue = append(tm.TaskQueue[:i], tm.TaskQueue[i+1:]...)
			break
		}
	}
	tm.Processing = append(tm.Processing, taskID)
	
	// 持久化任务状态
	tm.Save()
	tm.Mutex.Unlock()
	
	logger.GetLogger().Info("开始执行任务: %s", taskID)
	return nil
}

// PauseTask 暂停任务
func (tm *TaskManager) PauseTask(taskID string) error {
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()
	
	// 检查任务是否存在
	task, exists := tm.Tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}
	
	// 检查任务状态
	if task.Status != TaskStatusDownloading {
		return fmt.Errorf("任务状态不允许暂停: %s", task.Status)
	}
	
	// 取消任务上下文
	if task.CancelFunc != nil {
		task.CancelFunc()
	}
	
	// 更新任务状态
	task.Mutex.Lock()
	task.Status = TaskStatusPaused
	task.Mutex.Unlock()
	
	// 将任务从处理中移回队列
	for i, id := range tm.Processing {
		if id == taskID {
			tm.Processing = append(tm.Processing[:i], tm.Processing[i+1:]...)
			break
		}
	}
	tm.TaskQueue = append(tm.TaskQueue, taskID)
	
	// 持久化任务状态
	tm.Save()
	
	logger.GetLogger().Info("暂停任务: %s", taskID)
	return nil
}

// CancelTask 取消任务
func (tm *TaskManager) CancelTask(taskID string) error {
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()
	
	// 检查任务是否存在
	task, exists := tm.Tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}
	
	// 检查任务状态
	if task.Status == TaskStatusCompleted || task.Status == TaskStatusFailed || task.Status == TaskStatusCanceled {
		return fmt.Errorf("任务状态不允许取消: %s", task.Status)
	}
	
	// 取消任务上下文
	if task.CancelFunc != nil {
		task.CancelFunc()
	}
	
	// 更新任务状态
	task.Mutex.Lock()
	task.Status = TaskStatusCanceled
	task.Mutex.Unlock()
	
	// 从队列或处理中移除任务
	for i, id := range tm.TaskQueue {
		if id == taskID {
			tm.TaskQueue = append(tm.TaskQueue[:i], tm.TaskQueue[i+1:]...)
			break
		}
	}
	for i, id := range tm.Processing {
		if id == taskID {
			tm.Processing = append(tm.Processing[:i], tm.Processing[i+1:]...)
			break
		}
	}
	
	// 持久化任务状态
	tm.Save()
	
	logger.GetLogger().Info("取消任务: %s", taskID)
	return nil
}

// CompleteTask 完成任务
func (tm *TaskManager) CompleteTask(taskID string, result *downloader.DownloadResult) error {
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()
	
	// 检查任务是否存在
	task, exists := tm.Tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}
	
	// 更新任务状态
	task.Mutex.Lock()
	task.Status = TaskStatusCompleted
	task.Progress = 100
	task.Result = result
	if result != nil {
		task.FileSize = result.FileSize
	}
	now := time.Now()
	task.CompletedAt = &now
	task.Mutex.Unlock()
	
	// 从处理中移除任务
	for i, id := range tm.Processing {
		if id == taskID {
			tm.Processing = append(tm.Processing[:i], tm.Processing[i+1:]...)
			break
		}
	}
	
	// 持久化任务状态
	tm.Save()
	
	logger.GetLogger().Info("任务完成: %s", taskID)
	return nil
}

// FailTask 标记任务失败
func (tm *TaskManager) FailTask(taskID string, err error) error {
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()
	
	// 检查任务是否存在
	task, exists := tm.Tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}
	
	// 更新任务状态
	task.Mutex.Lock()
	task.Status = TaskStatusFailed
	if err != nil {
		task.Error = err.Error()
	}
	now := time.Now()
	task.CompletedAt = &now
	task.Mutex.Unlock()
	
	// 从处理中移除任务
	for i, id := range tm.Processing {
		if id == taskID {
			tm.Processing = append(tm.Processing[:i], tm.Processing[i+1:]...)
			break
		}
	}
	
	// 持久化任务状态
	tm.Save()
	
	logger.GetLogger().Info("任务失败: %s, 错误: %v", taskID, err)
	return nil
}

// UpdateTaskProgress 更新任务进度
func (tm *TaskManager) UpdateTaskProgress(taskID string, progress float64, speed, eta string) error {
	tm.Mutex.RLock()
	task, exists := tm.Tasks[taskID]
	tm.Mutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}
	
	task.Mutex.Lock()
	task.Progress = progress
	task.Speed = speed
	task.ETA = eta
	task.Mutex.Unlock()
	
	// 每5%进度持久化一次，避免频繁IO操作
	if int(progress)%5 == 0 {
		tm.Save()
	}
	
	return nil
}

// GetTask 获取任务信息
func (tm *TaskManager) GetTask(taskID string) (*DownloadTask, error) {
	tm.Mutex.RLock()
	defer tm.Mutex.RUnlock()
	
	task, exists := tm.Tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("任务不存在: %s", taskID)
	}
	
	return task, nil
}

// ListTasks 列出所有任务
func (tm *TaskManager) ListTasks() []*DownloadTask {
	tm.Mutex.RLock()
	defer tm.Mutex.RUnlock()
	
	tasks := make([]*DownloadTask, 0, len(tm.Tasks))
	for _, task := range tm.Tasks {
		tasks = append(tasks, task)
	}
	
	return tasks
}

// GetPendingTasks 获取等待中的任务
func (tm *TaskManager) GetPendingTasks() []*DownloadTask {
	tm.Mutex.RLock()
	defer tm.Mutex.RUnlock()
	
	tasks := make([]*DownloadTask, 0)
	for _, taskID := range tm.TaskQueue {
		if task, exists := tm.Tasks[taskID]; exists {
			tasks = append(tasks, task)
		}
	}
	
	return tasks
}

// GetProcessingTasks 获取处理中的任务
func (tm *TaskManager) GetProcessingTasks() []*DownloadTask {
	tm.Mutex.RLock()
	defer tm.Mutex.RUnlock()
	
	tasks := make([]*DownloadTask, 0)
	for _, taskID := range tm.Processing {
		if task, exists := tm.Tasks[taskID]; exists {
			tasks = append(tasks, task)
		}
	}
	
	return tasks
}

// Save 持久化任务状态
func (tm *TaskManager) Save() error {
	if tm.PersistFile == "" {
		return nil
	}
	
	// 确保持久化文件目录存在
	dir := filepath.Dir(tm.PersistFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	
	// 创建一个不包含上下文的任务副本
	tasksCopy := make(map[string]*DownloadTask)
	tm.Mutex.RLock()
	for id, task := range tm.Tasks {
		taskCopy := *task
		taskCopy.Ctx = nil
		taskCopy.CancelFunc = nil
		taskCopy.Result = nil
		tasksCopy[id] = &taskCopy
	}
	
	// 保存任务队列和处理中的任务
	taskQueue := tm.TaskQueue
	processing := tm.Processing
	tm.Mutex.RUnlock()
	
	// 构建保存数据
	saveData := map[string]interface{}{
		"tasks":      tasksCopy,
		"task_queue": taskQueue,
		"processing": processing,
	}
	
	// 写入文件
	data, err := json.MarshalIndent(saveData, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化任务数据失败: %w", err)
	}
	
	if err := os.WriteFile(tm.PersistFile, data, 0644); err != nil {
		return fmt.Errorf("写入任务数据失败: %w", err)
	}
	
	return nil
}

// Load 加载任务状态
func (tm *TaskManager) Load() error {
	if tm.PersistFile == "" {
		return nil
	}
	
	// 检查文件是否存在
	if _, err := os.Stat(tm.PersistFile); os.IsNotExist(err) {
		return nil
	}
	
	// 读取文件
	data, err := os.ReadFile(tm.PersistFile)
	if err != nil {
		return fmt.Errorf("读取任务数据失败: %w", err)
	}
	
	// 解析数据
	var loadData map[string]interface{}
	if err := json.Unmarshal(data, &loadData); err != nil {
		return fmt.Errorf("解析任务数据失败: %w", err)
	}
	
	// 加载任务
	if tasksData, ok := loadData["tasks"].(map[string]interface{}); ok {
		for id, taskData := range tasksData {
			if taskJSON, err := json.Marshal(taskData); err == nil {
				var task DownloadTask
				if err := json.Unmarshal(taskJSON, &task); err == nil {
					tm.Tasks[id] = &task
				}
			}
		}
	}
	
	// 加载任务队列
	if queueData, ok := loadData["task_queue"].([]interface{}); ok {
		for _, idData := range queueData {
			if id, ok := idData.(string); ok {
				tm.TaskQueue = append(tm.TaskQueue, id)
			}
		}
	}
	
	// 加载处理中的任务
	if processingData, ok := loadData["processing"].([]interface{}); ok {
		for _, idData := range processingData {
			if id, ok := idData.(string); ok {
				tm.Processing = append(tm.Processing, id)
			}
		}
	}
	
	logger.GetLogger().Info("加载任务状态: %d 个任务, %d 个等待中, %d 个处理中", 
		len(tm.Tasks), len(tm.TaskQueue), len(tm.Processing))
	
	return nil
}

// NextTask 获取下一个待处理的任务
func (tm *TaskManager) NextTask() (*DownloadTask, error) {
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()
	
	if len(tm.TaskQueue) == 0 {
		return nil, fmt.Errorf("没有待处理的任务")
	}
	
	// 检查并发数限制
	if len(tm.Processing) >= tm.MaxConcurrent {
		return nil, fmt.Errorf("达到最大并发数限制: %d", tm.MaxConcurrent)
	}
	
	// 获取第一个任务
	taskID := tm.TaskQueue[0]
	task, exists := tm.Tasks[taskID]
	if !exists {
		// 任务不存在，从队列中移除
		tm.TaskQueue = tm.TaskQueue[1:]
		tm.Save()
		return nil, fmt.Errorf("任务不存在: %s", taskID)
	}
	
	// 创建任务上下文
	ctx, cancel := context.WithCancel(context.Background())
	task.Ctx = ctx
	task.CancelFunc = cancel
	
	// 更新任务状态
	task.Mutex.Lock()
	task.Status = TaskStatusDownloading
	now := time.Now()
	task.StartedAt = &now
	task.Mutex.Unlock()
	
	// 将任务从队列移到处理中
	tm.TaskQueue = tm.TaskQueue[1:]
	tm.Processing = append(tm.Processing, taskID)
	
	// 持久化任务状态
	tm.Save()
	
	logger.GetLogger().Info("开始处理任务: %s (URL: %s)", taskID, task.URL)
	
	return task, nil
}
