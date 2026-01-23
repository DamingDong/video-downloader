package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger 定义日志记录器接口
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	BatchStart(totalTasks, concurrency int)
	BatchComplete(success, fail, skip, total int)
	DownloadSuccess(videoID, title string, retryCount int, fileSize int64)
	DownloadFail(videoID, title string, err error, retryCount int)
	DownloadSkip(videoID, title string)
	Close()
}

// SimpleLogger 实现简单的日志记录器
type SimpleLogger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
}

var ( // 全局日志记录器
	globalLogger Logger
	once         sync.Once
)

// LogLevel 定义日志级别
type LogLevel int

// 日志级别常量
const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// 日志级别字符串映射
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// InitLogger 初始化日志记录器
func InitLogger(logDir string, level LogLevel) (Logger, error) {
	// 确保日志目录存在
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 创建日志文件
	logFile := filepath.Join(logDir, time.Now().Format("2006-01-02")+".log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %w", err)
	}

	// 创建日志记录器
	logger := &SimpleLogger{
		debugLogger: log.New(io.MultiWriter(os.Stdout, file), "[DEBUG] ", log.Ldate|log.Ltime),
		infoLogger:  log.New(io.MultiWriter(os.Stdout, file), "[INFO] ", log.Ldate|log.Ltime),
		warnLogger:  log.New(io.MultiWriter(os.Stdout, file), "[WARN] ", log.Ldate|log.Ltime),
		errorLogger: log.New(io.MultiWriter(os.Stderr, file), "[ERROR] ", log.Ldate|log.Ltime),
	}

	return logger, nil
}

// GetLogger 获取全局日志记录器
func GetLogger() Logger {
	once.Do(func() {
		globalLogger = NewSimpleLogger()
	})
	return globalLogger
}

// NewSimpleLogger 创建新的简单日志记录器
func NewSimpleLogger() Logger {
	return &SimpleLogger{
		debugLogger: log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime),
		infoLogger:  log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime),
		warnLogger:  log.New(os.Stdout, "[WARN] ", log.Ldate|log.Ltime),
		errorLogger: log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime),
	}
}

// formatMessage 格式化日志消息
func formatMessage(format string, args ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	return fmt.Sprintf("%s %s", timestamp, message)
}

// Debug 记录调试日志
func (l *SimpleLogger) Debug(format string, args ...interface{}) {
	message := formatMessage(format, args...)
	l.debugLogger.Println(message)
}

// Info 记录信息日志
func (l *SimpleLogger) Info(format string, args ...interface{}) {
	message := formatMessage(format, args...)
	l.infoLogger.Println(message)
}

// Warn 记录警告日志
func (l *SimpleLogger) Warn(format string, args ...interface{}) {
	message := formatMessage(format, args...)
	l.warnLogger.Println(message)
}

// Error 记录错误日志
func (l *SimpleLogger) Error(format string, args ...interface{}) {
	message := formatMessage(format, args...)
	l.errorLogger.Println(message)
}

// BatchStart 记录批处理开始
func (l *SimpleLogger) BatchStart(totalTasks, concurrency int) {
	message := formatMessage("开始批处理下载任务: %d 个任务, 并发数: %d", totalTasks, concurrency)
	l.infoLogger.Println(message)
}

// BatchComplete 记录批处理完成
func (l *SimpleLogger) BatchComplete(success, fail, skip, total int) {
	message := formatMessage("批处理下载任务完成: 成功=%d, 失败=%d, 跳过=%d, 总计=%d", success, fail, skip, total)
	l.infoLogger.Println(message)
}

// DownloadSuccess 记录下载成功
func (l *SimpleLogger) DownloadSuccess(videoID, title string, retryCount int, fileSize int64) {
	fileSizeMB := float64(fileSize) / (1024 * 1024)
	message := formatMessage("下载成功: %s (ID: %s, 重试: %d, 大小: %.2f MB)", title, videoID, retryCount, fileSizeMB)
	l.infoLogger.Println(message)
}

// DownloadFail 记录下载失败
func (l *SimpleLogger) DownloadFail(videoID, title string, err error, retryCount int) {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	} else {
		errMsg = "未知错误"
	}
	message := formatMessage("下载失败: %s (ID: %s, 重试: %d, 错误: %s)", title, videoID, retryCount, errMsg)
	l.errorLogger.Println(message)
}

// DownloadSkip 记录下载跳过
func (l *SimpleLogger) DownloadSkip(videoID, title string) {
	message := formatMessage("下载跳过: %s (ID: %s)", title, videoID)
	l.infoLogger.Println(message)
}

// Close 关闭日志记录器
func (l *SimpleLogger) Close() {
	// 简单实现，因为标准日志记录器不需要关闭
	l.infoLogger.Println(formatMessage("日志记录器已关闭"))
}
