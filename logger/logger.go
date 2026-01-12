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

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

type Logger struct {
	mu            sync.Mutex
	level         LogLevel
	fileLogger    *log.Logger
	consoleLogger *log.Logger
	logFile       *os.File
}

var (
	instance *Logger
	once     sync.Once
)

func InitLogger(logDir string, level LogLevel) (*Logger, error) {
	var initErr error
	once.Do(func() {
		instance = &Logger{
			level:         level,
			consoleLogger: log.New(os.Stdout, "", log.LstdFlags),
		}

		if logDir != "" {
			if err := os.MkdirAll(logDir, 0755); err != nil {
				initErr = fmt.Errorf("创建日志目录失败: %w", err)
				return
			}

			timestamp := time.Now().Format("20060102_150405")
			logPath := filepath.Join(logDir, fmt.Sprintf("download_%s.log", timestamp))

			file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				initErr = fmt.Errorf("打开日志文件失败: %w", err)
				return
			}

			instance.logFile = file
			instance.fileLogger = log.New(io.MultiWriter(os.Stdout, file), "", log.LstdFlags)
		}
	})

	return instance, initErr
}

func GetLogger() *Logger {
	if instance == nil {
		InitLogger("", INFO)
	}
	return instance
}

func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= DEBUG {
		l.log("[DEBUG]", format, v...)
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= INFO {
		l.log("[INFO]", format, v...)
	}
}

func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level <= WARN {
		l.log("[WARN]", format, v...)
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= ERROR {
		l.log("[ERROR]", format, v...)
	}
}

func (l *Logger) log(level, format string, v ...interface{}) {
	message := fmt.Sprintf("%s %s", level, fmt.Sprintf(format, v...))

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.fileLogger != nil {
		l.fileLogger.Println(message)
	} else {
		l.consoleLogger.Println(message)
	}
}

func (l *Logger) DownloadStart(videoID, title string) {
	l.Info("开始下载: %s (ID: %s)", title, videoID)
}

func (l *Logger) DownloadSuccess(videoID, title string, retryCount int, fileSize int64) {
	l.Info("下载成功: %s (ID: %s, 重试: %d, 大小: %s)",
		title, videoID, retryCount, formatFileSize(fileSize))
}

func (l *Logger) DownloadFail(videoID, title string, err error, retryCount int) {
	l.Error("下载失败: %s (ID: %s, 尝试: %d, 错误: %v)",
		title, videoID, retryCount, err)
}

func (l *Logger) DownloadSkip(videoID, title string) {
	l.Info("跳过已下载: %s (ID: %s)", title, videoID)
}

func (l *Logger) BatchStart(total int, maxConcurrency int) {
	l.Info("开始处理 %d 个URL，最大并发数: %d", total, maxConcurrency)
}

func (l *Logger) BatchComplete(success, fail, skip, total int) {
	l.Info("下载完成统计: 成功=%d, 失败=%d, 跳过=%d, 总计=%d",
		success, fail, skip, total)
}

func formatFileSize(bytes int64) string {
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
