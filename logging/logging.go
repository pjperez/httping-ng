package logging

import (
	"fmt"
	"time"
)

type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
	LevelDebug Level = "DEBUG"
)

// Log prints a standardized log message
func Log(level Level, component string, format string, args ...any) {

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("[%s] [%s] [%s] %s\n", timestamp, level, component, msg)
}

// Convenience wrappers

func Info(component string, format string, args ...any) {
	Log(LevelInfo, component, format, args...)
}

func Warn(component string, format string, args ...any) {
	Log(LevelWarn, component, format, args...)
}

func Error(component string, format string, args ...any) {
	Log(LevelError, component, format, args...)
}

func Debug(component string, format string, args ...any) {
	Log(LevelDebug, component, format, args...)
}
