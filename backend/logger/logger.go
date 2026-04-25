package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

// Level 日志级别
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var (
	stdLogger   *log.Logger
	level       = INFO
	enableFile  = false
	logFile     *os.File
)

// Init 初始化日志系统
func Init(levelStr string) {
	// 设置日志输出到标准输出
	stdLogger = log.New(os.Stdout, "", 0)
	
	// 解析日志级别
	switch levelStr {
	case "debug":
		level = DEBUG
	case "info":
		level = INFO
	case "warn":
		level = WARN
	case "error":
		level = ERROR
	default:
		level = INFO
	}
	
	// 尝试启用文件日志
	initFileLogger()
}

// initFileLogger 初始化文件日志
func initFileLogger() {
	// 获取日志目录
	logDir := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		logDir = os.Getenv("APPDATA")
	}
	if logDir == "" {
		logDir = "."
	}
	
	// 打开日志文件
	logPath := fmt.Sprintf("%s/.devcleaner/devcleaner.log", logDir)
	if err := os.MkdirAll(logPath[:len(logPath)-len("/devcleaner.log")], 0755); err != nil {
		return
	}
	
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	
	logFile = file
	enableFile = true
	
	// 同时输出到文件和控制台
	stdLogger = log.New(os.Stdout, "", 0)
}

// Close 关闭日志文件
func Close() {
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}
	enableFile = false
}

// write 内部日志函数
func write(lvl Level, format string, v ...interface{}) {
	if lvl < level {
		return
	}
	
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := ""
	
	switch lvl {
	case DEBUG:
		levelStr = "DEBUG"
	case INFO:
		levelStr = "INFO"
	case WARN:
		levelStr = "WARN"
	case ERROR:
		levelStr = "ERROR"
	}
	
	msg := fmt.Sprintf(format, v...)
	logLine := fmt.Sprintf("[%s] [%s] %s", timestamp, levelStr, msg)
	
	// 输出到控制台
	stdLogger.Println(logLine)
	
	// 输出到文件
	if enableFile && logFile != nil {
		fmt.Fprintln(logFile, logLine)
	}
}

// Debug 调试日志
func Debug(format string, v ...interface{}) {
	write(DEBUG, format, v...)
}

// Info 信息日志
func Info(format string, v ...interface{}) {
	write(INFO, format, v...)
}

// Warn 警告日志
func Warn(format string, v ...interface{}) {
	write(WARN, format, v...)
}

// Error 错误日志
func Error(format string, v ...interface{}) {
	write(ERROR, format, v...)
}

// WithError 带错误的日志
func WithError(err error, format string, v ...interface{}) {
	write(ERROR, fmt.Sprintf("%s: %v", fmt.Sprintf(format, v...), err))
}
