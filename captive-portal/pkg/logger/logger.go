package logger

import (
	"log"
	"os"
)

type Logger struct {
	level string
}

func New(level string) *Logger {
	return &Logger{level: level}
}

func (l *Logger) Info(msg string, args ...interface{}) {
	log.Printf("[INFO] %s %v", msg, args)
}

func (l *Logger) Error(msg string, args ...interface{}) {
	log.Printf("[ERROR] %s %v", msg, args)
}

func (l *Logger) Warn(msg string, args ...interface{}) {
	log.Printf("[WARN] %s %v", msg, args)
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	if l.level == "debug" {
		log.Printf("[DEBUG] %s %v", msg, args)
	}
}

func (l *Logger) Fatal(msg string, args ...interface{}) {
	log.Printf("[FATAL] %s %v", msg, args)
	os.Exit(1)
}
