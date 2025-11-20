package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Logger struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	file     *os.File
	logDir   string
}

func NewLogger(logDir string) (*Logger, error) {
	// Create logs directory if not exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logger := &Logger{
		logDir: logDir,
	}

	if err := logger.rotate(); err != nil {
		return nil, err
	}

	// Start rotation goroutine
	go logger.autoRotate()

	return logger, nil
}

func (l *Logger) rotate() error {
	// Close existing file if any
	if l.file != nil {
		l.file.Close()
	}

	// Generate filename with date
	filename := fmt.Sprintf("app-%s.log", time.Now().Format("2006-01-02"))
	filepath := filepath.Join(l.logDir, filename)

	// Open/create log file
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	l.file = file

	// Create multi-writer (file + stdout)
	multiWriter := io.MultiWriter(os.Stdout, file)

	// Initialize loggers
	l.infoLog = log.New(multiWriter, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	l.errorLog = log.New(multiWriter, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)

	return nil
}

func (l *Logger) autoRotate() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		// Check if we need to rotate (every 3 days)
		currentFile := filepath.Join(l.logDir, fmt.Sprintf("app-%s.log", time.Now().Format("2006-01-02")))
		if _, err := os.Stat(currentFile); os.IsNotExist(err) {
			l.rotate()
		}

		// Clean old logs (older than 30 days)
		l.cleanOldLogs(30)
	}
}

func (l *Logger) cleanOldLogs(daysToKeep int) {
	files, err := filepath.Glob(filepath.Join(l.logDir, "app-*.log"))
	if err != nil {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -daysToKeep)

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			os.Remove(file)
		}
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.infoLog.Output(2, fmt.Sprintf(format, v...))
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.errorLog.Output(2, fmt.Sprintf(format, v...))
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.infoLog.Output(2, fmt.Sprintf(format, v...))
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.errorLog.Output(2, fmt.Sprintf(format, v...))
}

func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}