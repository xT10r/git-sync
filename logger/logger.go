// Copyright 2024 Aleksey Dobshikov
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
	"log"
	"os"
)

type Logger struct {
	logger *log.Logger
	level  Level
}

type Level int64

const (
	Info Level = iota
	Warning
	Debug
	Error
)

const (
	InfoPrefix    = "[INFO] "
	DebugPrefix   = "[DEBUG] "
	WarningPrefix = "[WARN] "
	ErrorPrefix   = "[ERROR] "
)

var globalLogger *Logger

var (
	defaultFlags = log.LstdFlags | log.Lshortfile | log.Ltime
)

func InitLogeer(level Level, prefix string) {
	globalLogger = NewLogger(level, prefix)
}

func NewLogger(level Level, specialPrefix string) *Logger {

	var prefix string
	if specialPrefix == "" {
		switch level {
		case Info:
			prefix = InfoPrefix
		case Debug:
			prefix = DebugPrefix
		case Warning:
			prefix = WarningPrefix
		case Error:
			prefix = ErrorPrefix
		}
	}

	return &Logger{
		logger: log.New(os.Stdout, prefix, defaultFlags),
		level:  level,
	}
}

func GetLogger() *Logger {
	if globalLogger == nil {
		InitLogeer(Info, "")
	}
	return globalLogger
}

func (l *Logger) SetLevel(level Level) {
	l.level = level
}

func (l *Logger) Print(format string, v ...interface{}) {
	switch l.level {
	case Debug:
		if Debug >= l.level {
			l.logger.Printf(format, v...)
		}
	case Info:
		if Info >= l.level {
			l.logger.Printf(format, v...)
		}
	case Warning:
		if Warning >= l.level {
			l.logger.Printf(format, v...)
		}
	case Error:
		if Error >= l.level {
			l.logger.Printf(format, v...)
		}
	}
}

func (l *Logger) Info(message string) {
	if l.level <= Info {
		l.logger.Printf("%s%s\n", InfoPrefix, message)
	}
}

func (l *Logger) Warning(message string) {
	if l.level <= Warning {
		l.logger.Printf("%s%s\n", WarningPrefix, message)
	}
}

func (l *Logger) Error(message string) {
	if l.level <= Error {
		l.logger.Printf("%s%s\n", ErrorPrefix, message)
	}
}
