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

type Logger struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	file     *os.File
	logDir   string
	mu       sync.Mutex
}

func NewLogger(logDir string) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logger := &Logger{
		logDir: logDir,
	}

	if err := logger.rotate(); err != nil {
		return nil, err
	}

	return logger, nil
}

func (l *Logger) rotate() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		l.file.Sync()
		l.file.Close()
	}

	filename := fmt.Sprintf("app-%s.log", time.Now().Format("2006-01-02"))
	logPath := filepath.Join(l.logDir, filename)

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	l.file = file
	multiWriter := io.MultiWriter(os.Stdout, file)

	l.infoLog = log.New(multiWriter, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	l.errorLog = log.New(multiWriter, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)

	return nil
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	message := fmt.Sprintf(format, v...)
	l.infoLog.Output(2, message)
	if l.file != nil {
		l.file.Sync()
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	message := fmt.Sprintf(format, v...)
	l.errorLog.Output(2, message)
	if l.file != nil {
		l.file.Sync()
	}
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.Info(format, v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Error(format, v...)
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		l.file.Sync()
		return l.file.Close()
	}
	return nil
}

func (l *Logger) GetLogPath() string {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Name()
	}
	return ""
}