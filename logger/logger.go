// Copyright 2025 Aleksey Dobshikov
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

type Logger struct {
	logger *log.Logger
}

const (
	infoPrefix    string = "[INFO]  "
	debugPrefix   string = "[DEBUG] "
	warningPrefix string = "[WARN]  "
	errorPrefix   string = "[ERROR] "
)

var (
	logger    *Logger
	debugMode bool
)

func init() {
	// Ensure debug mode is false by default
	debugMode = false
}

// getCallerInfo retrieves the caller's file and line number, skipping logger internal calls
func getCallerInfo() (string, int) {
	// Start from frame 2 (0 = runtime.Caller, 1 = getCallerInfo, 2 = caller)
	for i := 2; i < 10; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		// Skip logger internal files
		if strings.Contains(file, "logger.go") {
			continue
		}

		// Extract just the filename (not the full path)
		for j := len(file) - 1; j > 0; j-- {
			if file[j] == '/' || file[j] == '\\' {
				return file[j+1:], line
			}
		}
		return file, line
	}
	return "unknown", 0
}

func NewLogger() *Logger {
	return &Logger{
		logger: log.New(os.Stdout, "", 0), // No flags, we'll format manually
	}
}

func GetLogger() *Logger {
	if logger == nil {
		logger = NewLogger()
	}
	return logger
}

// Adding a helper function to get the logger with file info
func GetLoggerWithFile() *Logger {
	if logger == nil {
		logger = NewLogger()
	}
	return logger
}

func (l *Logger) GetInfoPrefix() string {
	return infoPrefix
}

func (l *Logger) GetDebugPrefix() string {
	return debugPrefix
}

func (l *Logger) GetWarnPrefix() string {
	return warningPrefix
}

func (l *Logger) GetErrPrefix() string {
	return errorPrefix
}

func (l *Logger) SetOutput(w io.Writer) {
	l.logger.SetOutput(w)
}

// formatMessage creates a consistent log message format
func (l *Logger) formatMessage(prefix, format string, v ...interface{}) string {
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	return fmt.Sprintf("%s%s %s", prefix, timestamp, fmt.Sprintf(format, v...))
}

// Info записывает сообщение с префиксом INFO в лог.
func (l *Logger) Info(format string, v ...interface{}) {
	l.logger.Print(l.formatMessage(infoPrefix, format, v...))
}

// Debug записывает сообщение с префиксом DEBUG в лог.
func (l *Logger) Debug(format string, v ...interface{}) {
	// Check if debug mode is enabled
	if !debugMode {
		return
	}

	message := l.formatMessage(debugPrefix, format, v...)

	// If debug mode is enabled, include caller information
	file, line := getCallerInfo()
	l.logger.Printf("%s (%s:%d)", message, file, line)
}

// Warning записывает сообщение с префиксом WARNING в лог.
func (l *Logger) Warning(format string, v ...interface{}) {
	l.logger.Print(l.formatMessage(warningPrefix, format, v...))
}

// Error записывает сообщение с префиксом ERROR в лог.
func (l *Logger) Error(format string, v ...interface{}) error {
	message := l.formatMessage(errorPrefix, format, v...)
	l.logger.Print(message)
	return fmt.Errorf(format, v...)
}

// SetDebug enables or disables debug mode
func SetDebug(debug bool) {
	debugMode = debug
}

// IsDebug returns whether debug mode is enabled
func IsDebug() bool {
	return debugMode
}

// SetDetailedLogging enables or disables detailed logging with file and line information
func SetDetailedLogging(detailed bool) {
	// This function is kept for compatibility but no longer used
	// since we now handle detailed logging differently
}

// Global logging functions
func Info(format string, v ...interface{}) {
	GetLogger().Info(format, v...)
}

func Debug(format string, v ...interface{}) {
	GetLogger().Debug(format, v...)
}

func Warning(format string, v ...interface{}) {
	GetLogger().Warning(format, v...)
}

func Error(format string, v ...interface{}) error {
	return GetLogger().Error(format, v...)
}

// Fatal logs an error message and exits the program
func Fatal(format string, v ...interface{}) {
	GetLogger().Error(format, v...)
	os.Exit(1)
}
